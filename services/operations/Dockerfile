FROM python:3.14-slim
COPY --from=ghcr.io/astral-sh/uv:0.12.19 /uv /usr/local/bin/uv
WORKDIR /app
COPY services/operations/pyproject.toml services/operations/uv.lock ./
COPY services/operations/src ./src
RUN uv sync --frozen --no-dev && useradd --uid 10001 --create-home reference
USER 10001:10001
CMD ["/app/.venv/bin/uvicorn", "operations.apps.service:create_app", "--factory", "--host", "0.0.0.0", "--port", "8080"]
