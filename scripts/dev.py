#!/usr/bin/env python3
"""Manage only this reference's isolated development Compose project."""
import argparse
import base64
import json
import http.client
import os
from pathlib import Path
import secrets
import socket
import subprocess
import time
import urllib.error
import urllib.request
from secret_files import atomic_private_write, compose_environment, read_configuration, resolve_files

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
    values = read_configuration(path) if path.exists() or path.is_symlink() else {}
    for key in ("POSTGRES_PASSWORD", "BROKER_PASSWORD", "API_KEY", "DELIVERY_KEY", "OPERATOR_PASSWORD",
                "SESSION_PASSWORD", "REALTIME_REDIS_PASSWORD", "CENTRIFUGO_API_KEY", "CONNECT_PROXY_SECRET",
                "STOREFRONT_API_KEY", "OPERATIONS_API_KEY", "ENGAGEMENT_API_KEY", "SESSION_REALTIME_KEY", "API_BROKER_PASSWORD"):
        if key + "_FILE" not in values:
            values.setdefault(key, secrets.token_hex(24))
    for owner in OWNERS:
        for suffix in ("_DB_PASSWORD", "_BROKER_PASSWORD", "_REALTIME_KEY"):
            key = owner.upper() + suffix
            if key + "_FILE" not in values:
                values.setdefault(key, secrets.token_hex(24))
    for key, port in PORTS.items():
        values.setdefault(key, str(free_port() if disposable else port))
    values.setdefault("COMPOSE_PROJECT_NAME", "cafe-reference-test-"+secrets.token_hex(4) if disposable else "cafe-reference")
    resolved = resolve_files(values, path.parent)
    atomic_private_write(path, "\n".join(f"{key}={value}" for key, value in values.items())+"\n")
    return resolved

def load(path):
    return resolve_files(read_configuration(path), path.parent)

def compose(path, *args, **kwargs):
    kwargs.setdefault("env", compose_environment(load(path), OWNERS))
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
                               {"configure": "^$", "write": rf"^ref\.{owner}\.delivery$|^cafe\.(events|replies)$", "read": queue})
                broker_request(values, "PUT", f"topic-permissions/reference/{user}",
                               {"exchange": "cafe.events", "write": rf"^(domain|integration)\.{owner}\..*$",
                                "read": rf"^(domain\.{owner}\..*|integration\..*)$"})
                broker_request(values, "PUT", f"topic-permissions/reference/{user}",
                               {"exchange": "cafe.replies", "write": rf"^reply\.{owner}$", "read": "^$"})
            broker_request(values, "PUT", "users/cafe_api",
                           {"password": values["API_BROKER_PASSWORD"], "tags": ""})
            broker_request(values, "PUT", "permissions/reference/cafe_api",
                           {"configure": "^$", "write": r"^cafe\.requests$", "read": r"^ref\.api\.[a-z]+\.replies$"})
            broker_request(values, "PUT", "topic-permissions/reference/cafe_api",
                           {"exchange": "cafe.requests", "write": r"^request\.(menu|ordering|preparation|collection|loyalty|communication)\.(command|query)$", "read": "^$"})
            return
        except (urllib.error.URLError, OSError, http.client.RemoteDisconnected):
            if time.monotonic() >= deadline:
                raise
            time.sleep(0.5)

def up(path):
    values = configure(path)
    compose(path, "up", "-d", "--wait", "postgres", "rabbitmq", "valkey", "realtime-history")
    # Upgrade existing demonstration databases using the administrator, before runtime startup.
    from migrate import migrate
    migrate(path)
    broker_users(values)
    compose(path, "up", "--build", "--abort-on-container-exit", "--exit-code-from", "topology", "topology")
    compose(path, "up", "-d", "--build", "storefront", "operations", "engagement", "delivery-simulator",
            "api", "centrifugo", "realtime-gateway", "ingress")
    wait_ready(values)
    print("Development services are ready. Credentials remain in", path)
    return values

def wait_ready(values, services=("STOREFRONT", "OPERATIONS", "ENGAGEMENT", "API", "WEB")):
    from readiness import report
    deadline = time.monotonic()+90
    while True:
        state = report(values, services)
        if state['ready']:
            return
        if time.monotonic() > deadline:
            unavailable = [name for name, status in state['checks'].items() if status != 'ready']
            raise RuntimeError("Development readiness timed out: "+", ".join(unavailable))
        time.sleep(0.25)


def test_environment(values):
    result = dict(os.environ, APP_ENV="development", CAFE_DISPOSABLE_PROJECT=values["COMPOSE_PROJECT_NAME"])
    for owner in OWNERS:
        prefix = owner.upper()
        result[prefix+"_DATABASE_URL"] = f'postgres://cafe_{owner}:{values[prefix+"_DB_PASSWORD"]}@127.0.0.1:{values["PG_PORT"]}/cafe_{owner}?sslmode=disable'
        result[prefix+"_BROKER_URL"] = f'amqp://cafe_{owner}:{values[prefix+"_BROKER_PASSWORD"]}@127.0.0.1:{values["AMQP_PORT"]}/reference'
    result["API_BROKER_URL"] = f'amqp://cafe_api:{values["API_BROKER_PASSWORD"]}@127.0.0.1:{values["AMQP_PORT"]}/reference'
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
