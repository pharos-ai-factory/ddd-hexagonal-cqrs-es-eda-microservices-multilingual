#!/usr/bin/env python3
"""Run fast executable specifications in their owning language, with local reports."""
import os
from pathlib import Path
import subprocess
import sys

ROOT = Path(__file__).resolve().parents[1]
REPORTS = ROOT / ".local/bdd"


def run(*args, cwd=ROOT, env=None):
    subprocess.run(args, cwd=cwd, env=env, check=True)


def main():
    REPORTS.mkdir(parents=True, exist_ok=True)
    run("node", "scripts/check_specifications.mjs")
    run(sys.executable, "scripts/go.py", "test", "-race", "-count=1",
        "./contexts/menu/application", "./contexts/ordering/application",
        env=dict(os.environ, BDD_REPORT_DIR=str(REPORTS)))
    run("uv", "run", "--frozen", "pytest", "-q", "tests/bdd", "--cucumberjson="+str(REPORTS/"operations.json"),
        cwd=ROOT/"services/operations")
    run("pnpm", "--filter", "@cafe/engagement", "test:bdd", "--format", "json:"+str(REPORTS/"engagement.json"))
    run("node", "scripts/check_bdd_reports.mjs", "fast")
    print("Fast Gherkin scenarios passed. PostgreSQL and cross-service scenarios: pnpm test:integration")


if __name__ == "__main__":
    main()
