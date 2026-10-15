#!/usr/bin/env python3
"""Generate declared transport bindings from context-owned wire schemas."""

from http_contracts import generate as generate_http_contracts
from broker_permissions import generate as generate_broker_permissions
from contract_sources import PRIVATE, catalogue, command_subscriptions, sources as contract_sources
from generate_requests import generate_requests
from generate_selected import generate_selected, qualify_python_imports
from generate_events import generate as generate_events
from protobuf_tools import ensure_go_plugin
from contract_selection import declarations
import tempfile
import json
from hashlib import sha256
from migration_catalogue import generate as generate_migration_metadata
from pathlib import Path
import shutil
import subprocess

ROOT = Path(__file__).resolve().parents[1]
declarations(ROOT)
plugin = ensure_go_plugin(ROOT)
standard_protos = subprocess.check_output(
    [
        "uv",
        "run",
        "--no-project",
        "--with",
        "grpcio-tools==1.84.0",
        "python",
        "-c",
        "import grpc_tools; from pathlib import Path; print(Path(grpc_tools.__file__).parent/'_proto')",
    ],
    text=True,
).strip()
python_out = ROOT / "services/operations/src/operations/adaptors/generated"
python_out.mkdir(parents=True, exist_ok=True)
generate_selected(plugin)
generate_events()
for contract in ("menu_private", "ordering_private"):
    owner = contract.removesuffix("_private")
    definitions = contract_sources(contract)
    commands = ROOT / PRIVATE[owner] / "internal_commands.proto"
    if commands.exists():
        definitions[f"{owner}_commands.proto"] = commands
        definitions["cafe/requests/v1/validation.proto"] = (
            ROOT / "contracts/shared/messaging/v1/validation.proto"
        )
    if len({Path(name).stem for name in definitions}) != len(definitions):
        raise ValueError(f"{contract}: Protobuf filenames must be unique for the shared Go package")
    go_relative = PRIVATE[owner] + "/generated"
    go_out = ROOT / go_relative
    go_out.mkdir(parents=True, exist_ok=True)
    for path in go_out.rglob("*.pb.go"):
        path.unlink()
    # Logical import names remain stable while developer-facing source folders
    # follow context ownership. Staging contains specifications, never runtime code.
    with tempfile.TemporaryDirectory() as directory:
        proto_root = Path(directory)
        for name, source in definitions.items():
            target = proto_root / name
            target.parent.mkdir(parents=True, exist_ok=True)
            shutil.copyfile(source, target)
        module = (
            "github.com/pharos-ai-factory/ddd-hexagonal-cqrs-es-eda-microservices-multilingual/" + go_relative
        )
        arguments = [
            "uv",
            "run",
            "--no-project",
            "--with",
            "grpcio-tools==1.84.0",
            "python",
            "-m",
            "grpc_tools.protoc",
            f"-I{proto_root}",
            f"-I{standard_protos}",
            f"--plugin=protoc-gen-go={plugin}",
            f"--go_out={go_out}",
            "--go_opt=module=" + module,
        ]
        subprocess.run(
            [*arguments, *(name for name in definitions if not name.startswith("cafe/requests/"))],
            cwd=ROOT,
            check=True,
        )
for service in ("storefront", "api", "operations"):
    generate_requests(service, ROOT, plugin, standard_protos)
subprocess.run(
    [
        "uv",
        "run",
        "--no-project",
        "--with",
        "grpcio-tools==1.84.0",
        "python",
        "scripts/generate_request_names.py",
    ],
    cwd=ROOT,
    check=True,
)
# Internal commands are implementation resources owned by their receiving context.
for owner in ("preparation", "collection"):
    source = ROOT / f"services/operations/src/operations/contexts/{owner}/adaptors/messaging"
    for filename in ("internal_commands.proto", "private_messages.proto"):
        with tempfile.TemporaryDirectory() as directory:
            stage = Path(directory)
            logical = f"cafe/internal/{owner}/{filename}"
            target = stage / logical
            target.parent.mkdir(parents=True)
            shutil.copyfile(source / filename, target)
            dependency = stage / "cafe/requests/v1/validation.proto"
            dependency.parent.mkdir(parents=True, exist_ok=True)
            shutil.copyfile(ROOT / "contracts/shared/messaging/v1/validation.proto", dependency)
            subprocess.run(
                [
                    "uv",
                    "run",
                    "--no-project",
                    "--with",
                    "grpcio-tools==1.84.0",
                    "python",
                    "-m",
                    "grpc_tools.protoc",
                    f"-I{stage}",
                    f"-I{standard_protos}",
                    f"--python_out={python_out}",
                    f"--pyi_out={python_out}",
                    logical,
                ],
                check=True,
            )
qualify_python_imports(python_out)
(python_out / "catalogue.json").write_text(json.dumps(catalogue("operations"), indent=2) + "\n")
engagement_out = ROOT / "services/engagement/src/adaptors/generated"
(engagement_out / "catalogue.json").write_text(json.dumps(catalogue("engagement"), indent=2) + "\n")
command_consumers = {item["consumer"] for item in command_subscriptions()}
topology = [{**entry, "command": entry["consumer"] in command_consumers} for entry in catalogue("topology")]
(ROOT / "services/storefront/apps/topology/generated.json").write_text(json.dumps(topology, indent=2) + "\n")
generate_broker_permissions()
shutil.copyfile(
    ROOT / "services/storefront/apps/topology/generated.json",
    ROOT / "services/storefront/apps/replay/generated.json",
)
checksums = {
    str(version): sha256((ROOT / f"devops/postgres/bootstrap/{name}").read_bytes()).hexdigest()
    for version, name in ((1, "0001_initial.sql"), (2, "0002_realtime.sql"))
}
for directory in (python_out, ROOT / "services/engagement/src/adaptors/generated"):
    (directory / "persistence.json").write_text(json.dumps(checksums, indent=2) + "\n")
subprocess.run(["node", "scripts/generate-typescript.mjs"], cwd=ROOT, check=True)

subprocess.run(["python3", "scripts/generate_request_payloads.py"], cwd=ROOT, check=True)
subprocess.run(["python3", "scripts/generate_response_payloads.py"], cwd=ROOT, check=True)
generate_migration_metadata()
generate_http_contracts()
