"""Generate RabbitMQ authority from owner subscriptions and request contracts."""

import json
import re

from contract_selection import owners
from request_payloads import load as request_payloads
from contract_sources import ROOT, OWNERS, catalogue, subscriptions

GENERATED = "services/storefront/apps/topology/permissions.generated.json"
TOPICS = ("cafe.events", "cafe.requests", "cafe.replies")


def exact(names):
    """Escape and anchor an explicit resource or routing-key allow-list."""
    return "^(" + "|".join(re.escape(name) for name in sorted(set(names))) + ")$" if names else "^$"


def build(root=ROOT):
    """Include durable commands, retries, dead letters and API request/reply queues."""
    payloads = request_payloads(root)["payloads"]
    selected = owners("api", "requests", root)
    public_owners = {m["owner"] for m in payloads if m["visibility"] == "PUBLIC"}
    request_owners = tuple(owner for owner in selected if owner in public_owners)
    private = {(m["owner"], m["route"]) for m in payloads if m["visibility"] == "PRIVATE"}
    for sub in subscriptions(root):
        if "command" in sub and (sub["consumer"].split(".")[0], sub["command"]) not in private:
            raise ValueError("Durable subscription must name an owner-private command")
    events = catalogue("topology", root)
    commands = {item["consumer"] for item in subscriptions(root) if "command" in item}
    delivery = {owner: f"ref.{owner}.delivery" for owner in OWNERS}
    exchanges = {*TOPICS, *delivery.values()}
    queues, dead_letter_sources = {}, set()

    def queue(name, owner, dead_letters=False):
        if name in queues or name in exchanges:
            raise ValueError("Broker resource name collision: " + name)
        queues[name] = owner
        if dead_letters:
            dead_letter_sources.add(name)

    for event in events:
        consumer = event["consumer"]
        owner = consumer.split(".", 1)[0]
        if owner not in OWNERS or not re.fullmatch(r"[a-z]+\.[a-z][a-z0-9.-]*", consumer):
            raise ValueError("Invalid broker consumer identity: " + consumer)
        variants = [consumer, consumer + ".command"] if consumer in commands else [consumer]
        for variant in variants:
            for suffix in ("", ".retry", ".dead"):
                queue("ref." + variant + suffix, owner, suffix == ".retry")
    for owner in request_owners:
        for suffix in ("commands", "queries"):
            queue(f"ref.{owner}.{suffix}", owner, True)
            queue(f"ref.{owner}.{suffix}.dead", owner)
        queue(f"ref.api.{owner}.replies", "api")

    users = {}

    def user(name, configure, write, read, topic_write=None, topic_read=None):
        users["cafe_" + name] = {
            "password_key": name.upper() + "_BROKER_PASSWORD",
            "permissions": {"configure": exact(configure), "write": exact(write), "read": exact(read)},
            "topics": [
                {
                    "exchange": exchange,
                    "write": (topic_write or {}).get(exchange, "^$"),
                    "read": (topic_read or {}).get(exchange, "^$"),
                }
                for exchange in TOPICS
            ],
        }

    request_keys = [f"request.{owner}.{kind}" for owner in request_owners for kind in ("command", "query")]
    reply_keys = [f"reply.{owner}" for owner in request_owners]
    for owner in OWNERS:
        user(
            owner,
            [],
            ["cafe.events", delivery[owner], *(["cafe.replies"] if owner in request_owners else [])],
            [name for name, context in queues.items() if context == owner],
            {
                "cafe.events": rf"^(domain|integration)\.{re.escape(owner)}\..+$",
                "cafe.replies": exact([f"reply.{owner}"] if owner in request_owners else []),
            },
        )
    user(
        "api",
        [],
        ["cafe.requests"],
        [name for name, context in queues.items() if context == "api"],
        {"cafe.requests": exact(request_keys)},
    )
    # RabbitMQ requires queue read and DLX write when declaring dead-letter arguments.
    user(
        "topology",
        exchanges | queues.keys(),
        queues.keys() | set(delivery.values()),
        exchanges | dead_letter_sources,
        topic_read={
            "cafe.events": exact([f"{item['visibility']}.{item['name']}" for item in events]),
            "cafe.requests": exact(request_keys),
            "cafe.replies": exact(reply_keys),
        },
    )
    return {
        "request_owners": list(request_owners),
        "request_contracts": [
            {key: m[key] for key in ("owner", "name", "role", "visibility", "route")} for m in payloads
        ],
        "queues": {
            owner: sorted(name for name, context in queues.items() if context == owner)
            for owner in (*OWNERS, "api")
        },
        "exchanges": sorted(exchanges),
        "dead_letter_sources": sorted(dead_letter_sources),
        "users": users,
    }


def generate(root=ROOT):
    (root / GENERATED).write_text(json.dumps(build(root), indent=2) + "\n")
