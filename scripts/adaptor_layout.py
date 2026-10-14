"""Keep service and context boundaries discoverable by responsibility (ADR 0021)."""

import re

ROOTS = {
    "services/operations/src/operations/adaptors/": {
        "persistence",
        "messaging",
        "realtime",
        "http",
        "generated",
    },
    "services/engagement/src/adaptors/": {"persistence", "messaging", "realtime", "http", "generated"},
    "services/storefront/adaptors/": {"persistence", "messaging", "realtime", "http", "command"},
    "services/api/adaptors/": {"config", "http", "messaging", "openapi", "realtime", "sessions"},
    "services/web/src/adaptors/": {"http", "realtime", "generated"},
}
ROOT_MODULES = {
    "__init__.py",
    "command_execution.py",
    "diagnostics.py",
    "command-execution.ts",
    "command-execution.test.ts",
    "diagnostics.ts",
    "diagnostics.test.ts",
}
CONTEXT = re.compile(
    r"services/(?:storefront|operations/src/operations|engagement/src)/contexts/[^/]+/adaptors/"
)


def layout_violation(path):
    if not path.endswith((".py", ".ts", ".tsx", ".go")):
        return None
    if match := CONTEXT.match(path):
        relative = path[match.end() :]
        if relative != "__init__.py" and relative.split("/")[0] not in {
            "persistence",
            "messaging",
            "delivery",
        }:
            return "context adaptor must use its persistence, messaging or outbound delivery package"
    for prefix, packages in ROOTS.items():
        if path.startswith(prefix):
            relative = path.removeprefix(prefix)
            if relative not in ROOT_MODULES and relative.split("/")[0] not in packages:
                return "shared adaptor must use its documented responsibility package"
    return None


def import_violation(path, target):
    if path.endswith((".integration.ts", ".test.ts", ".test-support.ts")):
        return None
    for prefix in ROOTS:
        if not path.startswith(prefix):
            continue
        role = path.removeprefix(prefix).split("/")[0]
        if (
            re.search(r"(?:^|/)contexts/[^/]+/(?:application|adaptors|domain)(?:/|$)", target)
            and "/generated/" not in "/" + target
        ):
            return "shared adaptor cannot import context implementation; use generated registration"
        if role == "persistence" and target.startswith(
            (
                "pika",
                "urllib",
                "fastapi",
                "httpx",
                "requests",
                "amqplib",
                "node:http",
                "node:https",
                "net/http",
                "github.com/rabbitmq/",
            )
        ):
            return "shared persistence adaptor cannot perform network delivery"
        if role in {"messaging", "realtime"} and target.startswith(
            (
                "psycopg",
                "pg",
                "database/sql",
                "github.com/jackc/",
            )
        ):
            return "shared delivery adaptor must use persistence capabilities for database access"
    return None
