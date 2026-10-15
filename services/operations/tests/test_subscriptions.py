"""Publication declarations own one consumer; aggregate reactions also own a command consumer."""

from operations.adaptors.generated.message_names import (
    OrderingIntegrationEvent,
    PreparationDomainEvent,
    PreparationIntegrationEvent,
    PreparationSubscription,
)

from dataclasses import replace
from unittest.mock import Mock
import pytest
from operations.apps.runtime import complete_subscriptions, SubscriptionDefinition, SubscriptionWorker


def test_manifest_completeness() -> None:
    definition = SubscriptionDefinition(
        PreparationSubscription.ACCEPT_ORDER, OrderingIntegrationEvent.ORDER_PLACED, "acceptOrder"
    )
    worker = SubscriptionWorker("preparation", definition.consumer, Mock(), definition.event, False)
    pair = (worker, replace(worker, command=True))
    assert complete_subscriptions("preparation", [definition], pair) == pair
    for invalid in ((), pair[:1], pair + pair, (replace(worker, event="incorrect"), pair[1])):
        with pytest.raises(ValueError):
            complete_subscriptions("preparation", [definition], invalid)
    with pytest.raises(ValueError):
        complete_subscriptions("preparation", [definition, definition], pair)


def test_publication_subscriptions_require_one_worker_and_one_reaction_mode() -> None:
    definition = SubscriptionDefinition(
        PreparationSubscription.PUBLISH_DRINKS_READY,
        PreparationDomainEvent.DRINKS_READY,
        publication=PreparationIntegrationEvent.DRINKS_READY,
    )
    worker = SubscriptionWorker(
        "preparation", definition.consumer, lambda _url, _stop: None, definition.event, False
    )
    assert complete_subscriptions("preparation", [definition], (worker,)) == (worker,)
    for workers in ((), (worker, worker)):
        with pytest.raises(ValueError):
            complete_subscriptions("preparation", [definition], workers)
    ambiguous = SubscriptionDefinition(
        definition.consumer, definition.event, "acceptOrder", definition.publication
    )
    with pytest.raises(ValueError, match="one reaction"):
        complete_subscriptions("preparation", [ambiguous], (worker,))


def test_realtime_subscriptions_require_one_explicit_output() -> None:
    definition = SubscriptionDefinition(
        PreparationSubscription.REALTIME_DRINKS_READY, PreparationDomainEvent.DRINKS_READY, realtime="ticket"
    )
    worker = SubscriptionWorker("preparation", definition.consumer, Mock(), definition.event, False)
    assert complete_subscriptions("preparation", [definition], (worker,)) == (worker,)
    for invalid in (
        replace(definition, publication=PreparationIntegrationEvent.DRINKS_READY),
        replace(definition, command="acceptOrder"),
        replace(definition, realtime=""),
    ):
        with pytest.raises(ValueError, match="one reaction"):
            complete_subscriptions("preparation", [invalid], (worker,))
    with pytest.raises(ValueError, match="Incomplete"):
        complete_subscriptions("preparation", [definition], ())
