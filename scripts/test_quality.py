"""Prove edit selection, watcher recovery and the static gate's negative cases."""

import json
import os
from pathlib import Path
import subprocess
import tempfile
import unittest
from unittest.mock import patch

from check import GO_TOOLS, check, commands
from check_watch import difference, watch
from source_files import ROOT, SERVICES, affected_services, changed, inventory


class QualityTests(unittest.TestCase):
    def test_go_analysers_reject_ignored_errors_and_invalid_calls(self):
        local = ROOT / ".local"
        local.mkdir(exist_ok=True)
        with tempfile.TemporaryDirectory(prefix="quality-go-", dir=local) as directory:
            root = Path(directory)
            (root / "go.mod").write_text("module quality-fixture\n\ngo 1.27.1\n")
            (root / "main.go").write_text(
                'package main\nimport ("os"; "fmt"; "strings")\n'
                'func main() { os.Chdir("."); fmt.Printf("%d", "text"); strings.Trim("a", "aa") }\n'
            )
            env = dict(os.environ, CAFE_GO_PROJECT=str(root.relative_to(ROOT)))
            lanes = [
                (["vet", "./..."], "fmt.Printf"),
                (["run", GO_TOOLS[0], "./..."], "SA1024"),
                (["run", GO_TOOLS[1], "./..."], "os.Chdir"),
            ]
            for arguments, expected in lanes:
                result = subprocess.run(
                    ["python3", "scripts/go.py", *arguments],
                    cwd=ROOT,
                    env=env,
                    capture_output=True,
                    text=True,
                )
                self.assertNotEqual(result.returncode, 0)
                self.assertIn(expected, result.stdout + result.stderr)

    def test_inventory_includes_untracked_files_and_ignores_build_output(self):
        with tempfile.TemporaryDirectory() as directory:
            root = Path(directory)
            subprocess.run(["git", "init", "-q", directory], check=True)
            (root / ".gitignore").write_text("dist/\n")
            (root / "new.py").write_text("answer = 42\n")
            (root / "dist").mkdir()
            (root / "dist/output.js").write_text("generated")
            self.assertIn("new.py", inventory(root))
            self.assertNotIn("dist/output.js", inventory(root))

    def test_changed_selection_combines_tracked_changes_and_untracked_files(self):
        with patch("source_files.git_paths", side_effect=[{"deleted.ts"}, {"new.py"}]):
            self.assertEqual(changed(), {"deleted.ts", "new.py"})

    def test_shared_inputs_recheck_consumers_and_deletions_keep_their_owner(self):
        self.assertEqual(affected_services({"services/web/src/deleted.ts"}), ["web"])
        self.assertEqual(affected_services({"tsconfig.strict.json"}), list(SERVICES))
        self.assertEqual(affected_services({"contracts/menu/messaging/new.proto"}), list(SERVICES))
        self.assertEqual(difference({"removed.py": (1, 1)}, {"new.py": (1, 1)}), {"removed.py", "new.py"})

    def test_failed_watch_run_retries_after_an_edit_during_the_check(self):
        original = {"services/web/src/old.ts": (1, 1)}
        edited = {"services/web/src/new.ts": (2, 1)}
        with (
            patch("check_watch.snapshot", side_effect=[original, edited, edited, KeyboardInterrupt]),
            patch("check_watch.time.sleep"),
            patch("check.check", side_effect=[True, False]) as run,
        ):
            with self.assertRaises(KeyboardInterrupt):
                watch(["web"])
        self.assertEqual(run.call_count, 2)
        self.assertEqual(run.call_args.args, (["web"],))

    def test_failure_is_reported_after_independent_checks_run(self):
        steps = [(["first"], ROOT, None), (["second"], ROOT, None)]
        with patch("check.commands", return_value=steps), patch("check.subprocess.run") as run:
            run.side_effect = [subprocess.CompletedProcess([], 1), subprocess.CompletedProcess([], 0)]
            self.assertTrue(check(["web"]))
            self.assertEqual(run.call_count, 2)

    def test_service_selection_checks_the_project_and_global_boundaries(self):
        selected = [args for args, _, _ in commands(["engagement"])]
        self.assertIn(["pnpm", "--filter", "@cafe/engagement", "type-check"], selected)
        self.assertTrue(any("scripts/check_architecture.py" in args for args in selected))
        self.assertFalse(any("@cafe/web" in args for args in selected))

    def test_typescript_flags_reject_unsafe_indexing_optional_values_and_overrides(self):
        with tempfile.TemporaryDirectory() as directory:
            root = Path(directory)
            config = {
                "extends": str(ROOT / "tsconfig.strict.json"),
                "compilerOptions": {
                    "noEmit": True,
                    "types": [],
                    "target": "ES2023",
                    "skipLibCheck": True,
                },
                "files": ["fixture.test.ts"],
            }
            (root / "tsconfig.json").write_text(json.dumps(config))
            source = root / "fixture.test.ts"
            source.write_text("export {};\nconst numbers: number[] = [];\nconst value = numbers[0];\n")
            args = ["pnpm", "exec", "tsc", "--project", str(root / "tsconfig.json")]
            valid = subprocess.run(args, cwd=ROOT, capture_output=True, text=True)
            self.assertEqual(valid.returncode, 0, valid.stdout + valid.stderr)
            source.write_text(
                source.read_text() + "const required: number = numbers[0];\n"
                "const optional: {value?: number} = {value: undefined};\n"
                "class Base { run() {} }\nclass Child extends Base { run() {} }\n"
            )
            invalid = subprocess.run(args, cwd=ROOT, capture_output=True, text=True)
            self.assertNotEqual(invalid.returncode, 0)
            for code in ("TS2322", "TS2375", "TS4114"):
                self.assertIn(code, invalid.stdout)

    def test_engagement_checks_adjacent_tests_and_the_pool_probe(self):
        result = subprocess.run(
            ["pnpm", "exec", "tsc", "-p", "services/engagement/tsconfig.check.json", "--listFilesOnly"],
            cwd=ROOT,
            capture_output=True,
            text=True,
            check=True,
        )
        for name in ("src/apps/runtime.test.ts", "tests/pool-failure.mts", "tests/bdd/world.ts"):
            self.assertIn(name, result.stdout)
