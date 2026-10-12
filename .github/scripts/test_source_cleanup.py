"""Exercise the owned source-bootstrap cleanup boundary without building Tools."""
import os
from pathlib import Path
import subprocess
import tempfile
import unittest

SCRIPT = Path(__file__).with_name("cleanup-source-bootstrap.sh")


class CleanupTests(unittest.TestCase):
    def run_cleanup(self, runner, target, success):
        environment = dict(os.environ, RUNNER_TEMP=str(runner),
                           GITHUB_ENV=str(runner / "environment"))
        environment.pop("GOLIB_SOURCE_INSTALL_ROOT", None)
        if target is not None:
            environment["GOLIB_SOURCE_INSTALL_ROOT"] = str(target)
        result = subprocess.run(["bash", str(SCRIPT)], env=environment,
                                capture_output=True, timeout=5, check=False)
        self.assertEqual(result.returncode == 0, success, result.stderr.decode())

    def test_regular_owned_workspace_is_removed_and_tempdir_restored(self):
        with tempfile.TemporaryDirectory(prefix="assurance-cleanup-test-") as task:
            root = Path(task)
            target = root / "golib-source-install.owned"
            (target / "bin").mkdir(parents=True)
            binary = target / "bin/golib"
            binary.write_bytes(b"owned binary")
            binary.chmod(0o400)
            self.run_cleanup(root, target, True)
            self.assertFalse(target.exists())
            self.assertEqual((root / "environment").read_text(), f"TMPDIR={root}\n")
            self.run_cleanup(root, target, True)

    def test_absent_setup_is_a_noop(self):
        with tempfile.TemporaryDirectory(prefix="assurance-cleanup-test-") as task:
            self.run_cleanup(Path(task), None, True)

    def test_unowned_symlink_nonregular_and_nested_paths_refuse(self):
        for shape in ("outside", "nested", "root-symlink", "child-symlink", "file", "fifo", "hard-link"):
            with self.subTest(shape=shape), tempfile.TemporaryDirectory(
                prefix="assurance-cleanup-test-"
            ) as task:
                root = Path(task)
                outside = root / "outside"
                outside.mkdir()
                sentinel = outside / "sentinel"
                sentinel.write_bytes(b"preserve")
                target = root / "golib-source-install.owned"
                if shape == "outside":
                    target = outside
                elif shape == "nested":
                    target = target / "nested"
                    target.mkdir(parents=True)
                elif shape == "root-symlink":
                    target.symlink_to(outside, target_is_directory=True)
                elif shape == "file":
                    target.write_bytes(b"preserve")
                else:
                    target.mkdir()
                    child = target / "child"
                    if shape == "child-symlink":
                        child.symlink_to(sentinel)
                    elif shape == "fifo":
                        os.mkfifo(child)
                    else:
                        os.link(sentinel, child)
                self.run_cleanup(root, target, False)
                self.assertEqual(sentinel.read_bytes(), b"preserve")
                self.assertTrue(target.exists())


if __name__ == "__main__":
    unittest.main()
