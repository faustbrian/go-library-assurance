"""Monitor orchestration, using a process double only for the Go boundary."""
import json
import os
from pathlib import Path
import shutil
import subprocess
import tempfile
import unittest


SCRIPT = Path(__file__).with_name("monitor-advisories.sh")


class AdvisoryTests(unittest.TestCase):
    def run_monitor(self, fail_at=""):
        with tempfile.TemporaryDirectory(prefix="assurance-advisory-test-") as owned:
            root = Path(owned)
            scripts = root / ".github/scripts"
            scripts.mkdir(parents=True)
            if SCRIPT.exists():
                shutil.copyfile(SCRIPT, scripts / SCRIPT.name)
            else:
                (scripts / SCRIPT.name).write_text("")
            scenario = root / "scenarios/compatibility-consumer"
            scenario.mkdir(parents=True)
            (scenario / "go.mod").write_text("module example.invalid/scenario\ngo 1.27.0\n")
            binaries = root / "bin"
            binaries.mkdir()
            log = root / "calls.jsonl"
            go = binaries / "go"
            go.write_text("""#!/usr/bin/env python3
import json, os, sys
keys = ('GOCACHE', 'GOMODCACHE', 'GOTMPDIR', 'GOWORK', 'GOENV', 'GOTOOLCHAIN',
        'GOPROXY', 'GOSUMDB', 'GOPRIVATE', 'GONOPROXY', 'GONOSUMDB', 'GOFLAGS', 'GOVULNDB')
with open(os.environ['CALL_LOG'], 'a') as log:
    log.write(json.dumps({'args': sys.argv[1:], 'cwd': os.getcwd(),
                          'env': {key: os.environ.get(key) for key in keys}}) + '\\n')
if os.environ.get('FAIL_AT') and os.environ['FAIL_AT'] in ' '.join(sys.argv[1:]):
    sys.exit(23)
""")
            go.chmod(0o700)
            environment = dict(os.environ, PATH=str(binaries) + os.pathsep + os.environ["PATH"],
                               RUNNER_TEMP=str(root), CALL_LOG=str(log), FAIL_AT=fail_at,
                               GOVULNDB="https://untrusted.invalid", GOPROXY="off", GOFLAGS="-mod=mod")
            result = subprocess.run(["bash", str(scripts / SCRIPT.name)], env=environment,
                                    capture_output=True, timeout=5, check=False)
            calls = [json.loads(line) for line in log.read_text().splitlines()] if log.exists() else []
            for call in calls:
                for cache in ("GOCACHE", "GOMODCACHE", "GOTMPDIR"):
                    self.assertFalse(Path(call["env"][cache]).exists(), "owned cache survived")
            self.assertEqual(list(root.glob("assurance-advisories.*")), [], "owned task survived")
            return result, calls, str(scenario.resolve())

    def test_success_checks_integrity_then_mutable_advisories_without_compositions(self):
        result, calls, scenario = self.run_monitor()
        self.assertEqual(result.returncode, 0, result.stderr.decode())
        self.assertEqual([call["args"] for call in calls], [
            ["mod", "download"], ["mod", "verify"],
            ["run", "golang.org/x/vuln/cmd/govulncheck@v1.6.0", "-mode=source", "-scan=module", "-db=https://vuln.go.dev"],
        ])
        for call in calls:
            self.assertEqual(call["cwd"], scenario)
            self.assertEqual(call["env"]["GOWORK"], "off")
            self.assertEqual(call["env"]["GOTOOLCHAIN"], "local")
            self.assertEqual(call["env"]["GOPROXY"], "https://proxy.golang.org")
            self.assertEqual(call["env"]["GOSUMDB"], "sum.golang.org")
            self.assertEqual(call["env"]["GOVULNDB"], "https://vuln.go.dev")
            self.assertEqual(call["env"]["GOFLAGS"], "-mod=readonly -p=1")
            for key in ("GOPRIVATE", "GONOPROXY", "GONOSUMDB"):
                self.assertEqual(call["env"][key], "")

    def test_failure_and_findings_fail_the_job_and_clean_owned_resources(self):
        for boundary, count in (("download", 1), ("verify", 2), ("govulncheck", 3)):
            with self.subTest(boundary=boundary):
                result, calls, _ = self.run_monitor(boundary)
                self.assertEqual(result.returncode, 23)
                self.assertEqual(len(calls), count)


if __name__ == "__main__":
    unittest.main()
