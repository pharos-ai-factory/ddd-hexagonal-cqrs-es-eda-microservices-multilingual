"""Format handwritten source; use --all for the complete repository baseline."""

import argparse
import subprocess
import sys

from source_files import ROOT, SERVICES, changed, handwritten, inventory


def format_files(names, check=False):
    groups = {"go": [], "python": [], "prettier": []}
    for name in sorted(names):
        path = (ROOT / name).resolve()
        if not path.is_relative_to(ROOT):
            raise ValueError("Formatting paths must belong to this repository")
        relative = path.relative_to(ROOT).as_posix()
        if not handwritten(relative):
            continue
        group = "go" if path.suffix == ".go" else "python" if path.suffix in {".py", ".pyi"} else "prettier"
        groups[group].append(relative)
    failures = []
    for group, names in groups.items():
        if not names:
            continue
        if group == "go":
            command = [sys.executable, "scripts/go.py", "gofmt", "-l" if check else "-w", *names]
            result = subprocess.run(command, cwd=ROOT, capture_output=check, text=True, check=False)
            if check and result.stdout:
                print(result.stdout, end="")
                failures.append(group)
            if result.stderr:
                print(result.stderr, file=sys.stderr, end="")
        elif group == "python":
            result = subprocess.run(
                [
                    "uv",
                    "run",
                    "--project",
                    "services/operations",
                    "--frozen",
                    "ruff",
                    "format",
                    *(["--check"] if check else []),
                    *names,
                ],
                cwd=ROOT,
                check=False,
            )
        else:
            result = subprocess.run(
                ["pnpm", "exec", "prettier", "--check" if check else "--write", *names], cwd=ROOT, check=False
            )
        if result.returncode:
            failures.append(group)
    return bool(failures)


def main():
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("--check", action="store_true")
    parser.add_argument("--all", action="store_true")
    parser.add_argument("--service", choices=SERVICES)
    parser.add_argument("paths", nargs="*")
    args = parser.parse_args()
    if args.paths and (args.all or args.service):
        parser.error("Choose explicit paths or repository/service selection")
    names = args.paths or (inventory() if args.all or args.service else changed())
    if args.service:
        names = [name for name in names if name.startswith(f"services/{args.service}/")]
    raise SystemExit(format_files(names, args.check))


if __name__ == "__main__":
    main()
