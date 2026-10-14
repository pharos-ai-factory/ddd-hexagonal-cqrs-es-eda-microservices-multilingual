"""Generate service-specific public event and realtime bindings from descriptors."""

from pathlib import Path
import re
import subprocess
import tempfile
from contract_sources import ROOT
from contract_selection import owners
from protobuf_selection import plan


def generate_selected(plugin, root=ROOT):
    for kind in ("events", "realtime"):
        for service in ("storefront", "operations", "engagement", "web"):
            if service == "storefront":
                output = root / f"services/storefront/contracts/{kind}/generated"
                patterns = ("*.pb.go",)
            elif service == "operations":
                output = root / "services/operations/src/operations/adaptors/generated"
                patterns = ("*_pb2.py", "*_pb2.pyi")
            else:
                output = root / f"services/{service}/src/adaptors/generated"
                patterns = (kind + ".json",)
            with tempfile.TemporaryDirectory() as directory:
                stage = Path(directory)
                descriptor = stage / "schema.pb"
                names = plan(kind, owners(service, kind, root), descriptor, root)
                generated = stage / "generated"
                generated.mkdir()
                if names:
                    arguments = [
                        "uv",
                        "run",
                        "--no-project",
                        "--with",
                        "grpcio-tools==1.84.0",
                        "python",
                        "-m",
                        "grpc_tools.protoc",
                        "--descriptor_set_in=" + str(descriptor),
                    ]
                    if service == "storefront":
                        module = (
                            "github.com/pharos-ai-factory/ddd-hexagonal-cqrs-es-eda-microservices-multilingual/"
                            + output.relative_to(root).as_posix()
                        )
                        arguments += [
                            f"--plugin=protoc-gen-go={plugin}",
                            f"--go_out={generated}",
                            "--go_opt=module=" + module,
                        ]
                    elif service == "operations":
                        arguments += [f"--python_out={generated}", f"--pyi_out={generated}"]
                    else:
                        arguments = [
                            "node",
                            "scripts/descriptor_json.mjs",
                            str(descriptor),
                            str(generated / (kind + ".json")),
                        ]
                    subprocess.run(
                        arguments + (names if service in ("storefront", "operations") else []),
                        cwd=root,
                        check=True,
                    )
                managed = (
                    output / ("cafe/v1" if kind == "events" else "cafe/realtime/v1")
                    if service == "operations"
                    else output
                )
                for pattern in patterns:
                    for path in managed.rglob(pattern):
                        path.unlink()
                for path in generated.rglob("*"):
                    if path.is_file():
                        target = output / path.relative_to(generated)
                        target.parent.mkdir(parents=True, exist_ok=True)
                        target.write_bytes(path.read_bytes())


def qualify_python_imports(output):
    for path in (output / "cafe").rglob("*"):
        if path.suffix in {".py", ".pyi"}:
            path.write_text(
                re.sub(
                    r"(?m)^from (cafe(?:\.[\w]+)*) import ",
                    r"from operations.adaptors.generated.\1 import ",
                    path.read_text(),
                )
            )
