"""Exercise publication policy and codecs; real infrastructure proves atomicity."""

from collections.abc import Callable
from operations.adaptors import codec
from operations.adaptors.generated.cafe.v1.events_pb2 import Event
from operations.foundation.application import Metadata, Outcome, Publication


class PublicationProbe[P]:
    """Capture the published payload selected by a domain-event handler."""

    def __init__(self, publication: Callable[[P], Publication]) -> None:
        self.publication = publication
        self.publications: list[Publication] = []

    def publish(self, metadata: Metadata, payload: P) -> Outcome:
        self.publications.append(self.publication(payload))
        return Outcome(aggregateId=metadata.target, version=1, status="publication_queued")


def private_delivery(owner: str, kind: str, metadata: Metadata, publication: Publication) -> Event:
    _, body = codec.encode(owner, kind, metadata.target, 1, metadata, publication)
    event, _ = codec.decode(body)
    return event
