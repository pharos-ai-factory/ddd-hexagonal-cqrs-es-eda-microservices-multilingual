"""Resolve all application providers without opening infrastructure."""

from collections.abc import Iterator
from unittest.mock import Mock
from dependency_injector import providers
from operations.adaptors.persistence.postgres_database import PostgresContextDatabase
from operations.apps.composition.preparation import (
    PreparationContainer,
    subscriptions as preparation_subscriptions,
    requests as preparation_requests,
)
from operations.apps.composition.collection import (
    CollectionContainer,
    subscriptions as collection_subscriptions,
    requests as collection_requests,
)


def database_resource(owner: str, closed: list[str]) -> Iterator[PostgresContextDatabase]:
    database = Mock(spec=PostgresContextDatabase)
    database.owner = owner
    try:
        yield database
    finally:
        closed.append(owner)


def test_context_graphs_and_resources_are_independent() -> None:
    preparation, collection = PreparationContainer(), CollectionContainer()
    closed: list[str] = []
    for container, owner in ((preparation, "preparation"), (collection, "collection")):
        container.database.override(providers.Resource(database_resource, owner, closed))
        container.init_resources()
        for name, provider in container.providers.items():
            if name != "config":
                provider()
        assert container.database().owner == owner
        assert container.database() is container.database()
    assert len(preparation_subscriptions(preparation)) == 4
    assert len(collection_subscriptions(collection)) == 4
    assert set(preparation_requests(preparation).handlers) == {
        "startPreparation",
        "completePreparation",
        "getTicket",
        "listTickets",
    }
    assert set(collection_requests(collection).handlers) == {"collectOrder", "getPickup", "listPickups"}
    assert closed == []
    preparation.shutdown_resources()
    assert closed == ["preparation"]
    collection.shutdown_resources()
    assert closed == ["preparation", "collection"]
