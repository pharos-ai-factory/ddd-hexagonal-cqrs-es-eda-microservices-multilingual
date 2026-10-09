"""Owner HTTP exposes operational endpoints; business requests use RabbitMQ."""
import asyncio
from collections.abc import Iterator
from contextlib import ExitStack
import json
from pathlib import Path
from unittest.mock import MagicMock, patch
import pytest
from fastapi import FastAPI
from starlette.routing import Route
from starlette.types import Message
from operations.apps import service
from operations.apps.composition import preparation as preparation_di, collection as collection_di

KEY = "test-operational-key-" * 2


@pytest.fixture
def app() -> Iterator[FastAPI]:
    with (
        ExitStack() as resources,
        patch.dict("os.environ", {"APP_ENV": "development"}),
        patch.object(service, "secret", return_value=KEY),
        patch("operations.adaptors.http.secret", return_value=KEY),
        patch.object(preparation_di, "PreparationContainer", return_value=MagicMock()),
        patch.object(collection_di, "CollectionContainer", return_value=MagicMock()),
        patch.object(preparation_di, "subscriptions", return_value=()),
        patch.object(collection_di, "subscriptions", return_value=()),
        patch.object(service, "diagnostics", return_value={"databases": {}}),
    ):
        yield service.configure_app(resources)


def response(app: FastAPI, method: str, path: str, key: str | None = KEY) -> tuple[int, dict[str, str]]:
    async def invoke() -> tuple[int, dict[str, str]]:
        messages: list[Message] = []
        received = False

        async def receive() -> Message:
            nonlocal received
            if not received:
                received = True
                return {"type": "http.request", "body": b"{}", "more_body": False}
            await asyncio.Event().wait()
            return {"type": "http.disconnect"}

        async def send(message: Message) -> None:
            messages.append(message)

        headers = [(b"authorization", ("Bearer "+key).encode())] if key else []
        await app({"type": "http", "asgi": {"version": "3.0", "spec_version": "2.4"},
                   "http_version": "1.1", "scheme": "http", "method": method, "path": path,
                   "raw_path": path.encode(), "query_string": b"", "headers": headers,
                   "server": ("test", 80), "client": ("test", 1)}, receive, send)
        start = next(message for message in messages if message["type"] == "http.response.start")
        return start["status"], {k.decode(): v.decode() for k, v in start["headers"]}
    return asyncio.run(invoke())


def test_operational_surface_and_authentication(app: FastAPI) -> None:
    assert {(route.path, method) for route in app.routes if isinstance(route, Route)
            for method in route.methods or ()} == {("/healthz", "GET"), ("/diagnostics", "GET")}
    assert response(app, "GET", "/healthz", None)[0] == 200
    assert response(app, "GET", "/diagnostics", None)[0] == 401
    status, headers = response(app, "GET", "/diagnostics")
    assert status == 200 and headers["cache-control"] == "no-store"


def test_public_business_paths_have_no_owner_http_entry_point(app: FastAPI) -> None:
    root = Path(__file__).resolve().parents[3]
    document = json.loads((root/"contracts/services/api/http_api/api.openapi.json").read_text())
    paths = [path.removeprefix("/api").replace("{id}", "11111111-1111-4111-8111-111111111111")
             for path in document["paths"] if path.startswith("/api/v1/")]
    assert paths
    for path in paths:
        for method in ("GET", "POST"):
            assert response(app, method, path)[0] == 404, (method, path)
