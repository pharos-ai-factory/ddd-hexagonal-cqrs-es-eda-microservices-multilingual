"""Compile declared request/reply roots and their required schema dependencies."""

from pathlib import Path
import subprocess
import tempfile
from contract_sources import sources
from contract_selection import owners as selected_owners
from protobuf_selection import plan
from request_contracts import envelopes


def generate_requests(service, root, plugin, standard_protos):
    owners = selected_owners(service, "requests", root)
    go_relative = (
        "services/api/adaptors/messaging/generated"
        if service == "api"
        else "services/storefront/contracts/requests/generated"
    )
    output = (
        root / go_relative
        if service != "operations"
        else root / "services/operations/src/operations/adaptors/generated"
    )
    with tempfile.TemporaryDirectory() as directory:
        staging = Path(directory)
        generated = staging / "generated"
        generated.mkdir()
        descriptor = staging / "requests.pb"
        inputs = sorted(plan("requests", owners, descriptor, root))
        definitions = sources("requests", root)
        for name in inputs:
            target = staging / name
            target.parent.mkdir(parents=True, exist_ok=True)
            target.write_bytes(definitions[name].read_bytes())
        arguments = [
            "uv",
            "run",
            "--no-project",
            "--with",
            "grpcio-tools==1.84.0",
            "python",
            "-m",
            "grpc_tools.protoc",
            "-I" + str(staging),
            "-I" + standard_protos,
        ]
        if service == "operations":
            arguments += ["--python_out=" + str(generated), "--pyi_out=" + str(generated)]
        else:
            package = (
                "github.com/pharos-ai-factory/ddd-hexagonal-cqrs-es-eda-microservices-multilingual/"
                + go_relative
            )
            arguments += [
                "--plugin=protoc-gen-go=" + str(plugin),
                "--go_out=" + str(generated),
                "--go_opt=module=" + package,
            ]
            if service == "api":
                for name in inputs:
                    owner = name.split("/contexts/")[1].split("/")[0] if "/contexts/" in name else None
                    arguments.append(
                        "--go_opt=M"
                        + name
                        + "="
                        + package
                        + "/"
                        + str(Path(name).parent)
                        + ("/shared" if Path(name).stem in ("common", "validation") else "")
                        + ";"
                        + (
                            owner + "v1"
                            if owner
                            else ("sharedv1" if Path(name).stem in ("common", "validation") else "requestsv1")
                        )
                    )
        if inputs:
            subprocess.run([*arguments, *inputs], cwd=root, check=True)
            if service != "operations":
                target = generated / "cafe/requests/v1/envelopes.go"
                target.parent.mkdir(parents=True, exist_ok=True)
                target.write_text(envelopes(owners, package))
                subprocess.run(
                    ["python3", str(Path(__file__).with_name("go.py")), "gofmt", "-w", str(target)],
                    cwd=root,
                    check=True,
                )
        # Compilation must succeed before old bindings are replaced or pruned.
        managed = output if service != "operations" else output / "cafe/requests"
        patterns = ("*.go",) if service != "operations" else ("*_pb2.py", "*_pb2.pyi")
        for pattern in patterns:
            for path in managed.rglob(pattern):
                path.unlink()
        for path in generated.rglob("*"):
            if path.is_file():
                target = output / path.relative_to(generated)
                target.parent.mkdir(parents=True, exist_ok=True)
                target.write_bytes(path.read_bytes())
