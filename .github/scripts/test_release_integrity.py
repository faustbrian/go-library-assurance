"""Exercise release recipe refusal/cleanup with process doubles, not release proof."""
import hashlib
import io
import json
import os
from pathlib import Path
import shutil
import subprocess
import tarfile
import tempfile
import unittest


SCRIPT = Path(__file__).with_name("qualify-release-integrity.sh")
SOURCE = "09b43f49385650c44dd4151a12b0877ca6b48a88"
PLATFORMS = ("linux_amd64", "linux_arm64", "darwin_amd64", "darwin_arm64")


DOUBLE = r'''#!/usr/bin/env python3
import json, os, pathlib, shutil, sys
args=sys.argv[1:]; tool=pathlib.Path(sys.argv[0]).name
failure=os.environ.get('FAIL_AT','')
expected={'GOENV':'off','GOWORK':'off','GOTOOLCHAIN':'local','GOPROXY':'https://proxy.golang.org',
 'GOSUMDB':'sum.golang.org','GOPRIVATE':'','GONOPROXY':'','GONOSUMDB':'','GOFLAGS':'-mod=readonly -p=1',
 'CGO_ENABLED':'0','GOAMD64':'v1','GOARM64':'v8.0','GOEXPERIMENT':''}
if any(os.environ.get(key)!=value for key,value in expected.items()): sys.exit('unsafe invocation environment')
with open(os.environ['CALL_LOG'],'a') as out:
 out.write(json.dumps({'tool':tool,'args':args,'cwd':os.getcwd(),
  'cache':os.environ.get('GOCACHE'),'mod':os.environ.get('GOMODCACHE'),'tmp':os.environ.get('GOTMPDIR')})+'\n')
fixtures=pathlib.Path(os.environ['FIXTURES'])
source='09b43f49385650c44dd4151a12b0877ca6b48a88'
if tool=='gh':
 if args[0]=='--version': print('gh version fixture'); sys.exit()
 if args[:2]==['release','view']:
  print(json.dumps({'tagName':'v2.5.1','isDraft':False,'isPrerelease':False})); sys.exit()
 if args[:2]==['release','download']:
  name=args[args.index('--pattern')+1]; dest=pathlib.Path(args[args.index('--dir')+1])/name
  shutil.copyfile(fixtures/name,dest)
  if failure=='checksums' and name=='checksums.txt': dest.write_text('invalid checksum\n')
  sys.exit()
 if args[:2]==['attestation','verify']:
  if failure=='attestation': sys.exit(23)
  print('[{"fixture":"not actual attestation proof"}]'); sys.exit()
elif tool=='git':
 if 'init' in args:
  root=pathlib.Path(args[-1]); root.mkdir()
  (root/'go.mod').write_text('module fixture.tools\ngo 1.27.0\n'); (root/'go.sum').write_text('fixture sum\n')
 if 'rev-parse' in args: print(source)
 sys.exit()
elif tool=='go':
 if args[0]=='env':
  data={k:os.environ.get(k,'') for k in args[2:]}
  data.update(GOVERSION='go1.27.2',GOHOSTOS='linux',GOHOSTARCH='amd64',GOOS='linux',GOARCH='amd64')
  if failure=='compiler': data['GOVERSION']='go1.27.0'
  print(json.dumps(data)); sys.exit()
 if args[:3]==['mod','download','-json']:
  module,version=args[3].split('@'); library='jsonapi' if 'jsonapi' in module else 'localized'
  if 'CycloneDX' in module:
   print(json.dumps({'Path':module,'Version':version,'Sum':'h1:fixture-generator'})); sys.exit()
  data={
   'jsonapi':('6f44f8b7cd089d36bc36e67e6aec38faa5ad1aa0','h1:6kfqQ3swtoHvKcaHVqftn6YnakRxkxWFl7K7BvMKSjc=','h1:MV6Dr+s3ii1yQK/FIZ9Z861Z+7ToQhiuxvZssu1nSEY='),
   'localized':('534315549084994f3d2a2e088400ddc2bf3a6b68','h1:gTAPWcue3YTBZtHKNLAe4hwBSr8U4TecrXOFhz1nPa0=','h1:tnQlj1cAbCt0tc0JpAfxkbd/G6RHVUBPGjUoxx14m2Q=')}
  revision,total,modsum=data[library]; root=pathlib.Path(os.environ['GOMODCACHE'])/library; root.mkdir()
  (root/'go.mod').write_text('module '+module+'\ngo 1.27.0\n'); (root/'go.sum').write_text('fixture sum\n')
  if failure=='origin': revision='0'*40
  if failure=='sum': total='h1:wrong'
  print(json.dumps({'Path':module,'Version':version,'Dir':str(root),'Sum':total,'GoModSum':modsum,
   'Origin':{'VCS':'git','URL':'https://'+module.rsplit('/v',1)[0],'Hash':revision,'Ref':'refs/tags/'+version}})); sys.exit()
 if args[0]=='list':
  graph={'Path':'fixture.module','Main':True}
  if failure=='replacement': graph['Replace']={'Dir':'/untrusted'}
  print(json.dumps(graph)); sys.exit()
 if args[0]=='run' and 'CycloneDX' in args[1]:
  if failure=='generator': sys.exit(23)
  output=pathlib.Path(args[args.index('-output')+1]); output.write_text(json.dumps({
   'bomFormat':'CycloneDX','specVersion':'1.6','metadata':{'component':{'version':'(devel)'}}}))
  if failure=='source-mutation': pathlib.Path('go.sum').write_text('changed\n')
  sys.exit()
 if args[0]=='build':
  if failure=='build': sys.exit(23)
  pathlib.Path(args[args.index('-o')+1]).write_bytes(('binary-'+os.environ['GOOS']+'_'+os.environ['GOARCH']).encode()); sys.exit()
 if args[:4]==['run','./cmd/golib','release','package']:
  output=pathlib.Path(args[args.index('--output')+1]); shutil.copyfile(fixtures/output.name,output)
  if failure=='comparison': output.write_bytes(b'not matching archive')
  sys.exit()
 if args[:2]==['mod','verify']: print('all modules verified'); sys.exit()
 if args[:2]==['version','-m']: print('fixture build identity'); sys.exit()
 if args[:4]==['run','./cmd/golib','release','archive-check']: sys.exit()
sys.exit('unexpected process-double call: '+tool+' '+repr(args))
'''


class ReleaseIntegrityTests(unittest.TestCase):
    def exercise(self, failure=""):
        with tempfile.TemporaryDirectory(prefix="assurance-release-test-") as owned:
            root = Path(owned)
            fixtures = root / "fixtures"
            fixtures.mkdir()
            entries = []
            for platform in PLATFORMS:
                path = fixtures / ("golib_2.5.1_" + platform + ".tar.gz")
                with tarfile.open(path, "w:gz") as archive:
                    body = ("binary-" + platform).encode()
                    member = tarfile.TarInfo("golib")
                    member.size = len(body)
                    archive.addfile(member, io.BytesIO(body))
                entries.append({"name": path.name, "sha256": hashlib.sha256(path.read_bytes()).hexdigest()})
            manifest = fixtures / "release-manifest.json"
            manifest.write_text(json.dumps({"schema_version": 1, "version": "v2.5.1", "source_commit": SOURCE,
                                            "provenance": "github-artifact-attestation", "assets": entries}))
            checksums = [dict(name=manifest.name, sha256=hashlib.sha256(manifest.read_bytes()).hexdigest())] + entries
            (fixtures / "checksums.txt").write_text("".join(item["sha256"] + "  " + item["name"] + "\n" for item in checksums))
            binaries = root / "bin"
            binaries.mkdir()
            for name in ("go", "git", "gh"):
                binary = binaries / name
                binary.write_text(DOUBLE)
                binary.chmod(0o700)
            log = root / "calls.jsonl"
            environment = dict(os.environ, PATH=str(binaries) + os.pathsep + os.environ["PATH"],
                               RUNNER_TEMP=str(root), CALL_LOG=str(log), FIXTURES=str(fixtures), FAIL_AT=failure,
                               GOPROXY="off", GOFLAGS="-mod=mod")
            result = subprocess.run(["bash", str(SCRIPT)], env=environment, capture_output=True, timeout=15, check=False)
            calls = [json.loads(line) for line in log.read_text().splitlines()] if log.exists() else []
            evidence = root / "assurance-release-integrity-evidence"
            self.assertEqual(list(root.glob("assurance-release-integrity.*")), [], "owned cache task survived")
            for call in calls:
                for key in ("cache", "mod", "tmp"):
                    self.assertFalse(Path(call[key]).exists(), "owned cache survived")
            self.assertEqual(int((evidence / "exit-status.txt").read_text()), result.returncode)
            outputs = {str(path.relative_to(evidence)): path.read_text() for path in evidence.rglob("*") if path.is_file()}
            return result, calls, outputs

    def test_success_retains_retrospective_identity_without_rewriting_detected_root(self):
        result, calls, outputs = self.exercise()
        self.assertEqual(result.returncode, 0, result.stderr.decode())
        for library in ("jsonapi", "localized"):
            metadata = json.loads(outputs[library + "/metadata.json"])
            self.assertTrue(metadata["retrospective"])
            self.assertFalse(metadata["original_release_pipeline"])
            self.assertEqual(metadata["autodetected_bom_root_version"], "(devel)")
            self.assertEqual(json.loads(outputs[library + "/sbom.cdx.json"])["metadata"]["component"]["version"], "(devel)")
        for platform in PLATFORMS:
            self.assertTrue(json.loads(outputs[platform + ".comparison.json"])["binary_and_archive_byte_equal"])
        self.assertIn("evidence.sha256", outputs)
        self.assertEqual(len([call for call in calls if call["tool"] == "gh" and call["args"][:2] == ["attestation", "verify"]]), 6)

    def test_failure_guards_reject_and_cleanup_instead_of_manufacturing_proof(self):
        for boundary in ("compiler", "attestation", "checksums", "origin", "sum", "replacement",
                         "generator", "source-mutation", "build", "comparison"):
            with self.subTest(boundary=boundary):
                result, calls, outputs = self.exercise(boundary)
                self.assertNotEqual(result.returncode, 0, result.stderr.decode())
                self.assertNotIn("evidence.sha256", outputs)
                if boundary in ("compiler", "attestation", "checksums", "origin", "sum", "replacement", "generator", "source-mutation"):
                    self.assertFalse(any(call["tool"] == "go" and call["args"][0] == "build" for call in calls))


if __name__ == "__main__":
    unittest.main()
