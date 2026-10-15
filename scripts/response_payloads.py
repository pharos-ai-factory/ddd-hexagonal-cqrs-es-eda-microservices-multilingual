"""Resolve schema response associations and their complete plain payload graph."""


def inspect_responses(requests, messages, pool, fields_type):
    required = pool.FindExtensionByName("cafe.requests.v1.required_output")
    result = {}
    active = set()

    def visit(name, owner):
        if name not in messages:
            raise ValueError("Unknown response: " + name)
        file, message = messages[name]
        scope = "shared" if file.package == "cafe.requests.v1" else file.package.split(".")[1]
        if scope not in (owner, "shared"):
            raise ValueError("Foreign response ownership: " + name)
        if name in active:
            raise ValueError("Recursive responses require a generator extension: " + name)
        if name in result:
            return
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
            raise ValueError(
                "Response nested declarations/enums/oneofs require a generator extension: " + name
            )
        active.add(name)
        fields = []
        for f in message.field:
            if f.type not in (3, 4, 5, 8, 9, 11, 13):
                raise ValueError("Unsupported response field: " + name + "." + f.name)
            rule = fields_type.FromString(f.options.SerializeToString()).Extensions[required]
            repeated = f.label == f.LABEL_REPEATED
            if rule and (f.type != 11 or repeated):
                raise ValueError("Required output applies to singular message fields: " + name)
            target = f.type_name.lstrip(".")
            if f.type == 11:
                visit(target, owner)
            fields.append(
                {
                    "name": f.name,
                    "json": f.json_name,
                    "type": f.type,
                    "repeated": repeated,
                    "target": target,
                    "optional": (f.proto3_optional or (f.type == 11 and not rule)) and not repeated,
                }
            )
        active.remove(name)
        result[name] = {"name": message.name, "type": name, "owner": scope, "fields": fields}

    for request in requests:
        response = request["response"]
        if not response:
            raise ValueError(request["name"] + ": must declare its response")
        owner = request["owner"]
        if request["role"] == "Query" and not response.startswith(f"cafe.{owner}.requests.v1."):
            raise ValueError(request["name"] + ": query response must belong to its owner")
        if request["role"] == "Command" and response != "cafe.requests.v1.Outcome":
            raise ValueError(request["name"] + ": command executor requires Outcome response")
        if request["visibility"] == "PUBLIC":
            _, envelope = messages[f"cafe.{owner}.requests.v1.Reply"]
            alternatives = [f for f in envelope.field if f.HasField("oneof_index")]
            matches = [f for f in alternatives if f.type_name == "." + response]
            errors = [
                f
                for f in alternatives
                if f.name == "error" and f.type_name == ".cafe.requests.v1.RequestError"
            ]
            if len(matches) != 1 or len(errors) != 1:
                raise ValueError(
                    request["name"] + ": response must select one owner Reply alternative and RequestError"
                )
            if matches[0].name == "error":
                raise ValueError(request["name"] + ": success response cannot be RequestError")
            request["response_field"] = matches[0].name
            request["response_json"] = matches[0].json_name
        visit(response, owner)
    visit("cafe.requests.v1.RequestError", "shared")
    return list(result.values())
