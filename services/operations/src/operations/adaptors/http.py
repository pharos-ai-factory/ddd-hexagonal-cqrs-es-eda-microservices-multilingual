"""Authentication and error responses for the operational HTTP surface."""
from operations.foundation.secrets import secret
import hmac
from collections.abc import Awaitable, Callable
from fastapi import FastAPI, Request
from fastapi.responses import JSONResponse, Response


def authenticate(app: FastAPI) -> None:
    key = secret("API_KEY")
    if len(key) < 32:
        raise RuntimeError("A development service key is required")

    @app.exception_handler(Exception)
    async def unavailable(request: Request, error: Exception) -> JSONResponse:
        return JSONResponse({"code": "temporarily_unavailable"}, 503)

    @app.middleware("http")
    async def auth(request: Request, call_next: Callable[[Request], Awaitable[Response]]) -> Response:
        if request.url.path != "/healthz" and not hmac.compare_digest(
            request.headers.get("authorization", ""), "Bearer "+key
        ):
            return JSONResponse({"code": "unauthorised"}, 401)
        return await call_next(request)
