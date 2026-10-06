#!/usr/bin/env python3
"""Use the local Go toolchain or the pinned development container."""
import os
import json
from pathlib import Path
import shutil
import subprocess
import sys

ROOT = Path(__file__).resolve().parents[1]
PROJECT = ROOT / os.environ.get("CAFE_GO_PROJECT", "services/storefront")
local = ROOT / ".tools/go/bin/go"
go = str(local) if local.exists() else shutil.which("go")
if os.environ.get("CAFE_GO_CONTAINER") == "1":
    go = None
arguments = sys.argv[1:]
formatting = arguments[:1] == ["gofmt"]
if formatting:
    arguments = arguments[1:]
if go:
    tool = go
    if formatting:
        goroot = subprocess.check_output([go, "env", "GOROOT"], text=True).strip()
        tool = str(Path(goroot)/"bin/gofmt")
    args = [tool, *arguments]
else:
    cache = ROOT / ".local/go-cache"
    cache.mkdir(parents=True, exist_ok=True)
    args = ["docker", "run", "--rm"]
    security = json.loads(subprocess.check_output(
        ["docker", "info", "--format", "{{json .SecurityOptions}}"], text=True))
    if "name=rootless" not in security:
        args += ["--user", f"{os.getuid()}:{os.getgid()}"]
    args += ["-e", "GOPATH=/cache", "-e", "GOCACHE=/cache/build", "-v", f"{ROOT}:/src",
            "-v", f"{cache}:/cache", "-w", "/src" if formatting else "/src/"+str(PROJECT.relative_to(ROOT))]
    if os.environ.get("DATABASE_ADMIN_URL") or os.environ.get("BROKER_URL") or os.environ.get("CAFE_HTTP_API_URL"):
        # Integration services publish random loopback ports on the host.
        args += ["--network", "host"]
    for name in os.environ:
        if name.endswith(("_DATABASE_URL", "_BROKER_URL")) or name in {
            "DATABASE_ADMIN_URL", "BROKER_ADMIN_URL", "BROKER_URL", "APP_ENV", "UPDATE_FIXTURES",
            "CAFE_DISPOSABLE_PROJECT", "CAFE_HTTP_API_URL", "CAFE_HTTP_API_KEY",
        }:
            args += ["-e", name]
    if os.environ.get("GOBIN"):
        relative = Path(os.environ["GOBIN"]).relative_to(ROOT)
        args += ["-e", "GOBIN=/src/"+str(relative)]
    if os.environ.get("BDD_REPORT_DIR"):
        relative = Path(os.environ["BDD_REPORT_DIR"]).resolve().relative_to(ROOT)
        args += ["-e", "BDD_REPORT_DIR=/src/"+str(relative)]
    args += ["golang:1.27.1", "gofmt" if formatting else "go", *arguments]
raise SystemExit(subprocess.call(args, cwd=ROOT if formatting else PROJECT))
