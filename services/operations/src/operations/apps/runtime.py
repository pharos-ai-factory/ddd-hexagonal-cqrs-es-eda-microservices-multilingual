"""Common subscription wiring; owner declarations supply only their typed application policy."""

from operations.adaptors.generated.message_names import SubscriptionName, IntegrationEventName
from collections.abc import Callable, Mapping, Sequence
from dataclasses import dataclass
from functools import partial
from threading import Event
from operations.adaptors.messaging.incoming_event import IncomingEvent
from operations.adaptors.messaging.rabbitmq_consumer import EventSubscription, consume
from operations.adaptors.messaging.internal_command_codec import InternalCommandCodec
from operations.adaptors.generated.cafe.v1.events_pb2 import Event as WireEvent
from operations.foundation.application import Metadata, Outcome


@dataclass(frozen=True)
class SubscriptionWorker:
    """Resolved owner worker that the service lifecycle starts and joins."""

    owner: str
    consumer: str
    run: Callable[[str, Event], None]
    event: str
    command: bool


@dataclass(frozen=True)
class SubscriptionDefinition:
    """Owner manifest declaration used by both topology generation and runtime binding."""

    consumer: str
    event: str
    command: str = ""
    publication: str = ""


def command_subscription[E, C: Mapping[str, object]](
    codec: InternalCommandCodec[C],
    definitions: Sequence[SubscriptionDefinition],
    incoming_event: IncomingEvent[E],
    target: Callable[[E], str],
    on_event: Callable[[Metadata, E], Outcome],
    execute: Callable[[Metadata, C], Outcome],
) -> tuple[SubscriptionWorker, SubscriptionWorker]:
    definition = next((item for item in definitions if item.consumer == codec.consumer), None)
    command_name = codec.envelope.DESCRIPTOR.fields_by_name[codec.command].json_name
    if (
        definition is None
        or definition.publication
        or definition.command != command_name
        or definition.event != incoming_event.name
    ):
        raise ValueError("Subscription manifest disagrees with command binding")
    event = definition.event

    def unused_parse(value: WireEvent) -> C:
        raise AssertionError("Command consumer uses its private command decoder")

    incoming = EventSubscription(codec.owner, codec.consumer, event, incoming_event.parse, target, on_event)
    command = EventSubscription(
        codec.owner, codec.consumer, event, unused_parse, lambda _: "", execute, codec.decode
    )

    def run[P](subscription: EventSubscription[P], url: str, stop: Event) -> None:
        consume(url, subscription, stop)

    return (
        SubscriptionWorker(codec.owner, codec.consumer, partial(run, incoming), event, False),
        SubscriptionWorker(codec.owner, codec.consumer, partial(run, command), event, True),
    )


def complete_subscriptions(
    owner: str, definitions: Sequence[SubscriptionDefinition], workers: tuple[SubscriptionWorker, ...]
) -> tuple[SubscriptionWorker, ...]:
    """Validate the entire context manifest, including missing and repeated registrations."""
    expected = {item.consumer for item in definitions}
    if len(expected) != len(definitions):
        raise ValueError("Duplicate subscription declaration")
    for definition in definitions:
        if bool(definition.command) == bool(definition.publication):
            raise ValueError("Subscription requires exactly one reaction mode")
        actual = [worker for worker in workers if worker.consumer == definition.consumer]
        if (
            len(actual) != (2 if definition.command else 1)
            or sum(worker.command for worker in actual) != (1 if definition.command else 0)
            or any(worker.owner != owner or worker.event != definition.event for worker in actual)
        ):
            raise ValueError("Incomplete subscription: " + definition.consumer)
    if any(worker.consumer not in expected for worker in workers):
        raise ValueError("Undeclared subscription")
    return workers


def publication_subscription[E](
    owner: str,
    consumer: SubscriptionName,
    definitions: Sequence[SubscriptionDefinition],
    incoming_event: IncomingEvent[E],
    publication: IntegrationEventName,
    on_event: Callable[[Metadata, E], Outcome],
) -> tuple[SubscriptionWorker]:
    """Bind one private subscription whose handler atomically stages an integration event."""
    definition = next((item for item in definitions if item.consumer == consumer), None)
    if (
        definition is None
        or definition.command
        or definition.event != incoming_event.name
        or definition.publication != publication
        or not consumer.startswith(owner + ".")
    ):
        raise ValueError("Subscription manifest disagrees with publication binding")
    incoming = EventSubscription(
        owner, consumer, incoming_event.name, incoming_event.parse, lambda _: "", on_event, source_target=True
    )
    return (
        SubscriptionWorker(
            owner, consumer, lambda url, stop: consume(url, incoming, stop), incoming_event.name, False
        ),
    )
