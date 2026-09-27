#!/usr/bin/env python3
"""Manage only this reference's isolated development Compose project."""
import argparse
import base64
import json
import http.client
from hashlib import sha256
import os
from pathlib import Path
import secrets
import socket
import subprocess
import time
import urllib.error
import urllib.request

ROOT = Path(__file__).resolve().parents[1]
OWNERS = ("menu", "ordering", "preparation", "collection", "loyalty", "communication")
PORTS = {"PG_PORT": 25432, "AMQP_PORT": 25673, "BROKER_HTTP_PORT": 25674,
         "STOREFRONT_PORT": 28081, "OPERATIONS_PORT": 28082,
         "ENGAGEMENT_PORT": 28083, "DELIVERY_PORT": 28084, "API_PORT": 28080,
         "WEB_PORT": 28000, "REALTIME_ADMIN_PORT": 28090}

def free_port():
    with socket.socket() as sock:
        sock.bind(("127.0.0.1", 0))
        return sock.getsockname()[1]

def configure(path, disposable=False):
    values = load(path) if path.exists() else {}
    for key in ("POSTGRES_PASSWORD", "BROKER_PASSWORD", "API_KEY", "DELIVERY_KEY", "OPERATOR_PASSWORD",
                "SESSION_PASSWORD", "REALTIME_REDIS_PASSWORD", "CENTRIFUGO_API_KEY", "CONNECT_PROXY_SECRET",
                "STOREFRONT_API_KEY", "OPERATIONS_API_KEY", "ENGAGEMENT_API_KEY"):
        values.setdefault(key, secrets.token_hex(24))
    for owner in OWNERS:
        values.setdefault(owner.upper()+"_DB_PASSWORD", secrets.token_hex(24))
        values.setdefault(owner.upper()+"_BROKER_PASSWORD", secrets.token_hex(24))
    for key, port in PORTS.items():
        values.setdefault(key, str(free_port() if disposable else port))
    values.setdefault("COMPOSE_PROJECT_NAME", "cafe-reference-test-"+secrets.token_hex(4) if disposable else "cafe-reference")
    path.parent.mkdir(parents=True, exist_ok=True)
    fd = os.open(path, os.O_WRONLY | os.O_CREAT | os.O_TRUNC, 0o600)
    with os.fdopen(fd, "w") as stream:
        stream.write("\n".join(f"{key}={value}" for key, value in values.items())+"\n")
    return values

def load(path):
    return dict(line.split("=", 1) for line in path.read_text().splitlines() if line and not line.startswith("#"))

def compose(path, *args, **kwargs):
    return subprocess.run(["docker", "compose", "--env-file", str(path), "-f",
                           str(ROOT / "devops/compose.yaml"), *args], cwd=ROOT, check=True, **kwargs)

def broker_request(values, method, resource, body):
    url = f'http://127.0.0.1:{values["BROKER_HTTP_PORT"]}/api/{resource}'
    auth = base64.b64encode(("administrator:"+values["BROKER_PASSWORD"]).encode()).decode()
    request = urllib.request.Request(url, data=json.dumps(body).encode(), method=method,
                                    headers={"Authorization": "Basic "+auth, "Content-Type": "application/json"})
    with urllib.request.urlopen(request, timeout=10) as response:
        return response.read()

def broker_users(values):
    deadline = time.monotonic()+60
    while True:
        try:
            for owner in OWNERS:
                user = "cafe_"+owner
                broker_request(values, "PUT", f"users/{user}",
                               {"password": values[owner.upper()+"_BROKER_PASSWORD"], "tags": ""})
                queue = rf"^ref\.{owner}\..*"
                broker_request(values, "PUT", f"permissions/reference/{user}",
                               {"configure": queue+r"|^cafe\.events$", "write": queue+r"|^cafe\.events$", "read": queue})
                broker_request(values, "PUT", f"topic-permissions/reference/{user}",
                               {"exchange": "cafe.events", "write": rf"^(domain|integration)\.{owner}\..*$",
                                "read": rf"^(domain\.{owner}\..*|integration\..*)$"})
            return
        except (urllib.error.URLError, OSError, http.client.RemoteDisconnected):
            if time.monotonic() >= deadline:
                raise
            time.sleep(0.5)

def up(path):
    values = configure(path)
    compose(path, "up", "-d", "--wait", "postgres", "rabbitmq", "valkey")
    # Upgrade existing demonstration databases using the administrator, before runtime startup.
    for owner in OWNERS:
        checksum = sha256((ROOT/"contracts/persistence/0002_realtime.sql").read_bytes()).hexdigest()
        compose(path, "exec", "-T", "postgres", "psql", "-U", "postgres", "-d", "cafe_"+owner,
                "-v", "ON_ERROR_STOP=1", "-v", "role=cafe_"+owner, "-v", "checksum="+checksum,
                "-f", "/reference/0002_realtime.sql",
                stdout=subprocess.DEVNULL)
    broker_users(values)
    compose(path, "up", "--build", "--abort-on-container-exit", "--exit-code-from", "topology", "topology")
    compose(path, "up", "-d", "--build", "storefront", "operations", "engagement", "delivery-simulator",
            "api", "centrifugo", "web", "ingress")
    wait_ready(values)
    print("Development services are ready. Credentials remain in", path)
    return values

def wait_ready(values, services=("STOREFRONT", "OPERATIONS", "ENGAGEMENT", "API", "WEB")):
    deadline = time.monotonic()+90
    for service in services:
        while True:
            try:
                with urllib.request.urlopen(f'http://127.0.0.1:{values[service+"_PORT"]}/healthz', timeout=2) as response:
                    if response.status == 200:
                        break
            except (urllib.error.URLError, OSError, http.client.RemoteDisconnected):
                pass
            if time.monotonic()>deadline:
                raise RuntimeError("Development services did not become healthy")
            time.sleep(0.25)

def test_environment(values):
    result = dict(os.environ, APP_ENV="development", CAFE_DISPOSABLE_PROJECT=values["COMPOSE_PROJECT_NAME"])
    for owner in OWNERS:
        prefix = owner.upper()
        result[prefix+"_DATABASE_URL"] = f'postgres://cafe_{owner}:{values[prefix+"_DB_PASSWORD"]}@127.0.0.1:{values["PG_PORT"]}/cafe_{owner}?sslmode=disable'
        result[prefix+"_BROKER_URL"] = f'amqp://cafe_{owner}:{values[prefix+"_BROKER_PASSWORD"]}@127.0.0.1:{values["AMQP_PORT"]}/reference'
    result["BROKER_ADMIN_URL"] = f'amqp://administrator:{values["BROKER_PASSWORD"]}@127.0.0.1:{values["AMQP_PORT"]}/reference'
    result["DATABASE_ADMIN_URL"] = f'postgres://postgres:{values["POSTGRES_PASSWORD"]}@127.0.0.1:{values["PG_PORT"]}/cafe_ordering?sslmode=disable'
    return result

def main():
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("action", choices=("configure", "up", "down", "logs"))
    parser.add_argument("--env-file", type=Path, default=ROOT/".local/dev.env")
    args = parser.parse_args()
    if args.action == "configure":
        configure(args.env_file)
        print("Development configuration is ready:", args.env_file)
    elif args.action == "up":
        up(args.env_file)
    elif args.action == "down":
        compose(args.env_file, "down")
    else:
        compose(args.env_file, "logs", "--tail", "100")

if __name__ == "__main__":
    main()
