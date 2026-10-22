"""Register plain command handlers with the aggregate's shared execution policy."""

from collections.abc import Callable, Mapping, Sequence

from google.protobuf.message import Message
from operations.adaptors.generated.request_names import CommandName

from operations.adaptors.command_execution import AggregateTransaction, CommandExecutor, FeatureCommandHandler
from operations.adaptors.messaging.incoming_event import IncomingEvent
from operations.adaptors.messaging.internal_command_codec import InternalCommandCodec
from operations.adaptors.messaging.rabbitmq_request_registry import RabbitMQRequestRegistry
from operations.apps.runtime import SubscriptionDefinition, SubscriptionWorker, command_subscription
from operations.foundation.application import Metadata, Outcome, RealtimeBusPort, RealtimeMessage
from operations.foundation.write_repository import WriteRepository


class CommandRegistration[A]:
    """Supply transaction-bound repositories to handlers registered for one aggregate."""

    def __init__(self, transaction: AggregateTransaction[A]) -> None:
        self._transaction = transaction

    def request[C: Mapping[str, object]](
        self,
        registry: RabbitMQRequestRegistry,
        name: CommandName,
        parse: Callable[[Message], C],
        handler: Callable[[WriteRepository[A], RealtimeBusPort[RealtimeMessage]], FeatureCommandHandler[C]],
    ) -> None:
        """Connect a published command and its parser directly to its feature handler."""
        registry.command(name, parse, CommandExecutor(self._transaction, handler).execute)

    def subscription[E, C: Mapping[str, object]](
        self,
        codec: InternalCommandCodec[C],
        definitions: Sequence[SubscriptionDefinition],
        incoming_event: IncomingEvent[E],
        target: Callable[[E], str],
        on_event: Callable[[Metadata, E], Outcome],
        handler: Callable[[WriteRepository[A], RealtimeBusPort[RealtimeMessage]], FeatureCommandHandler[C]],
    ) -> tuple[SubscriptionWorker, SubscriptionWorker]:
        """Connect a private command codec to its handler and durable event reaction."""
        return command_subscription(
            codec,
            definitions,
            incoming_event,
            target,
            on_event,
            CommandExecutor(self._transaction, handler).execute,
        )
