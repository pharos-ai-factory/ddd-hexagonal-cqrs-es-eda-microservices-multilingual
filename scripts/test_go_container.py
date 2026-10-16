"""The UID-mapped Go container puts analyser caches inside its writable mount."""

import json
import os
from pathlib import Path
import runpy
import unittest
from unittest.mock import patch

ROOT = Path(__file__).resolve().parents[1]


class ContainerCaches(unittest.TestCase):
    def test_staticcheck_cache_uses_the_mounted_cache_for_both_docker_modes(self):
        for security in ([], ["name=rootless"]):
            with (
                self.subTest(security=security),
                patch.dict(os.environ, {"CAFE_GO_CONTAINER": "1"}),
                patch("sys.argv", ["go.py", "run", "honnef.co/go/tools/cmd/staticcheck@v0.8.2", "./..."]),
                patch("subprocess.check_output", return_value=json.dumps(security)),
                patch("subprocess.call", return_value=0) as execute,
            ):
                with self.assertRaises(SystemExit) as result:
                    runpy.run_path(str(ROOT / "scripts/go.py"), run_name="__main__")
                self.assertEqual(0, result.exception.code)
                args = execute.call_args.args[0]
                self.assertIn("STATICCHECK_CACHE=/cache/staticcheck", args)
                self.assertIn(f"{ROOT / '.local/go-cache'}:/cache", args)
                self.assertEqual("name=rootless" not in security, "--user" in args)
                if "--user" in args:
                    self.assertEqual(f"{os.getuid()}:{os.getgid()}", args[args.index("--user") + 1])
