"""Resolve all application providers without opening infrastructure."""
from collections.abc import Iterator
from unittest.mock import Mock
from dependency_injector import providers
from operations.adaptors.postgres import PostgresContextDatabase
from operations.apps.composition.preparation import PreparationContainer, subscriptions as preparation_subscriptions
from operations.apps.composition.collection import CollectionContainer, subscriptions as collection_subscriptions


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
    assert len(preparation_subscriptions(preparation)) == 2
    assert len(collection_subscriptions(collection)) == 2
    assert closed == []
    preparation.shutdown_resources()
    assert closed == ["preparation"]
    collection.shutdown_resources()
    assert closed == ["preparation", "collection"]
