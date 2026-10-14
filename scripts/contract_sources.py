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
        result[ENTRYPOINTS[kind]] = root / "contracts/shared/messaging/v1/events.proto"
        options = root / "contracts/shared/messaging/v1/event_options.proto"
        if options.exists():
            result["cafe/v1/event_options.proto"] = options
        for owner in OWNERS:
            for path in sorted((root / f"contracts/{owner}/messaging/integration_events/v1").glob("*.proto")):
                result[f"cafe/v1/contexts/{owner}/{path.name}"] = path
    elif kind == "requests":
        for name in ("common", "validation"):
            result[f"cafe/requests/v1/{name}.proto"] = root / f"contracts/shared/messaging/v1/{name}.proto"
        for owner in OWNERS:
            for name in ("commands", "queries", "replies", "requests"):
                directory = (
                    "v1" if name == "requests" else ("queries/v1" if name == "replies" else name + "/v1")
                )
                path = root / f"contracts/{owner}/messaging/{directory}/{owner}_{name}.proto"
                if path.exists():
                    result[f"cafe/requests/v1/contexts/{owner}/{path.name}"] = path
    elif kind == "realtime":
        result[ENTRYPOINTS[kind]] = root / "contracts/shared/realtime/v1/realtime.proto"
        for owner in OWNERS:
            path = root / f"contracts/{owner}/realtime/v1/{owner}_snapshots.proto"
            result[f"cafe/realtime/v1/contexts/{owner}/{path.name}"] = path
    elif kind.endswith("_private") and kind.removesuffix("_private") in PRIVATE:
        owner = kind.removesuffix("_private")
        result[f"{kind}.proto"] = root / PRIVATE[owner] / "private_messages.proto"
    else:
        raise ValueError("Unknown source group: " + kind)
    return dict(sorted(result.items()))


CONTEXT_PATHS = {
    owner: f"services/{service}/contexts/{owner}/adaptors/messaging"
    for service, owners in (
        ("storefront", ("menu", "ordering")),
        ("operations/src/operations", ("preparation", "collection")),
        ("engagement/src", ("loyalty", "communication")),
    )
    for owner in owners
}


def subscriptions(root=ROOT):
    result = []
    for owner, directory in CONTEXT_PATHS.items():
        path = root / directory / "subscriptions.json"
        for item in json.loads(path.read_text()):
            if not item["consumer"].startswith(owner + "."):
                raise ValueError("Foreign subscription identity")
            result.append(item)
    if len({item["consumer"] for item in result}) != len(result):
        raise ValueError("Duplicate subscription")
    return result


def catalogue(service: str | None = None, root=ROOT) -> list[dict]:
    from event_contracts import load

    published = load(root)["events"]
    if service not in (None, "topology"):
        from contract_selection import owners

        selected = owners(service, "events", root)
        published = [item for item in published if item["owner"] in selected]
    private_owners = {
        "storefront": ("menu",),
        "engagement": ("loyalty", "communication"),
        "operations": (),
        "topology": tuple(PRIVATE),
    }
    for owner in private_owners.get(service, ()):
        entries = json.loads((root / PRIVATE[owner] / "private_messages.json").read_text())
        if any(
            item["owner"] != owner or item["visibility"] != "domain" or "consumer" in item for item in entries
        ):
            raise ValueError("Private delivery metadata crossed its owner")
        published = [*published, *entries]
    if service == "topology":
        definitions = {item["name"]: item for item in published}
        result = []
        for sub in subscriptions(root):
            event = definitions.get(sub["event"])
            if event is None:
                raise ValueError("Subscription names an unknown event: " + sub["event"])
            if event["visibility"] == "domain" and not sub["consumer"].startswith(event["owner"] + "."):
                raise ValueError("Private subscription crossed its owner")
            result.append({**event, "consumer": sub["consumer"]})
        return sorted(result, key=lambda item: (item["name"], item["consumer"]))
    return sorted(published, key=lambda item: item["name"])


def command_subscriptions(root=ROOT) -> list[dict]:
    catalogue("topology", root)  # Validate every reference and private owner boundary.
    return [item for item in subscriptions(root) if "command" in item]


if __name__ == "__main__":
    groups = {}
    for kind in (*ENTRYPOINTS, "menu_private", "loyalty_private", "communication_private"):
        groups[kind] = {
            "entrypoint": ENTRYPOINTS.get(kind, kind + ".proto"),
            "sources": {name: path.relative_to(ROOT).as_posix() for name, path in sources(kind).items()},
        }
    from request_contracts import OWNED

    groups["requests"]["owners"] = OWNED
    groups["events"]["subscriptions"] = catalogue("topology")
    print(json.dumps(groups))
