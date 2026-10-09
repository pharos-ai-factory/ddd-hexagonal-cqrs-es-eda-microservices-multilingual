"""Common subscription wiring; owner declarations supply only their typed application policy."""
from collections.abc import Callable, Mapping, Sequence
from dataclasses import dataclass
from functools import partial
from threading import Event
from operations.adaptors.delivery import EventSubscription, consume
from operations.adaptors.internal_commands import InternalCommandCodec
from operations.adaptors.generated.cafe.v1.events_pb2 import Event as WireEvent
from operations.foundation.application import Metadata, Outcome


@dataclass(frozen=True)
class SubscriptionWorker:
    """Resolved owner worker that the service lifecycle starts and joins."""
    owner: str
    consumer: str
    run: Callable[[str, Event], None]


@dataclass(frozen=True)
class SubscriptionDefinition:
    """Owner manifest declaration used by both topology generation and runtime binding."""
    consumer: str
    event: str
    command: str


def command_subscription[E, C: Mapping[str, object]](
    codec: InternalCommandCodec[C], definitions: Sequence[SubscriptionDefinition], parse: Callable[[WireEvent], E], target: Callable[[E], str],
    on_event: Callable[[Metadata, E], Outcome], execute: Callable[[Metadata, C], Outcome],
) -> tuple[SubscriptionWorker, SubscriptionWorker]:
    definition = next((item for item in definitions if item.consumer == codec.consumer), None)
    command_name = codec.envelope.DESCRIPTOR.fields_by_name[codec.command].json_name
    if definition is None or definition.command != command_name:
        raise ValueError("Subscription manifest disagrees with command binding")
    event = definition.event
    def unused_parse(value: WireEvent) -> C:
        raise AssertionError("Command consumer uses its private command decoder")
    incoming = EventSubscription(codec.owner, codec.consumer, event, parse, target, on_event)
    command = EventSubscription(codec.owner, codec.consumer, event, unused_parse, lambda _: "", execute, codec.decode)
    def run[P](subscription: EventSubscription[P], url: str, stop: Event) -> None:
        consume(url, subscription, stop)
    return (SubscriptionWorker(codec.owner, codec.consumer, partial(run, incoming)),
            SubscriptionWorker(codec.owner, codec.consumer, partial(run, command)))
