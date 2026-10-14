#!/usr/bin/env python3
"""Deterministic gates for every service; real infrastructure has a separate lane."""

import os
from pathlib import Path
import subprocess
import sys

ROOT = Path(__file__).resolve().parents[1]
REPORTS = ROOT / ".local/bdd"
REPORTS.mkdir(parents=True, exist_ok=True)


def run(*args, cwd=ROOT, env=None):
    subprocess.run(args, cwd=cwd, env=env, check=True)


run(sys.executable, "scripts/check.py")
run(sys.executable, "scripts/check_contracts.py")
run("uv", "run", "--no-project", "--with", "grpcio-tools==1.84.0", "python", "scripts/compatibility.py")
run("node", "--test", "scripts/check_command_boundaries.test.mjs")
run("node", "--test", "scripts/test_quality_lint.mjs")
run("node", "scripts/check_specifications.mjs")
run("node", "--test", "scripts/check_specifications.test.mjs")
run(sys.executable, "-m", "unittest", "discover", "-s", "scripts", "-p", "test_*.py")
for service in ("storefront", "api"):
    env = dict(os.environ, CAFE_GO_PROJECT="services/" + service, BDD_REPORT_DIR=str(REPORTS))
    run(sys.executable, "scripts/go.py", "test", "-race", "-count=1", "./...", env=env)
run("uv", "sync", "--frozen", cwd=ROOT / "services/operations")
run(
    "uv",
    "run",
    "--frozen",
    "pytest",
    "-q",
    "--cucumberjson=" + str(REPORTS / "operations.json"),
    cwd=ROOT / "services/operations",
)
for service in ("engagement", "web"):
    run("pnpm", "--filter", "@cafe/" + service, "test")
run(
    "pnpm", "--filter", "@cafe/engagement", "test:bdd", "--format", "json:" + str(REPORTS / "engagement.json")
)
run(
    "node",
    "--import",
    "tsx",
    "node_modules/@cucumber/cucumber/bin/cucumber.js",
    "--config",
    "tests/acceptance/cucumber.mjs",
    "--dry-run",
    "--format",
    "json:" + str(REPORTS / "workflows-dry-run.json"),
)
run("node", "scripts/check_bdd_reports.mjs", "fast")
print(
    "Deterministic verification passed for all five applications. Real infrastructure: pnpm test:integration"
)
