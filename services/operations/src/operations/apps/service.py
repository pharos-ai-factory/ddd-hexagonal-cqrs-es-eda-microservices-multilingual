"""Python composition root: exactly Preparation and Collection."""
from contextlib import asynccontextmanager, ExitStack
from asyncio import to_thread
from collections.abc import AsyncIterator, Callable
from functools import partial
import os
from threading import Event, Thread
from fastapi import FastAPI
from fastapi.responses import JSONResponse
from operations.contexts.preparation.adaptors.messaging import accept_order_codec
from operations.contexts.collection.adaptors.messaging import open_pickup_codec
from operations.adaptors.diagnostics import diagnostics
from operations.adaptors.delivery import realtime_relay, relay
from operations.adaptors.http import authenticate
from operations.adaptors.requests import RabbitMQRequestRegistry
from operations.adaptors.replies import relay as reply_relay
from operations.contexts.preparation.adaptors.messaging import requests as preparation_wire
from operations.contexts.collection.adaptors.messaging import requests as collection_wire
from operations.foundation.secrets import secret
from operations.apps.composition import preparation as preparation_di, collection as collection_di


def create_app() -> FastAPI:
    resources = ExitStack()
    try:
        return configure_app(resources)
    except BaseException:
        resources.close()
        raise


def configure_app(resources: ExitStack) -> FastAPI:
    if os.environ.get("APP_ENV") not in ("local", "development"):
        raise RuntimeError("This reference only runs in local or development environments")
    # Validate credentials synchronously before any worker can start.
    for owner in ("preparation", "collection"):
        secret(owner.upper()+"_BROKER_URL")
        secret(owner.upper()+"_REALTIME_KEY")
    preparation = preparation_di.PreparationContainer()
    preparation.config.database_url.from_value(secret("PREPARATION_DATABASE_URL"))
    collection = collection_di.CollectionContainer()
    collection.config.database_url.from_value(secret("COLLECTION_DATABASE_URL"))
    resources.callback(preparation.shutdown_resources)
    preparation.init_resources()
    resources.callback(collection.shutdown_resources)
    collection.init_resources()
    databases = {"preparation": preparation.database(), "collection": collection.database()}
    subscriptions = (*preparation_di.subscriptions(preparation), *collection_di.subscriptions(collection))
    preparation_requests = RabbitMQRequestRegistry("preparation")
    preparation_requests.command("startPreparation", preparation_wire.start, preparation.start().execute)
    preparation_requests.command("completePreparation", preparation_wire.complete, preparation.complete().execute)
    preparation_requests.queries("ticket", "tickets", preparation.queries(), preparation_wire.item, preparation_wire.listing)
    collection_requests = RabbitMQRequestRegistry("collection")
    collection_requests.command("collectOrder", collection_wire.collect, collection.collect().execute)
    collection_requests.queries("pickup", "pickups", collection.queries(), collection_wire.item, collection_wire.listing)
    stop = Event()
    workers: list[Thread] = []

    @asynccontextmanager
    async def lifespan(app: FastAPI) -> AsyncIterator[None]:
        def start(target: Callable[[], None]) -> None:
            thread = Thread(target=target, daemon=True)
            workers.append(thread)
            thread.start()

        try:
            start(partial(preparation_requests.run, secret("PREPARATION_BROKER_URL"), stop))
            start(partial(collection_requests.run, secret("COLLECTION_BROKER_URL"), stop))
            for owner, database in databases.items():
                start(partial(reply_relay, database, secret(owner.upper()+"_BROKER_URL"), stop))
                start(partial(relay, database, secret(owner.upper()+"_BROKER_URL"), stop))
                start(partial(realtime_relay, database, stop))
            start(partial(relay, databases["preparation"], secret("PREPARATION_BROKER_URL"), stop, accept_order_codec.command_header))
            start(partial(relay, databases["collection"], secret("COLLECTION_BROKER_URL"), stop, open_pickup_codec.command_header))
            for subscription in subscriptions:
                if subscription.consumer not in os.environ.get("PAUSED_CONSUMERS", "").split(","):
                    start(partial(subscription.run, secret(subscription.owner.upper()+"_BROKER_URL"), stop))
            yield
        finally:
            stop.set()
            def close() -> None:
                for worker in workers:
                    if worker.ident is not None:
                        worker.join()
                resources.close()
            await to_thread(close)

    app = FastAPI(lifespan=lifespan, docs_url=None, redoc_url=None, openapi_url=None)
    authenticate(app)
    app.get("/healthz")(lambda: {"status": "ok", "language": "python"})
    app.get("/diagnostics")(lambda: JSONResponse(diagnostics(databases), headers={"Cache-Control": "no-store"}))
    return app
