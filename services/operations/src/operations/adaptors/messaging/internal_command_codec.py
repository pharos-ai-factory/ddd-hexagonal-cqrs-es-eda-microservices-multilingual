"""Encode and validate owner-private command envelopes and receipt material."""

from operations.adaptors.generated.message_names import SubscriptionName, PrivateCommandName
import base64
import json
from collections.abc import Callable, Mapping
from typing import cast

from google.protobuf.json_format import MessageToDict, ParseDict
from google.protobuf.message import Message

from operations.foundation.application import Metadata
from operations.foundation.domain import identifier
from operations.foundation.identity import derived_id


class InternalCommandCodec[C: Mapping[str, object]]:
    """Validate typed owner commands independently from published event envelopes."""

    def __init__(
        self,
        owner: str,
        consumer: SubscriptionName,
        command: PrivateCommandName,
        envelope: type[Message],
        parse: Callable[[object], C],
    ) -> None:
        if not consumer.startswith(owner + "."):
            raise ValueError("Foreign command subscription")
        self.owner, self.consumer, self.command = owner, consumer, command
        self.envelope, self.parse = envelope, parse

    def validate(self, metadata: Metadata) -> None:
        for value in (metadata.id, metadata.target, metadata.correlation, metadata.source_id):
            identifier(value)
        if (
            metadata.consumer != self.consumer
            or metadata.id != derived_id(self.consumer, metadata.source_id)
            or len(metadata.source_hash) != 64
            or any(c not in "0123456789abcdef" for c in metadata.source_hash)
        ):
            raise ValueError("Invalid command identity")

    def encode(self, metadata: Metadata, command: C) -> bytes:
        self.validate(metadata)
        envelope = self.envelope()
        ParseDict(
            {
                "id": metadata.id,
                "consumer": metadata.consumer,
                "sourceEventId": metadata.source_id,
                "sourceHash": metadata.source_hash,
                "target": metadata.target,
                "correlationId": metadata.correlation,
                "receiptMaterial": base64.b64encode(json.dumps(dict(metadata.input)).encode()).decode(),
                self.command: dict(self.parse(command)),
            },
            envelope,
        )
        return envelope.SerializeToString(deterministic=True)

    def decode(self, body: bytes) -> tuple[Metadata, C]:
        envelope = self.envelope()
        envelope.ParseFromString(body)
        value = MessageToDict(envelope)
        if envelope.WhichOneof("command") != self.command:
            raise ValueError("Wrong command payload")
        material = json.loads(base64.b64decode(value["receiptMaterial"]))
        if not isinstance(material, dict):
            raise ValueError("Invalid receipt material")
        metadata = Metadata(
            id=identifier(value["id"]),
            target=identifier(value["target"]),
            name=self.consumer,
            correlation=identifier(value["correlationId"]),
            consumer=value["consumer"],
            source_id=identifier(value["sourceEventId"]),
            source_hash=value["sourceHash"],
            causation=identifier(value["sourceEventId"]),
            input=cast(dict[str, object], material),
        )
        self.validate(metadata)
        key = envelope.DESCRIPTOR.fields_by_name[self.command].json_name
        return metadata, self.parse(value[key])
