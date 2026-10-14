"""Decision 0017 explicitly replaces the original published event envelope."""

from copy import deepcopy


def accepted_baseline(old, new):
    """Retire the global payload oneof and accept newly declared boundary rules once.

    Current-format baselines receive ordinary compatibility comparison. Unrelated
    interfaces and the existing metadata and payload field identities stay checked.
    """
    name = "cafe.v1.Event"
    before = old.get(name, {}).get("fields", {})
    after = new.get(name, {}).get("fields", {})
    retired = {key: field for key, field in before.items() if field["oneof"] == "payload"}
    replacement = after.get("12", {})
    if not retired or replacement.get("name") != "payload" or replacement.get("type") != 12:
        return old
    result = deepcopy(old)
    for key in retired:
        del result[name]["fields"][key]
    pending = [field["type_name"].lstrip(".") for field in retired.values()]
    seen = set()
    while pending:
        typ = pending.pop()
        if typ in seen:
            continue
        seen.add(typ)
        for key, field in result.get(typ, {}).get("fields", {}).items():
            current = new.get(typ, {}).get("fields", {}).get(key)
            if current and not field["options"]:
                field["options"] = current["options"]
            if field["type"] == 11:
                pending.append(field["type_name"].lstrip("."))
    return result
