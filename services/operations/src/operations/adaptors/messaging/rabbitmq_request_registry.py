"""Dispatch RabbitMQ requests to owner handlers and arrange durable command replies."""

from collections.abc import Callable, Mapping
from threading import Event

import pika
from google.protobuf.message import Message
from pika.adapters.blocking_connection import BlockingChannel
from pika.spec import Basic

from operations.adaptors.diagnostics import failure
from operations.adaptors.generated.request_registry import Reply, Request
from operations.adaptors.messaging.rabbitmq_connection import broker_connection
from operations.adaptors.messaging.request_envelope import (
    new_reply,
    outcome_reply,
    parse_request,
    payload,
    validate,
)
from operations.adaptors.persistence.command_reply_outbox import ReplyIntent, current
from operations.foundation.application import Metadata, Outcome, QueryBusPort
from operations.adaptors.generated.request_names import CommandName, QueryName, COMMAND_TYPES, QUERY_TYPES


class RabbitMQRequestRegistry:
    """Maps owner RPC requests to typed application handlers and manages durable replies."""

    def __init__(self, owner: str) -> None:
        self.owner = owner
        self.handlers: dict[str, Callable[[Request, Message], Reply]] = {}

    def handle(self, request: Request) -> Reply:
        name, body = validate(request, self.owner)
        if name not in self.handlers:
            raise ValueError("Unknown owner request")
        try:
            reply = self.handlers[name](request, body)
        except Exception as error:
            failure(self.owner, "requests.handle", error, request.request_id)
            reply = new_reply(self.owner)
            reply.error.code = "temporarily_unavailable"
        reply.contract_version = 1
        reply.request_id = request.request_id
        reply.context = self.owner
        return reply

    def command[C: Mapping[str, object]](
        self,
        name: CommandName,
        parse: Callable[[Message], C],
        execute: Callable[[Metadata, C], Outcome],
    ) -> None:
        if not isinstance(name, COMMAND_TYPES.get(self.owner, ())):
            raise ValueError("Command identifier disagrees with owner")

        def handle(request: Request, body: Message) -> Reply:
            value = parse(body)
            wire = request.command.metadata
            metadata = Metadata(
                id=wire.command_id,
                target=wire.aggregate_id,
                name=self.owner + "." + body.DESCRIPTOR.name,
                correlation=wire.correlation_id,
                expected=wire.expected_version,
                input=value,
            )
            outcome = execute(metadata, value)
            reply = outcome_reply(self.owner, outcome)
            return reply

        if name in self.handlers:
            raise ValueError("Duplicate request handler: " + name)
        self.handlers[name] = handle

    def query[Q, R](
        self,
        name: QueryName,
        parse: Callable[[Message], Q],
        execute: Callable[[Q], R],
        reply: Callable[[Q, R], Reply],
    ) -> None:
        """Connect a generated query identifier to its owner parser, handler and reply mapper."""
        if not isinstance(name, QUERY_TYPES.get(self.owner, ())):
            raise ValueError("Query identifier disagrees with owner")
        if name in self.handlers:
            raise ValueError("Duplicate request handler: " + name)

        bus: QueryBusPort[Q, R] = HandlerQueryBus(execute)

        def handle(request: Request, body: Message) -> Reply:
            query = parse(body)
            return reply(query, bus.ask(query))

        self.handlers[name] = handle

    def run(self, url: str, stop: Event) -> None:
        from threading import Thread

        command = Thread(target=self._run_kind, args=(url, stop, "command"), daemon=True)
        command.start()
        try:
            self._run_kind(url, stop, "query")
        finally:
            stop.set()
            command.join()

    def _run_kind(self, url: str, stop: Event, kind: str) -> None:
        queue = "ref." + self.owner + (".commands" if kind == "command" else ".queries")
        while not stop.is_set():
            try:
                with broker_connection(url) as connection:
                    channel = connection.channel()
                    channel.confirm_delivery()
                    channel.basic_qos(prefetch_count=1)

                    def receive(
                        channel: BlockingChannel,
                        method: Basic.Deliver,
                        properties: pika.BasicProperties,
                        body: bytes,
                    ) -> None:
                        if method.delivery_tag is None:
                            raise ValueError("Missing delivery tag")
                        try:
                            if (
                                properties.content_type != "application/x-protobuf"
                                or properties.delivery_mode != 2
                                or properties.type != kind
                                or properties.app_id != "api"
                                or properties.message_id != properties.correlation_id
                                or method.routing_key != "request." + self.owner + "." + kind
                                or len(body) > 65536
                            ):
                                raise ValueError("Invalid request properties")
                            request = parse_request(self.owner, body)
                            if request.request_id != properties.message_id:
                                raise ValueError("Request identity mismatch")
                            validate(request, self.owner)
                            if payload(request)[0] != kind:
                                raise ValueError("Request kind disagrees with queue")
                        except Exception as error:
                            channel.basic_publish(
                                "ref." + self.owner + ".delivery",
                                queue + ".dead",
                                body,
                                properties=properties,
                                mandatory=True,
                            )
                            failure(self.owner, "requests.validate", error, dead=True)
                        else:

                            def encode(outcome: Outcome) -> bytes:
                                response = outcome_reply(self.owner, outcome)
                                response.contract_version = 1
                                response.context = self.owner
                                response.request_id = request.request_id
                                return response.SerializeToString(deterministic=True)

                            intent = ReplyIntent(request.request_id, encode)
                            token = current.set(intent if kind == "command" else None)
                            try:
                                reply = self.handle(request)
                            finally:
                                current.reset(token)
                            if intent.persisted and reply.HasField("outcome"):
                                channel.basic_ack(method.delivery_tag)
                                return
                            channel.basic_publish(
                                "cafe.replies",
                                "reply." + self.owner,
                                reply.SerializeToString(),
                                properties=pika.BasicProperties(
                                    content_type="application/x-protobuf",
                                    delivery_mode=2,
                                    type="reply",
                                    app_id=self.owner,
                                    message_id=request.request_id,
                                    correlation_id=request.request_id,
                                ),
                                mandatory=True,
                            )
                        # The owner transaction and confirmed reply precede ACK.
                        channel.basic_ack(method.delivery_tag)

                    channel.basic_consume(queue, receive, auto_ack=False)
                    while not stop.is_set():
                        connection.process_data_events(time_limit=1)
            except Exception as error:
                failure(self.owner, "requests.reconnect", error)
                stop.wait(1)


class HandlerQueryBus[Q, R]:
    """Route the application query after its owner parser has checked the wire payload."""

    def __init__(self, execute: Callable[[Q], R]) -> None:
        self._execute = execute

    def ask(self, query: Q) -> R:
        return self._execute(query)
