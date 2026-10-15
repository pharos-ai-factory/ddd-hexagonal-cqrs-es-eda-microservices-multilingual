"""Inspect command/query payloads, presence and visibility from compiler descriptors."""

from functools import lru_cache
import json
from pathlib import Path
import subprocess
import sys
from contract_sources import ROOT, PRIVATE, sources


def private_sources(owner, root=ROOT):
    return {
        f"cafe/internal/{owner}/internal_commands.proto": root / PRIVATE[owner] / "internal_commands.proto",
        "cafe/requests/v1/validation.proto": root / "contracts/shared/messaging/v1/validation.proto",
    }


def inspect(root=ROOT):
    from google.protobuf import descriptor_pool, message_factory
    from protobuf_selection import compile_sources

    definitions = sources("requests", root)
    for owner, directory in PRIVATE.items():
        if (root / directory / "internal_commands.proto").exists():
            definitions.update(private_sources(owner, root))
    descriptors = compile_sources(definitions)
    pool = descriptor_pool.DescriptorPool()
    for file in descriptors.file:
        pool.Add(file)
    options_type = message_factory.GetMessageClass(
        pool.FindMessageTypeByName("google.protobuf.MessageOptions")
    )
    fields_type = message_factory.GetMessageClass(pool.FindMessageTypeByName("google.protobuf.FieldOptions"))
    contract = pool.FindExtensionByName("cafe.requests.v1.contract")
    required = pool.FindExtensionByName("cafe.requests.v1.required_input")
    messages = {
        file.package + "." + m.name: (file, m) for file in descriptors.file for m in file.message_type
    }
    payloads = []
    for file in descriptors.file:
        if file.name.endswith("_requests.proto"):
            owner, visibility = file.package.split(".")[1], "PUBLIC"
            envelopes = [(m, m.name, "payload") for m in file.message_type if m.name in ("Command", "Query")]
        elif file.name.endswith("internal_commands.proto"):
            owner, visibility = file.name.split("/")[2], "PRIVATE"
            envelopes = [(m, "Command", "command") for m in file.message_type if m.name == "CommandEnvelope"]
        else:
            continue
        for envelope, role, oneof in envelopes:
            index = next(i for i, item in enumerate(envelope.oneof_decl) if item.name == oneof)
            for field in envelope.field:
                if not field.HasField("oneof_index") or field.oneof_index != index:
                    continue
                source, message = messages[field.type_name.lstrip(".")]
                options = options_type.FromString(message.options.SerializeToString())
                if not options.HasExtension(contract):
                    raise ValueError(message.name + ": command/query must declare owner and visibility")
                annotation = options.Extensions[contract]
                declared = {1: "PUBLIC", 2: "PRIVATE"}.get(annotation.visibility)
                source_owner = (
                    source.name.split("/contexts/")[1].split("/")[0]
                    if "/contexts/" in source.name
                    else source.name.split("/")[2]
                )
                if annotation.owner != owner or source_owner != owner or declared != visibility:
                    raise ValueError(message.name + ": owner/visibility disagrees with its delivery envelope")
                payloads.append(
                    {
                        "owner": owner,
                        "visibility": visibility,
                        "role": role,
                        "name": message.name if message.name.endswith(role) else message.name + role,
                        "wire_name": message.name,
                        "type": field.type_name.lstrip("."),
                        "route": field.json_name,
                        "field": field.name,
                        "source": source.name,
                    }
                )
    if len({(m["owner"], m["name"]) for m in payloads}) != len(payloads):
        raise ValueError("Duplicate command/query payload name")
    annotated = {
        name
        for name, (_, m) in messages.items()
        if options_type.FromString(m.options.SerializeToString()).HasExtension(contract)
    }
    if annotated != {m["type"] for m in payloads} or len(annotated) != len(payloads):
        raise ValueError("Every annotated command/query must belong to exactly one delivery envelope")
    helpers = [
        {
            "owner": "shared",
            "name": "PageRequest",
            "wire_name": "PageRequest",
            "type": "cafe.requests.v1.PageRequest",
            "role": "Value",
            "visibility": "PUBLIC",
            "source": "cafe/requests/v1/common.proto",
        }
    ]
    for payload in [*payloads, *helpers]:
        _, message = messages[payload["type"]]
        if (
            message.nested_type
            or message.enum_type
            or any(
                len(members := [f for f in message.field if f.HasField("oneof_index") and f.oneof_index == i])
                != 1
                or not members[0].proto3_optional
                for i, _ in enumerate(message.oneof_decl)
            )
        ):
            raise ValueError(message.name + ": nested types/enums/oneofs require a generator extension")
        fields = []
        for f in message.field:
            if f.type not in (3, 4, 5, 8, 9, 11, 13) or f.label == f.LABEL_REPEATED:
                raise ValueError(message.name + ": unsupported request field shape")
            if f.type == 11 and (
                payload["role"] != "Query" or f.type_name != ".cafe.requests.v1.PageRequest"
            ):
                raise ValueError(message.name + ": unsupported nested request payload")
            rules = fields_type.FromString(f.options.SerializeToString())
            is_required = rules.Extensions[required]
            if is_required and not (f.proto3_optional or f.type == 11):
                raise ValueError(message.name + ": required input needs explicit presence")
            fields.append(
                {
                    "name": f.name,
                    "json": f.json_name,
                    "type": f.type,
                    "presence": f.proto3_optional or f.type == 11,
                    "required": is_required,
                    "optional": (f.proto3_optional or f.type == 11) and not is_required,
                }
            )
        payload["fields"] = fields
    return {"payloads": payloads, "helpers": helpers}


@lru_cache
def load(root=ROOT):
    try:
        output = subprocess.check_output(
            [
                "uv",
                "run",
                "--no-project",
                "--with",
                "grpcio-tools==1.84.0",
                "python",
                str(Path(__file__).resolve()),
                str(root),
            ],
            text=True,
            stderr=subprocess.PIPE,
        )
    except subprocess.CalledProcessError as error:
        raise ValueError(error.stderr.strip().splitlines()[-1]) from error
    return json.loads(output)


if __name__ == "__main__":
    print(json.dumps(inspect(Path(sys.argv[1]))))
