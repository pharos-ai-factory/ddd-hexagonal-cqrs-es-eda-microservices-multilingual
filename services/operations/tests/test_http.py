import asyncio
from collections.abc import Callable
from typing import Never
import pytest
from fastapi import FastAPI
from starlette.requests import Request
from starlette.responses import JSONResponse
from starlette.routing import Route
from starlette.types import Message
from operations.adaptors.http import mount_command
from operations.adaptors.inputs import start_preparation
from operations.contexts.preparation.application import StartPreparationCommand
from operations.foundation.application import Metadata, Outcome

ID = "00000000-0000-4000-8000-000000000001"


def command_response(handler: Callable[[Metadata, StartPreparationCommand], Outcome], body: bytes = b"{}") -> JSONResponse:
    app = FastAPI()
    mount_command(app, "/commands/{identity}", "test.command", handler, start_preparation)

    async def receive() -> Message:
        return {"type": "http.request", "body": body, "more_body": False}

    request = Request({"type": "http", "method": "POST", "headers": [
        (b"idempotency-key", ID.encode()), (b"if-match", b"1")]}, receive)
    route = app.routes[-1]
    assert isinstance(route, Route)
    result: object = asyncio.run(route.endpoint(ID, request))
    assert isinstance(result, JSONResponse)
    return result


@pytest.mark.parametrize("error", [ValueError("Corrupt persisted state"), TypeError("Invalid stored shape")])
def test_handler_failures_remain_retryable_http_failures(error: Exception) -> None:
    def fail(_: Metadata, __: StartPreparationCommand) -> Never:
        raise error
    response = command_response(fail)
    assert response.status_code == 503
    assert b"temporarily_unavailable" in response.body


def test_malformed_json_does_not_reach_the_handler() -> None:
    def unexpected(_: Metadata, __: StartPreparationCommand) -> Never:
        pytest.fail("Malformed HTTP input reached the handler")
    assert command_response(unexpected, b"{").status_code == 400
