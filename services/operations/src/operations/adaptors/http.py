from operations.foundation.secrets import secret
import hmac
from collections.abc import Awaitable, Callable, Mapping
from fastapi import FastAPI, Request
from fastapi.responses import JSONResponse, Response
from starlette.concurrency import run_in_threadpool
from operations.foundation.application import Loaded, Metadata, Outcome, QueryPort
from operations.foundation.pagination import Page, PageRequest, PagedQueryPort
from operations.adaptors.pagination import PageResponse, page_request, page_response
from operations.foundation.domain import Rejection, identifier


def mount_paged_queries[S](app: FastAPI, path: str, queries: PagedQueryPort[S]) -> None:
    mount_queries(app, path, queries, queries.page)


def mount_queries[S](app: FastAPI, path: str, queries: QueryPort[S],
                     page: Callable[[PageRequest], Page[S]] | None = None) -> None:
    def listing(request: Request) -> list[Loaded[S]] | PageResponse[S]:
        pagination = page_request(request.query_params, path)
        if pagination:
            if page is None:
                raise Rejection("invalid_pagination", "This query does not support pagination")
            return page_response(page(pagination), path)
        return queries.list()

    def get(identity: str) -> Loaded[S] | JSONResponse:
        identifier(identity)
        result = queries.get(identity)
        return result if result else JSONResponse({"code": "not_found"}, 404)

    app.add_api_route(path, listing, methods=["GET"], response_model=None)
    app.add_api_route(path+"/{identity}", get, methods=["GET"], response_model=None)


def mount_command[C: Mapping[str, object]](
    app: FastAPI, path: str, name: str, handler: Callable[[Metadata, C], Outcome],
    parse: Callable[[object], C],
) -> None:
    async def execute(identity: str, request: Request) -> JSONResponse:
        try:
            raw = await request.body()
            if len(raw) > 65536:
                return JSONResponse({"code": "invalid_request"}, 400)
            body = parse(await request.json())
            key = identifier(request.headers.get("idempotency-key", ""))
            correlation = identifier(request.headers.get("x-correlation-id", key))
            identity = identifier(identity)
            try:
                version = int(request.headers["if-match"].strip('"'))
                if version < 0:
                    raise ValueError()
            except (KeyError, ValueError):
                return JSONResponse({"code": "expected_version_required"}, 428)
            metadata = Metadata(key, identity, name, correlation, version, body)
        except Rejection as error:
            return JSONResponse(error.outcome(), 400)
        except (ValueError, TypeError):
            return JSONResponse({"code": "invalid_request"}, 400)
        except Exception:
            return JSONResponse({"code": "temporarily_unavailable"}, 503)
        try:
            # The blocking owner transaction runs in the HTTP worker pool.
            outcome = await run_in_threadpool(handler, metadata, body)
            rejection = outcome.get("rejection")
            code = rejection["code"] if rejection else None
            status = 409 if code in ("version_conflict", "idempotency_conflict") else (
                404 if code == "not_found" else 422 if code else 200)
            return JSONResponse(outcome, status)
        except Exception:
            return JSONResponse({"code": "temporarily_unavailable"}, 503)
    app.add_api_route(path, execute, methods=["POST"])


def authenticate(app: FastAPI) -> None:
    key = secret("API_KEY")
    if len(key) < 32:
        raise RuntimeError("A development service key is required")

    @app.exception_handler(Rejection)
    async def rejected_query(request: Request, error: Rejection) -> JSONResponse:
        return JSONResponse(error.outcome(), 400)

    @app.exception_handler(Exception)
    async def unavailable_query(request: Request, error: Exception) -> JSONResponse:
        return JSONResponse({"code": "temporarily_unavailable"}, 503)

    @app.middleware("http")
    async def auth(request: Request, call_next: Callable[[Request], Awaitable[Response]]) -> Response:
        if request.url.path != "/healthz" and not hmac.compare_digest(
            request.headers.get("authorization", ""), "Bearer "+key
        ):
            return JSONResponse({"code": "unauthorised"}, 401)
        return await call_next(request)
