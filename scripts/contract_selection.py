"""Explicit service usage of published contract families; private formats stay local."""

import json
from contract_sources import ROOT, OWNERS, sources

SERVICES = ("storefront", "operations", "engagement", "api", "web")
KINDS = ("events", "requests", "realtime")


def roots(kind, owners, root=ROOT):
    """Select each declared owner's payload or request envelope as a root schema."""
    definitions = sources(kind, root)
    return [
        name
        for name in definitions
        if "/contexts/" in name
        and name.split("/contexts/")[1].split("/")[0] in owners
        and (kind != "requests" or name.endswith("_requests.proto"))
    ]


def declarations(root=ROOT):
    config = json.loads((root / "contracts/generation.json").read_text())
    if not isinstance(config, dict) or set(config) != {*KINDS, "openapi"}:
        raise ValueError("Declare events, requests, realtime and openapi separately")
    for kind, services in config.items():
        if not isinstance(services, dict) or set(services) - set(SERVICES):
            raise ValueError(kind + ": unknown service")
        available = (
            {"api", "storefront", "gateway", "provider"}
            if kind == "openapi"
            else {owner for owner in OWNERS if roots(kind, [owner], root)}
        )
        for service, roles in services.items():
            if not isinstance(roles, dict) or set(roles) != {"produces", "consumes"}:
                raise ValueError(service + ": expected produces and consumes")
            for role, entries in roles.items():
                if not isinstance(entries, list) or any(not isinstance(item, str) for item in entries):
                    raise ValueError(service + "." + role + ": expected contract names")
                if len(entries) != len(set(entries)) or set(entries) - available:
                    raise ValueError(service + "." + role + ": duplicate or unknown " + kind + " contract")
                if kind != "openapi" and entries:
                    if service == "api" and kind != "requests":
                        raise ValueError("API Protobuf contracts must be requests")
                    if service == "web" and (role != "consumes" or kind != "realtime"):
                        raise ValueError("Web consumes only realtime Protobuf contracts")
    return config


def owners(service, kind, root=ROOT):
    roles = declarations(root)[kind].get(service, {"produces": [], "consumes": []})
    selected = set(roles["produces"] + roles["consumes"])
    return tuple(owner for owner in OWNERS if owner in selected)
