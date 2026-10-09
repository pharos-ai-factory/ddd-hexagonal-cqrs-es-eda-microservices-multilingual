"""Every declaration owns exactly one incoming and one command consumer."""

from dataclasses import replace
from unittest.mock import Mock
import pytest
from operations.apps.runtime import complete_subscriptions, SubscriptionDefinition, SubscriptionWorker


def test_manifest_completeness() -> None:
    definition = SubscriptionDefinition("preparation.accept-order", "ordering.order-placed", "acceptOrder")
    worker = SubscriptionWorker("preparation", definition.consumer, Mock(), definition.event, False)
    pair = (worker, replace(worker, command=True))
    assert complete_subscriptions("preparation", [definition], pair) == pair
    for invalid in ((), pair[:1], pair + pair, (replace(worker, event="incorrect"), pair[1])):
        with pytest.raises(ValueError):
            complete_subscriptions("preparation", [definition], invalid)
    with pytest.raises(ValueError):
        complete_subscriptions("preparation", [definition, definition], pair)
