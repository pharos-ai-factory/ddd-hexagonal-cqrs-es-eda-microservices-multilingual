"""Read published event definitions and portable constraints from compiler descriptors."""

from functools import lru_cache
import json
from pathlib import Path
import re
import subprocess
import sys

ROOT = Path(__file__).resolve().parents[1]


def inspect(root):
    from google.protobuf import descriptor_pool, message_factory
    from protobuf_selection import compile_descriptors

    descriptors = compile_descriptors("events", root)
    pool = descriptor_pool.DescriptorPool()
    for file in descriptors.file:
        pool.Add(file)
    message_options = message_factory.GetMessageClass(
        pool.FindMessageTypeByName("google.protobuf.MessageOptions")
    )
    field_options = message_factory.GetMessageClass(
        pool.FindMessageTypeByName("google.protobuf.FieldOptions")
    )
    event_option = pool.FindExtensionByName("cafe.events.event")
    rules_option = pool.FindExtensionByName("cafe.events.rules")
    messages, events = [], []
    for file in descriptors.file:
        if "/contexts/" not in file.name:
            continue
        owner = file.name.split("/contexts/")[1].split("/")[0]
        for message in file.message_type:
            if message.nested_type or message.enum_type or message.oneof_decl:
                raise ValueError(
                    message.name
                    + ": nested types, enums and presence/oneof fields need an explicit generator extension"
                )
            fields = []
            for field in message.field:
                if field.type not in (3, 4, 5, 8, 9, 11, 13) or field.proto3_optional:
                    raise ValueError(
                        message.name + "." + field.name + ": unsupported event field type or presence"
                    )
                options = field_options.FromString(field.options.SerializeToString())
                rules = {f.name: value for f, value in options.Extensions[rules_option].ListFields()}
                repeated = field.label == field.LABEL_REPEATED
                if repeated and field.type != 11:
                    raise ValueError(message.name + ": repeated fields must contain messages")
                if (
                    any(
                        key in rules
                        for key in ("uuid", "non_empty", "timestamp", "pattern", "source_identity")
                    )
                    and field.type != 9
                ):
                    raise ValueError(field.name + ": string constraint on another field type")
                if any(key in rules for key in ("minimum", "maximum")) and field.type not in (3, 4, 5, 13):
                    raise ValueError(field.name + ": numeric constraint on another field type")
                if any(key in rules for key in ("min_items", "max_items")) and not repeated:
                    raise ValueError(field.name + ": collection constraint on a scalar")
                if rules.get("source_identity") and not rules.get("uuid"):
                    raise ValueError(field.name + ": source identity must be a UUID")
                if rules.get("minimum", -(2**53 - 1)) > rules.get("maximum", 2**53 - 1):
                    raise ValueError(field.name + ": inverted numeric bounds")
                if rules.get("min_items", 0) > rules.get("max_items", 2**32 - 1):
                    raise ValueError(field.name + ": inverted collection bounds")
                if any(abs(rules[key]) > 2**53 - 1 for key in ("minimum", "maximum") if key in rules):
                    raise ValueError(field.name + ": numeric constraint exceeds the portable range")
                if "pattern" in rules:
                    try:
                        re.compile(rules["pattern"])
                    except re.error as error:
                        raise ValueError(field.name + ": invalid pattern") from error
                    if (
                        not rules["pattern"].startswith("^")
                        or not rules["pattern"].endswith("$")
                        or "(?" in rules["pattern"]
                        or "\\" in rules["pattern"]
                    ):
                        raise ValueError(
                            field.name + ": use an anchored portable pattern without escapes or lookarounds"
                        )
                fields.append(
                    {
                        "name": field.name,
                        "json": field.json_name,
                        "type": field.type,
                        "message": field.type_name.lstrip("."),
                        "repeated": repeated,
                        "rules": rules,
                    }
                )
            value = {
                "name": message.name,
                "type": file.package + "." + message.name,
                "owner": owner,
                "source": file.name,
                "fields": fields,
            }
            messages.append(value)
            options = message_options.FromString(message.options.SerializeToString())
            if options.HasExtension(event_option):
                contract = options.Extensions[event_option]
                if contract.owner != owner or not re.fullmatch(
                    re.escape(owner) + r"\.[a-z][a-z0-9-]*", contract.name
                ):
                    raise ValueError(message.name + ": invalid event name or foreign owner")
                if not contract.kind or contract.version < 1:
                    raise ValueError(message.name + ": incomplete event contract")
                if sum(bool(f["rules"].get("source_identity")) for f in fields) > 1:
                    raise ValueError(message.name + ": multiple source identities")
                events.append(
                    {
                        "name": contract.name,
                        "owner": owner,
                        "kind": contract.kind,
                        "version": contract.version,
                        "visibility": "integration",
                        "type": value["type"],
                    }
                )
    for key in ("name", "type"):
        if len({event[key] for event in events}) != len(events):
            raise ValueError("Duplicate event " + key)
    # The current language targets have flat payload namespaces. Fail on collisions.
    if len({message["name"] for message in messages}) != len(messages):
        raise ValueError("Duplicate payload type name")
    by_type = {message["type"]: message for message in messages}

    def visit(typ, trail):
        if typ in trail:
            raise ValueError("Recursive event payload: " + typ)
        for field in by_type[typ]["fields"]:
            if field["message"]:
                if field["message"] not in by_type:
                    raise ValueError("Unsupported external payload: " + field["message"])
                visit(field["message"], trail | {typ})

    for message in messages:
        visit(message["type"], set())
    return {"events": sorted(events, key=lambda e: e["name"]), "messages": messages}


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
