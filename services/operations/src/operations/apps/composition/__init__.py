"""Dependency Injector containers own infrastructure and resource lifetimes."""
from collections.abc import Iterator
import json
from pathlib import Path
from operations.foundation.domain import record, text
from operations.apps.runtime import SubscriptionDefinition
from operations.adaptors.postgres import PostgresContextDatabase


def context_database(owner: str, url: str) -> Iterator[PostgresContextDatabase]:
    """Close a context's pool after its receiving and publishing workers have stopped."""
    database = PostgresContextDatabase(owner, url)
    try:
        yield database
    finally:
        database.pool.close()


def subscription_definitions(owner: str) -> list[SubscriptionDefinition]:
    """Read the same owner declarations used to generate durable broker topology."""
    path = Path(__file__).parents[2]/"contexts"/owner/"adaptors/messaging/subscriptions.json"
    values = json.loads(path.read_text())
    if not isinstance(values, list):
        raise ValueError("Subscription declarations must be a list")
    result = []
    for value in values:
        item = record(value)
        consumer = text(item["consumer"])
        if not consumer.startswith(owner+"."):
            raise ValueError("Foreign subscription")
        result.append(SubscriptionDefinition(consumer, text(item["event"]), text(item["command"])))
    return result
