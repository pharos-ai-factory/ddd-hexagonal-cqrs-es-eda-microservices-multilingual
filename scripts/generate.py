#!/usr/bin/env python3
"""Generate transport bindings for all runtimes from the shared wire schemas."""
import os
import json
from hashlib import sha256
from pathlib import Path
import shutil
import subprocess
ROOT = Path(__file__).resolve().parents[1]
plugin_dir = ROOT/".tools/bin"
plugin_dir.mkdir(parents=True, exist_ok=True)
env = dict(os.environ, GOBIN=str(plugin_dir))
subprocess.run(["python3", "scripts/go.py", "install", "google.golang.org/protobuf/cmd/protoc-gen-go@v1.36.12"], cwd=ROOT, env=env, check=True)
python_out = ROOT/"services/operations/src/operations/adaptors/generated"
python_out.mkdir(parents=True, exist_ok=True)
for contract, source in (("events", "cafe/v1/events.proto"), ("realtime", "cafe/realtime/v1/realtime.proto")):
    go_out = ROOT/f"services/storefront/contracts/{contract}/generated"
    go_out.mkdir(parents=True, exist_ok=True)
    subprocess.run(["uv", "run", "--no-project", "--with", "grpcio-tools==1.84.0", "python", "-m",
                    "grpc_tools.protoc", f"-Icontracts/{contract}/proto",
                    f"--plugin=protoc-gen-go={plugin_dir}/protoc-gen-go", f"--go_out={go_out}",
                    "--go_opt=paths=source_relative", f"--python_out={python_out}", f"--pyi_out={python_out}", source],
                   cwd=ROOT, check=True)
shutil.copyfile(ROOT/"contracts/events/catalogue.json", python_out/"catalogue.json")
shutil.copyfile(ROOT/"contracts/events/fixtures/reward-earned.v1.hex",
                ROOT/"services/storefront/contracts/events/fixtures/reward-earned.v1.hex")
shutil.copyfile(ROOT/"contracts/persistence/0002_realtime.sql",
               ROOT/"services/storefront/foundation/persistence/postgres/migrations/0002_realtime.sql")
checksums = {str(version): sha256((ROOT/f"contracts/persistence/{name}").read_bytes()).hexdigest()
             for version, name in ((1, "0001_initial.sql"), (2, "0002_realtime.sql"))}
for directory in (python_out, ROOT/"services/engagement/src/adaptors/generated"):
    (directory/"persistence.json").write_text(json.dumps(checksums, indent=2)+"\n")
subprocess.run(["node", "scripts/generate-typescript.mjs"], cwd=ROOT, check=True)
