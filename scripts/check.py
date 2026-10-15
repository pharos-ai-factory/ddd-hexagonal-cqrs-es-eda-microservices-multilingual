"""Run the same static checks after edits, before hand-off and in CI."""

import argparse
import os
import subprocess
import sys

from source_files import ROOT, SERVICES, affected_services, changed

GO_TOOLS = (
    "honnef.co/go/tools/cmd/staticcheck@v0.8.2",
    "github.com/kisielk/errcheck@v1.30.0",
)


def commands(services, full=False):
    """Yield complete project checks so edits are also checked against their callers."""
    if full:
        yield [sys.executable, "scripts/format.py", "--check", "--all"], ROOT, None
        yield ["pnpm", "docs:check"], ROOT, None
    for service in services:
        folder = ROOT / "services" / service
        if not full:
            yield [sys.executable, "scripts/format.py", "--check", "--service", service], ROOT, None
        if service in ("storefront", "api"):
            env = dict(os.environ, CAFE_GO_PROJECT="services/" + service)
            yield [sys.executable, "scripts/go.py", "vet", "./..."], ROOT, env
            for tool in GO_TOOLS:
                options = ["-ignoregenerated"] if "errcheck" in tool else []
                yield [sys.executable, "scripts/go.py", "run", tool, *options, "./..."], ROOT, env
        elif service == "operations":
            yield ["uv", "run", "--frozen", "mypy"], folder, None
            yield ["uv", "run", "--frozen", "ruff", "check", "src", "tests"], folder, None
        else:
            yield ["pnpm", "--filter", "@cafe/" + service, "type-check"], ROOT, None
            yield (
                [
                    "pnpm",
                    "exec",
                    "eslint",
                    "--max-warnings=0",
                    f"services/{service}/src",
                    *([f"services/{service}/tests"] if service == "engagement" else []),
                ],
                ROOT,
                None,
            )
    if full:
        yield (
            [
                "uv",
                "run",
                "--project",
                "services/operations",
                "--frozen",
                "ruff",
                "check",
                "scripts",
                "tests",
            ],
            ROOT,
            None,
        )
        for project in ("tests/acceptance", "tests/infrastructure", "scripts/browser"):
            yield ["pnpm", "exec", "tsc", "--project", project + "/tsconfig.json"], ROOT, None
        yield ["pnpm", "exec", "eslint", "--max-warnings=0", "tests", "scripts/browser"], ROOT, None
    yield [sys.executable, "scripts/check_architecture.py"], ROOT, None
    yield ["node", "scripts/check_command_boundaries.mjs"], ROOT, None
    yield (
        [sys.executable, "scripts/go.py", "test", "-count=1", "./tests/architecture"],
        ROOT,
        dict(os.environ, CAFE_GO_PROJECT="services/storefront"),
    )


def check(services, full=False):
    failures = []
    for arguments, cwd, env in commands(services, full):
        print("Checking: " + " ".join(arguments), flush=True)
        result = subprocess.run(arguments, cwd=cwd, env=env, check=False)
        if result.returncode:
            failures.append(" ".join(arguments))
    if failures:
        print("Failed checks:\n" + "\n".join(failures), file=sys.stderr)
    else:
        print("Static checks passed", flush=True)
    return bool(failures)


def main():
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("services", nargs="*")
    parser.add_argument("--changed", action="store_true")
    parser.add_argument("--watch", action="store_true")
    args = parser.parse_args()
    if any(service not in SERVICES for service in args.services):
        parser.error("Unknown service")
    if args.changed and args.services:
        parser.error("Choose services or --changed")
    if args.watch:
        from check_watch import watch

        watch(args.services)
    else:
        names = changed() if args.changed else None
        selected = affected_services(names) if names is not None else list(args.services or SERVICES)
        full = not args.services and (
            names is None or any(not name.startswith("services/") for name in names)
        )
        if names is not None and not names:
            print("No changed files")
            return
        raise SystemExit(check(selected, full))


if __name__ == "__main__":
    try:
        main()
    except KeyboardInterrupt:
        raise SystemExit(130) from None
