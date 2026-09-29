"""Python composition root: exactly Preparation and Collection."""
from contextlib import asynccontextmanager
from collections.abc import AsyncIterator, Callable
from functools import partial
import os
from threading import Event, Thread
from fastapi import FastAPI
from fastapi.responses import JSONResponse
from operations.adaptors import codec, inputs
from operations.adaptors.diagnostics import diagnostics
from operations.adaptors.delivery import Subscription, consume, realtime_relay, relay
from operations.adaptors.http import authenticate, mount_command, mount_paged_queries
from operations.adaptors.postgres import Commands, Database, Queries
from operations.contexts.preparation.application import AcceptOrder, CompletePreparation, StartPreparation
from operations.contexts.collection.application import CollectOrder, OpenPickup
from operations.contexts.preparation.domain import PreparationTicket, TicketSnapshot
from operations.contexts.collection.domain import Pickup, PickupSnapshot
from operations.contracts.events import DrinksReady, OrderPlaced
from operations.foundation.identity import derived_id
from operations.foundation.secrets import secret


def create_app() -> FastAPI:
    if os.environ.get("APP_ENV") not in ("local", "development"):
        raise RuntimeError("This reference only runs in local or development environments")
    # Validate credentials synchronously before any worker can start.
    for owner in ("preparation", "collection"):
        secret(owner.upper()+"_BROKER_URL")
        secret(owner.upper()+"_REALTIME_KEY")
    databases = {owner: Database(owner, secret(owner.upper()+"_DATABASE_URL"))
                 for owner in ("preparation", "collection")}
    def ticket_state(value: object) -> TicketSnapshot:
        return PreparationTicket.restore(value).snapshot()
    def pickup_state(value: object) -> PickupSnapshot:
        return Pickup.restore(value).snapshot()
    tickets = Commands(databases["preparation"], "ticket", ticket_state)
    pickups = Commands(databases["collection"], "pickup", pickup_state)
    stop = Event()
    workers: list[Thread] = []

    @asynccontextmanager
    async def lifespan(app: FastAPI) -> AsyncIterator[None]:
        def start(target: Callable[[], None]) -> None:
            thread = Thread(target=target, daemon=True)
            workers.append(thread)
            thread.start()

        for owner, database in databases.items():
            start(partial(relay, database, secret(owner.upper()+"_BROKER_URL"), stop))
            start(partial(realtime_relay, database, stop))
        accept = Subscription[OrderPlaced]("preparation", "preparation.accept-order", "ordering.order-placed",
            codec.order_placed, lambda event: derived_id("ticket", event["orderId"]), AcceptOrder(tickets).handle)
        collect = Subscription[DrinksReady]("collection", "collection.open-pickup", "preparation.drinks-ready",
            codec.drinks_ready, lambda event: derived_id("pickup", event["orderId"]), OpenPickup(pickups, derived_id).handle)
        start(partial(consume, secret("PREPARATION_BROKER_URL"), accept, stop))
        start(partial(consume, secret("COLLECTION_BROKER_URL"), collect, stop))
        yield
        stop.set()
        for worker in workers:
            worker.join(timeout=6)
        for database in databases.values():
            database.pool.close()

    app = FastAPI(lifespan=lifespan, docs_url=None, redoc_url=None, openapi_url=None)
    authenticate(app)
    app.get("/healthz")(lambda: {"status": "ok", "language": "python"})
    app.get("/diagnostics")(lambda: JSONResponse(diagnostics(databases), headers={"Cache-Control": "no-store"}))
    mount_paged_queries(app, "/v1/preparation/tickets", Queries(databases["preparation"], "ticket", ticket_state))
    mount_paged_queries(app, "/v1/collection/pickups", Queries(databases["collection"], "pickup", pickup_state))
    mount_command(app, "/v1/preparation/tickets/{identity}/start", "preparation.StartPreparation",
                  StartPreparation(tickets).execute, inputs.start_preparation)
    mount_command(app, "/v1/preparation/tickets/{identity}/complete", "preparation.CompletePreparation",
                  CompletePreparation(tickets).execute, inputs.complete_preparation)
    mount_command(app, "/v1/collection/pickups/{identity}/collect", "collection.CollectOrder",
                  CollectOrder(pickups).execute, inputs.collect_order)
    return app
