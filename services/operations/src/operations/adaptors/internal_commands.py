"""Owner-private Protobuf commands and atomic receiving receipts/outgoing intent."""
import base64
from collections.abc import Callable, Mapping
from typing import cast
import json
from google.protobuf.json_format import MessageToDict, ParseDict
from google.protobuf.message import Message
from operations.adaptors.postgres import PostgresContextDatabase
from operations.foundation.application import Metadata, Outcome
from operations.foundation.domain import identifier
from operations.foundation.identity import derived_id


class InternalCommandCodec[C: Mapping[str, object]]:
    """Validate typed owner commands independently from published event envelopes."""
    def __init__(self, owner: str, consumer: str, command: str, envelope: type[Message],
                 parse: Callable[[object], C]) -> None:
        if not consumer.startswith(owner+"."):
            raise ValueError("Foreign command subscription")
        self.owner, self.consumer, self.command = owner, consumer, command
        self.envelope, self.parse = envelope, parse

    def validate(self, metadata: Metadata) -> None:
        for value in (metadata.id, metadata.target, metadata.correlation, metadata.source_id):
            identifier(value)
        if (metadata.consumer != self.consumer or metadata.id != derived_id(self.consumer, metadata.source_id)
            or len(metadata.source_hash) != 64 or any(c not in "0123456789abcdef" for c in metadata.source_hash)):
            raise ValueError("Invalid command identity")

    def encode(self, metadata: Metadata, command: C) -> bytes:
        self.validate(metadata)
        envelope = self.envelope()
        ParseDict({"id": metadata.id, "consumer": metadata.consumer, "sourceEventId": metadata.source_id,
            "sourceHash": metadata.source_hash, "target": metadata.target, "correlationId": metadata.correlation,
            "receiptMaterial": base64.b64encode(json.dumps(dict(metadata.input)).encode()).decode(),
            self.command: dict(self.parse(command))}, envelope)
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
        metadata = Metadata(id=identifier(value["id"]), target=identifier(value["target"]), name=self.consumer,
            correlation=identifier(value["correlationId"]), consumer=value["consumer"],
            source_id=identifier(value["sourceEventId"]), source_hash=value["sourceHash"],
            causation=identifier(value["sourceEventId"]), input=cast(dict[str, object], material))
        self.validate(metadata)
        key = envelope.DESCRIPTOR.fields_by_name[self.command].json_name
        return metadata, self.parse(value[key])


class PostgresDurableCommandOutbox[C: Mapping[str, object]]:
    """Return acceptance only after the receiving receipt and exact command bytes commit."""
    def __init__(self, database: PostgresContextDatabase, codec: InternalCommandCodec[C]) -> None:
        if database.owner != codec.owner:
            raise ValueError("Command outbox crossed its owner")
        self.database, self.codec = database, codec

    def enqueue(self, metadata: Metadata, command: C) -> Outcome:
        body = self.codec.encode(metadata, command)
        with self.database.pool.connection() as connection:
            with connection.transaction():
                connection.execute("SELECT pg_advisory_xact_lock(hashtextextended(%s,0))",
                    ("handoff:"+metadata.consumer+":"+metadata.source_id,))
                previous = connection.execute("""SELECT id,fingerprint,target FROM cafe.internal_commands
                    WHERE consumer=%s AND event_id=%s""", (metadata.consumer, metadata.source_id)).fetchone()
                if previous:
                    if (str(previous["id"]) != metadata.id or previous["fingerprint"] != metadata.source_hash
                        or str(previous["target"]) != metadata.target):
                        raise ValueError("Conflicting event hand-off identity")
                else:
                    connection.execute("""INSERT INTO cafe.internal_commands(id,consumer,event_id,fingerprint,target,body)
                        VALUES(%s,%s,%s,%s,%s,%s)""", (metadata.id, metadata.consumer, metadata.source_id,
                            metadata.source_hash, metadata.target, body))
                    connection.execute("INSERT INTO cafe.internal_command_dispatches(event_id) VALUES(%s)", (metadata.id,))
        return Outcome(aggregateId=metadata.target, version=0, status="queued")
