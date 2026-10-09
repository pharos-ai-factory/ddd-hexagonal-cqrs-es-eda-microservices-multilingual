"""Locate published interfaces and owner-local delivery implementation resources."""
from pathlib import Path
import json

ROOT = Path(__file__).resolve().parents[1]
OWNERS = ("menu", "ordering", "preparation", "collection", "loyalty", "communication")
PRIVATE = {
    "menu": "services/storefront/contexts/menu/adaptors/messaging",
    "loyalty": "services/engagement/src/contexts/loyalty/adaptors/messaging",
    "communication": "services/engagement/src/contexts/communication/adaptors/messaging",
}
ENTRYPOINTS = {
    "events": "cafe/v1/events.proto",
    "requests": "cafe/requests/v1/common.proto",
    "realtime": "cafe/realtime/v1/realtime.proto",
}


def sources(kind: str, root: Path = ROOT) -> dict[str, Path]:
    result = {}
    if kind == "events":
        result[ENTRYPOINTS[kind]] = root/"contracts/shared/messaging/v1/events.proto"
        for owner in OWNERS:
            path = root/f"contracts/{owner}/messaging/integration_events/v1/{owner}_integration_events.proto"
            if path.exists():
                result[f"cafe/v1/contexts/{owner}/{path.name}"] = path
    elif kind == "requests":
        for name in ("common", "validation"):
            result[f"cafe/requests/v1/{name}.proto"] = root/f"contracts/shared/messaging/v1/{name}.proto"
        for owner in OWNERS:
            for name in ("commands", "queries", "replies", "requests"):
                directory = "v1" if name == "requests" else ("queries/v1" if name == "replies" else name+"/v1")
                path = root/f"contracts/{owner}/messaging/{directory}/{owner}_{name}.proto"
                if path.exists():
                    result[f"cafe/requests/v1/contexts/{owner}/{path.name}"] = path
    elif kind == "realtime":
        result[ENTRYPOINTS[kind]] = root/"contracts/shared/realtime/v1/realtime.proto"
        for owner in OWNERS:
            path = root/f"contracts/{owner}/realtime/v1/{owner}_snapshots.proto"
            result[f"cafe/realtime/v1/contexts/{owner}/{path.name}"] = path
    elif kind.endswith("_private") and kind.removesuffix("_private") in PRIVATE:
        owner = kind.removesuffix("_private")
        result[f"{kind}.proto"] = root/PRIVATE[owner]/"private_messages.proto"
    else:
        raise ValueError("Unknown source group: "+kind)
    return dict(sorted(result.items()))


def catalogue(service: str | None = None) -> list[dict]:
    published = []
    for owner in OWNERS:
        path = ROOT/f"contracts/{owner}/messaging/integration_events/v1/catalogue.json"
        if path.exists():
            entries = json.loads(path.read_text())
            if any(item["owner"] != owner or item["visibility"] != "integration" for item in entries):
                raise ValueError("Published catalogue contains a private or foreign event")
            published.extend(entries)
    private_owners = {"storefront": ("menu",), "engagement": ("loyalty", "communication"),
                      "operations": (), "topology": tuple(PRIVATE)}
    for owner in private_owners.get(service, ()):
        entries = json.loads((ROOT/PRIVATE[owner]/"private_messages.json").read_text())
        if any(item["owner"] != owner or item["visibility"] != "domain" or
               item["consumer"].split(".")[0] != owner for item in entries):
            raise ValueError("Private delivery metadata crossed its owner")
        published.extend(entries)
    return sorted(published, key=lambda item: item["name"])


def command_subscriptions(root: Path = ROOT) -> list[dict]:
    result = []
    for service, owners in (("operations/src/operations", ("preparation", "collection")),
                            ("engagement/src", ("loyalty", "communication"))):
        for owner in owners:
            entries = json.loads((root/f"services/{service}/contexts/{owner}/adaptors/messaging/subscriptions.json").read_text())
            if any(not item["consumer"].startswith(owner+".") for item in entries):
                raise ValueError("Foreign command subscription")
            result.extend(entries)
    identities = [entry['consumer'] for entry in result]
    if len(set(identities)) != len(identities):
        raise ValueError("Duplicate command subscription")
    events = {entry['consumer']: entry['name'] for entry in catalogue('topology')}
    if any(events.get(entry['consumer']) != entry['event'] for entry in result):
        raise ValueError("Command subscription disagrees with its incoming event")
    return result


if __name__ == "__main__":
    groups = {}
    for kind in (*ENTRYPOINTS, "menu_private", "loyalty_private", "communication_private"):
        groups[kind] = {"entrypoint": ENTRYPOINTS.get(kind, kind+".proto"),
                        "sources": {name: path.relative_to(ROOT).as_posix() for name, path in sources(kind).items()}}
    from request_contracts import OWNED
    groups["requests"]["owners"] = OWNED
    print(json.dumps(groups))
