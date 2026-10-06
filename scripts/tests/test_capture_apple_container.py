import importlib.util
import json
from pathlib import Path
import subprocess
import tempfile
import unittest
from unittest.mock import patch

spec = importlib.util.spec_from_file_location(
    "capture", Path(__file__).parents[1] / "capture-apple-container.py")
capture = importlib.util.module_from_spec(spec)
spec.loader.exec_module(capture)


class CaptureTests(unittest.TestCase):
    def test_read_only_commands_and_private_evidence(self):
        with tempfile.TemporaryDirectory() as root:
            output = Path(root) / "evidence"
            with patch.object(capture.subprocess, "run", return_value=
                              subprocess.CompletedProcess([], 0, b"[]", b"")) as run:
                self.assertTrue(capture.capture("/bin/container", output))
            for call in run.call_args_list:
                args = call.args[0][1:]
                self.assertTrue("--help" in args or "--version" in args or
                                args[:2] in (["system", "status"], ["network", "list"],
                                             ["volume", "list"]) or args[0] == "list")
                self.assertEqual(call.kwargs["timeout"], 30)
            self.assertEqual(output.stat().st_mode & 0o777, 0o700)
            for path in output.iterdir():
                self.assertEqual(path.stat().st_mode & 0o777, 0o600)
            metadata = json.loads((output / "capture.json").read_text())
            self.assertIn("unqualified", metadata["qualification"])
            with self.assertRaises(FileExistsError):
                capture.capture("/bin/container", output)

    def test_failures_and_timeout_are_not_reported_as_success(self):
        with tempfile.TemporaryDirectory() as root:
            with patch.object(capture.subprocess, "run", side_effect=
                              subprocess.TimeoutExpired("container", 30, output=b"partial")):
                self.assertFalse(capture.capture("/bin/container", Path(root) / "timeout"))
            metadata = json.loads((Path(root) / "timeout/capture.json").read_text())
            self.assertTrue(all(row["timed_out"] for row in metadata["commands"]))
            with patch.object(capture.subprocess, "run", return_value=
                              subprocess.CompletedProcess([], 1, b"", b"not running")):
                self.assertFalse(capture.capture("/bin/container", Path(root) / "failed"))


if __name__ == "__main__":
    unittest.main()
