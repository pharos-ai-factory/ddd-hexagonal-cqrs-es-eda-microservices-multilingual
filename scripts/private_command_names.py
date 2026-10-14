"""Read private command identifiers from their owner Protobuf envelopes."""

import json
from pathlib import Path
import subprocess
import sys
from contract_sources import PRIVATE


def read(owner, root):
    source = root / PRIVATE[owner] / "internal_commands.proto"
    if not source.exists():
        return []
    return json.loads(
        subprocess.check_output(
            [
                "uv",
                "run",
                "--no-project",
                "--with",
                "grpcio-tools==1.84.0",
                "python",
                str(Path(__file__).resolve()),
                str(source),
            ],
            text=True,
        )
    )


def compile_names(source):
    from google.protobuf.descriptor_pb2 import FileDescriptorSet
    from grpc_tools import protoc
    import tempfile

    with tempfile.TemporaryDirectory() as directory:
        output = Path(directory) / "commands.pb"
        if protoc.main(
            ["protoc", "-I" + str(source.parent), "--descriptor_set_out=" + str(output), str(source)]
        ):
            raise ValueError("Cannot compile private commands: " + str(source))
        descriptor = FileDescriptorSet.FromString(output.read_bytes())
        envelope = next(m for f in descriptor.file for m in f.message_type if m.name == "CommandEnvelope")
        index = next(i for i, oneof in enumerate(envelope.oneof_decl) if oneof.name == "command")
        return [
            (f.name, f.json_name)
            for f in envelope.field
            if f.HasField("oneof_index") and f.oneof_index == index
        ]


if __name__ == "__main__":
    print(json.dumps(compile_names(Path(sys.argv[1]))))
