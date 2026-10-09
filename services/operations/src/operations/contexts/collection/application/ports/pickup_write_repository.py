"""Authoritative aggregate persistence inside the collection transaction boundary."""
from operations.foundation.write_repository import WriteRepository
from operations.contexts.collection.domain.pickup import Pickup

type PickupWriteRepository = WriteRepository[Pickup]
