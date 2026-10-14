"""Build scoped compiler descriptors without changing authoritative wire schemas."""

import argparse
from pathlib import Path
import subprocess
import tempfile
from contract_sources import ENTRYPOINTS, ROOT, sources
from contract_selection import roots


def compile_descriptors(kind, root=ROOT):
    from google.protobuf.descriptor_pb2 import FileDescriptorSet
    import grpc_tools
    from grpc_tools import protoc

    with tempfile.TemporaryDirectory() as directory:
        stage = Path(directory)
        definitions = sources(kind, root)
        for logical, source in definitions.items():
            target = stage / logical
            target.parent.mkdir(parents=True, exist_ok=True)
            target.write_bytes(source.read_bytes())
        output = stage / "all.pb"
        result = protoc.main(
            [
                "protoc",
                "-I" + str(stage),
                "-I" + str(Path(grpc_tools.__file__).parent / "_proto"),
                "--include_imports",
                "--descriptor_set_out=" + str(output),
                *definitions,
            ]
        )
        if result:
            raise ValueError("Cannot compile " + kind + " contracts")
        return FileDescriptorSet.FromString(output.read_bytes())


def select(descriptors, kind, selected_roots):
    """Keep selected payloads in the technical oneof and follow actual imports."""
    from google.protobuf.descriptor_pb2 import FileDescriptorSet

    result = FileDescriptorSet()
    result.CopyFrom(descriptors)
    files = {file.name: file for file in result.file}
    selected = set(selected_roots)
    if not selected <= files.keys():
        raise ValueError("Unknown root schemas: " + ", ".join(sorted(selected - files.keys())))
    if not selected:
        result.ClearField("file")
        return result
    if kind == "events":
        selected.add(ENTRYPOINTS[kind])
    if kind == "realtime":
        envelope = files[ENTRYPOINTS[kind]]
        message_name, oneof = ("Event", "payload") if kind == "events" else ("Publication", "snapshot")
        message = next(item for item in envelope.message_type if item.name == message_name)
        index = next(i for i, item in enumerate(message.oneof_decl) if item.name == oneof)
        type_files = {
            "." + file.package + "." + message.name: file.name
            for file in files.values()
            for message in file.message_type
        }
        fields = [
            field
            for field in message.field
            if not field.HasField("oneof_index")
            or field.oneof_index != index
            or type_files[field.type_name] in selected
        ]
        # Reserve omitted alternatives in the local view. The published source
        # remains complete and owns every original field number and name.
        for field in message.field:
            if field not in fields:
                message.reserved_range.add(start=field.number, end=field.number + 1)
                message.reserved_name.append(field.name)
        del message.field[:]
        message.field.extend(fields)
        needed = {type_files[field.type_name] for field in fields if field.type_name}
        dependencies = [name for name in envelope.dependency if name in needed]
        if envelope.public_dependency or envelope.weak_dependency:
            raise ValueError("Technical envelope imports must be ordinary imports")
        del envelope.dependency[:]
        envelope.dependency.extend(dependencies)
        envelope.ClearField("source_code_info")
        selected.add(envelope.name)
    pending, closure = list(selected), set()
    while pending:
        name = pending.pop()
        if name in closure:
            continue
        closure.add(name)
        pending.extend(files[name].dependency)
    retained = [file for file in result.file if file.name in closure]
    del result.file[:]
    result.file.extend(retained)
    return result


def plan(kind, selected_owners, destination, root=ROOT):
    """Run compiler-only dependencies in the pinned environment."""
    output = subprocess.check_output(
        [
            "uv",
            "run",
            "--no-project",
            "--with",
            "grpcio-tools==1.84.0",
            "python",
            str(Path(__file__).resolve()),
            "--root",
            str(root),
            "--kind",
            kind,
            "--owners",
            *selected_owners,
            "--output",
            str(destination),
        ],
        text=True,
    )
    return [name for name in output.splitlines() if name]


if __name__ == "__main__":
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("--root", type=Path, default=ROOT)
    parser.add_argument("--kind", required=True, choices=("events", "requests", "realtime"))
    parser.add_argument("--owners", nargs="*", required=True)
    parser.add_argument("--output", type=Path, required=True)
    args = parser.parse_args()
    definitions = sources(args.kind, args.root)
    descriptor = select(
        compile_descriptors(args.kind, args.root), args.kind, roots(args.kind, args.owners, args.root)
    )
    args.output.write_bytes(descriptor.SerializeToString(deterministic=True))
    print("\n".join(file.name for file in descriptor.file if file.name in definitions))
