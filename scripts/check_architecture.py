#!/usr/bin/env python3
"""Check service ownership and inward dependencies in all implementation languages."""

from python_imports import python_imports
from command_registration_policy import violation as registration_rule
from request_contracts import OWNED
from event_boundaries import violation as event_boundary_violation
from generated_files import violation as generated_file_violation
from contract_sources import sources, catalogue, command_subscriptions
from migration_catalogue import metadata, OUTPUTS
from owner_http_boundary import violation as owner_http_violation
from adaptor_layout import layout_violation as adaptor_layout_violation
from adaptor_layout import import_violation as adaptor_import_violation
from hashlib import sha256
import json
import os
from request_contract_policy import plain_request_or_event
from response_contract_policy import violation as response_rule
from pathlib import Path
import re
import subprocess
import sys

ROOT = Path(__file__).resolve().parents[1]
MODULE = "github.com/pharos-ai-factory/ddd-hexagonal-cqrs-es-eda-microservices-multilingual"
SERVICES = {
    "storefront": {"menu", "ordering"},
    "operations": {"preparation", "collection"},
    "engagement": {"loyalty", "communication"},
}
IGNORED = {
    ".git",
    ".tools",
    ".local",
    "node_modules",
    "__pycache__",
    ".venv",
    ".next",
    "dist",
    "out",
    "test-results",
    "playwright-report",
    ".pytest_cache",
    ".ruff_cache",
    ".mypy_cache",
}
GENERATED = {
    "services/storefront/contracts/events/generated/cafe/v1/events.pb.go",
    "services/storefront/contracts/realtime/generated/cafe/realtime/v1/realtime.pb.go",
    "services/operations/src/operations/adaptors/generated/cafe/v1/events_pb2.py",
    "services/operations/src/operations/adaptors/generated/cafe/realtime/v1/realtime_pb2.py",
    "services/operations/src/operations/adaptors/generated/cafe/v1/events_pb2.pyi",
    "services/operations/src/operations/adaptors/generated/cafe/realtime/v1/realtime_pb2.pyi",
    "services/engagement/src/adaptors/generated/events.json",
    "services/engagement/src/adaptors/generated/realtime.json",
    "services/engagement/src/adaptors/generated/requests.json",
    "services/engagement/src/adaptors/generated/request-types.ts",
    "services/web/src/adaptors/generated/realtime.json",
    "services/api/adaptors/http/backend/generated/mapping.go",
    "services/storefront/adaptors/persistence/postgres/migrations/generated/contexts.json",
    "services/storefront/adaptors/persistence/postgres/migrations/generated/0001_initial.sql",
    "services/storefront/adaptors/persistence/postgres/migrations/generated/0002_realtime.sql",
    "services/operations/src/operations/adaptors/generated/context-persistence.json",
    "services/operations/src/operations/adaptors/generated/request_registry.py",
    "services/engagement/src/adaptors/generated/context-persistence.json",
    "services/api/adaptors/openapi/generated/api.openapi.json",
    "services/api/adaptors/openapi/generated/gateway.openapi.json",
    "services/storefront/adaptors/http/openapi/generated/storefront.openapi.json",
    "services/storefront/adaptors/http/openapi/generated/provider.openapi.json",
    "pnpm-lock.yaml",
    "services/operations/uv.lock",
}

# Derive exact compiler paths; each service consumes its owned request packages.

for contract in ("events", "realtime", "requests"):
    package = {"events": "cafe/v1", "realtime": "cafe/realtime/v1"}
    for logical in sources(contract):
        stem = Path(logical).stem
        relative = Path(logical).with_suffix("").as_posix()
        owner = logical.split("/contexts/")[1].split("/")[0] if "/contexts/" in logical else None
        if contract != "requests" or owner is None or owner in OWNED["storefront"]:
            go_folder = (
                str(Path(logical).parent) + ("/shared" if stem in ("common", "validation") else "")
                if contract == "requests"
                else package[contract]
            )
            GENERATED.add(f"services/storefront/contracts/{contract}/generated/{go_folder}/{stem}.pb.go")
        if contract == "requests":
            GENERATED.add(
                f"services/api/adaptors/messaging/generated/{Path(logical).parent}"
                + ("/shared" if stem in ("common", "validation") else "")
                + f"/{stem}.pb.go"
            )
        if contract != "requests" or owner is None or owner in OWNED["operations"]:
            for suffix in (".py", ".pyi"):
                GENERATED.add(f"services/operations/src/operations/adaptors/generated/{relative}_pb2{suffix}")
for folder in (
    "services/storefront/contracts/requests/generated",
    "services/api/adaptors/messaging/generated",
):
    for name in ("envelopes.go",):
        GENERATED.add(folder + "/cafe/requests/v1/" + name)
for owner in ("menu", "ordering"):
    GENERATED.add(f"services/storefront/contexts/{owner}/adaptors/messaging/generated/{owner}_private.pb.go")
GENERATED.add("services/storefront/apps/topology/generated.json")
GENERATED.add("services/storefront/apps/topology/permissions.generated.json")
GENERATED.add("services/storefront/apps/replay/generated.json")
for owner in ("loyalty", "communication"):
    GENERATED.add(
        f"services/engagement/src/contexts/{owner}/adaptors/messaging/generated/private_messages.json"
    )

for owner in ("preparation", "collection"):
    for suffix in ("py", "pyi"):
        GENERATED.add(
            f"services/operations/src/operations/adaptors/generated/cafe/internal/{owner}/internal_commands_pb2.{suffix}"
        )
    for suffix in ("py", "pyi"):
        GENERATED.add(
            f"services/operations/src/operations/adaptors/generated/cafe/internal/{owner}/private_messages_pb2.{suffix}"
        )
for owner in ("loyalty", "communication"):
    GENERATED.add(
        f"services/engagement/src/contexts/{owner}/adaptors/messaging/generated/internal_commands.json"
    )


# Isolated regeneration checks these schema-owned outputs.
for pattern in (
    "services/storefront/apps/support/subscriptions_generated.go",
    "services/storefront/contracts/events/generated/model/*.go",
    "services/storefront/contracts/requests/generated/model/*/*.go",
    "services/storefront/adaptors/messaging/generated/responses/*.go",
    "services/storefront/contexts/*/adaptors/messaging/request_responses_generated.go",
    "services/storefront/contexts/*/adaptors/messaging/generated/queryregistration/*.go",
    "services/storefront/contexts/*/adaptors/messaging/request_payloads_generated.go",
    "services/operations/src/operations/contracts/requests/generated/*.py",
    "services/operations/src/operations/contexts/*/adaptors/messaging/generated/request_payloads.py",
    "services/engagement/src/contracts/requests/generated/*.ts",
    "services/engagement/src/contexts/*/adaptors/messaging/generated/request-payloads.ts",
    "services/storefront/contracts/events/protobuf/*_generated.go",
    "services/storefront/contexts/*/adaptors/messaging/subscriptions_generated.go",
    "services/operations/src/operations/contracts/events/generated/*.py",
    "services/operations/src/operations/adaptors/generated/event_*.py",
    "services/operations/src/operations/adaptors/generated/private_event_registry.py",
    "services/engagement/src/contracts/events/generated/*.ts",
    "services/engagement/src/adaptors/generated/event-*",
    "services/engagement/src/adaptors/generated/private-event-registry.ts",
    "services/engagement/src/adaptors/generated/*-names.ts",
    "services/operations/src/operations/adaptors/generated/*_names.py",
    "services/storefront/contracts/events/generated/names/*.go",
    "services/*/**/generated/cafe/requests/v1/names.go",
    "services/web/src/adaptors/generated/http/**/*.ts",
):
    GENERATED.update(path.relative_to(ROOT).as_posix() for path in ROOT.glob(pattern))


def import_violation(package, imported):
    path = package.removeprefix(MODULE + "/")
    target = imported.removeprefix(MODULE + "/")
    internal = imported.startswith(MODULE + "/")
    if reason := adaptor_import_violation(path + "/", target):
        return reason
    if path.startswith("services/"):
        service = path.split("/")[1]
        if internal and not target.startswith("services/" + service + "/"):
            return "cross-service source dependency"
        path = "/".join(path.split("/")[2:])
        if internal:
            target = "/".join(target.split("/")[2:])
    else:
        service = "storefront"
    if imported.startswith(("go.uber.org/fx", "go.uber.org/dig")) and not path.startswith("apps/"):
        return "DI framework outside composition"
    if path.startswith("contexts/"):
        owner, ring = path.split("/")[1:3]
        if internal and target.startswith("contexts/") and target.split("/")[1] != owner:
            return "foreign context implementation"
        if (
            ring == "application"
            and any(part in path.split("/") for part in ("queries", "readmodels"))
            and target == f"contexts/{owner}/domain"
        ):
            return "query capabilities must use application read models"
        if ring in {"domain", "application"}:
            allowed = [f"contexts/{owner}/domain", "foundation/domain"]
            if ring == "application":
                allowed += [
                    f"contexts/{owner}/application",
                    "foundation/application",
                    "contracts/events/generated/model",
                    f"contracts/requests/generated/model/{owner}",
                    "contracts/requests/generated/model/shared",
                ]
            if internal and not any(target == p or target.startswith(p + "/") for p in allowed):
                return "outward core dependency"
            if not internal and (
                "." in imported.split("/")[0] or imported.split("/")[0] in {"net", "os", "database", "log"}
            ):
                return "infrastructure in core"
    if (
        path.startswith("foundation/")
        and internal
        and not target.startswith(("foundation/", "contracts/requests/generated/model/shared"))
    ):
        return "business dependency in technical foundation"
    if path.startswith("apps/") and target.startswith("contexts/"):
        if target.split("/")[1] not in SERVICES.get(service, set()):
            return "service imports a context it does not own"
        if path.startswith("apps/support"):
            return "shared composition contains context behaviour"
    if (
        service == "api"
        and path.startswith("application/")
        and internal
        and not target.startswith("application/")
    ):
        return "outward API application dependency"
    return None


def core_violation(path, target):
    """Normalised service-local Python/TypeScript module paths."""
    if reason := adaptor_import_violation(path, target):
        return reason
    if path.startswith("tests/acceptance/") and target.startswith("services/"):
        return "cross-service acceptance must use published boundaries"
    parts = path.split("/")
    if "foundation" in parts and "/contexts/" in "/" + target:
        return "business dependency in technical foundation"
    if target.startswith(("awilix", "dependency_injector")) and "apps" not in parts:
        return "DI framework outside composition"
    if "contexts" not in parts:
        if (
            "adaptors" in parts
            and not path.endswith("/adaptors/generated/private_event_registry.py")
            and "/contexts/" in "/" + target
            and "/generated/" not in "/" + target
            and not path.endswith((".integration.ts", ".test.ts"))
        ):
            return "shared adaptor imports context implementation"
        return None
    offset = parts.index("contexts")
    owner, ring = parts[offset + 1], parts[offset + 2].split(".")[0]
    if "contexts/" in target and target.split("contexts/")[1].split("/")[0] != owner:
        return "foreign context implementation"
    if (
        ring == "application"
        and any(p in parts for p in ("queries", "read_models", "read-models"))
        and "/domain/" in "/" + target
        and "/contexts/" in "/" + target
    ):
        return "query capabilities must use application read models"
    if ring in {"domain", "application"}:
        service = parts[1] if parts[0] == "services" else ""
        plain_contract = plain_request_or_event(service, owner, target)
        if any(p in target.split("/") for p in ("adaptors", "apps")) or (
            "generated" in target.split("/") and not plain_contract
        ):
            return "outward core dependency"
        if ring == "domain" and any(p in target.split("/") for p in ("application", "contracts")):
            return "outward domain dependency"
        if target.startswith(
            (
                "node:",
                "pg",
                "amqplib",
                "protobuf",
                "react",
                "next",
                "fastapi",
                "pika",
                "psycopg",
                "requests",
                "http",
                "socket",
                "os",
                "awilix",
                "dependency_injector",
            )
        ):
            return "infrastructure in core"
    return None


def files():
    for directory, directories, names in os.walk(ROOT):
        directories[:] = [name for name in directories if name not in IGNORED]
        for name in names:
            yield Path(directory) / name


def check():
    errors, packages = [], []
    event_names = {item["name"] for item in catalogue()}
    for service in ("storefront", "api"):
        result = subprocess.run(
            [sys.executable, "scripts/go.py", "list", "-json", "./..."],
            cwd=ROOT,
            env=dict(os.environ, CAFE_GO_PROJECT="services/" + service),
            capture_output=True,
            text=True,
            check=True,
        )
        decoder, text = json.JSONDecoder(), result.stdout.strip()
        while text:
            package, offset = decoder.raw_decode(text)
            packages.append(package)
            text = text[offset:].lstrip()
    for package in packages:
        for imported in package.get("Imports", []):
            reason = import_violation(package["ImportPath"], imported)
            if reason:
                errors.append(f"{package['ImportPath']}: {reason}: {imported}")
    for path in files():
        relative = path.relative_to(ROOT).as_posix()
        if path.suffix not in {
            ".go",
            ".py",
            ".pyi",
            ".md",
            ".sql",
            ".yaml",
            ".proto",
            ".ts",
            ".tsx",
            ".json",
            ".mjs",
            ".css",
            ".feature",
        }:
            continue
        content = path.read_text()
        if reason := generated_file_violation(relative, content, relative in GENERATED):
            errors.append(f"{relative}: {reason}")
        if len(content.splitlines()) > 450 and relative not in GENERATED:
            errors.append(f"{relative} exceeds 450 lines without a responsibility review")
        if reason := adaptor_layout_violation(relative):
            errors.append(f"{relative}: {reason}")
        if (
            relative in GENERATED
            or path.name.endswith(("_test.go", ".test.ts", ".spec.ts"))
            or path.name.startswith("test_")
        ):
            continue
        if relative.startswith("tests/acceptance/") and path.suffix == ".ts":
            for target in re.findall(r"(?:from\s+|import\s*)[\'\"]([^\'\"]+)", content):
                normal = (
                    (path.parent / target).resolve().relative_to(ROOT).as_posix()
                    if target.startswith(".")
                    else target
                )
                if reason := core_violation(relative, normal):
                    errors.append(f"{relative}: {reason}: {target}")
        if reason := event_boundary_violation(relative, content, event_names):
            errors.append(f"{relative}: {reason}")
        if reason := response_rule(relative, content) or registration_rule(relative, content):
            errors.append(f"{relative}: {reason}")
        if reason := owner_http_violation(relative, content):
            errors.append(f"{relative}: {reason}")
        if relative.startswith("services/"):
            service = relative.split("/")[1]
            if "/contexts/" in relative and path.name != "__init__.py":
                owner = relative.split("/contexts/")[1].split("/")[0]
                if owner not in SERVICES.get(service, set()):
                    errors.append(f"{relative}: service does not own context {owner}")
            if path.suffix in {".py", ".pyi"}:
                for target in python_imports(path, content):
                    if target.startswith("services/") and target.split("/")[1] != service:
                        errors.append(f"{relative}: cross-service source dependency")
                    if reason := core_violation(relative, target):
                        errors.append(f"{relative}: {reason}: {target}")
            if path.suffix in {".ts", ".tsx"}:
                for target in re.findall(r"(?:from\s+|import\s*)[\'\"]([^\'\"]+)", content):
                    normal = (
                        (path.parent / target).resolve().relative_to(ROOT).as_posix()
                        if target.startswith(".")
                        else target
                    )
                    if normal.startswith("services/") and normal.split("/")[1] != service:
                        errors.append(f"{relative}: cross-service source dependency")
                    if reason := core_violation(relative, normal):
                        errors.append(f"{relative}: {reason}: {target}")
            if re.search(r"/domain(?:/|\.py)", relative) and re.search(
                r"\btime\.Now\s*\(|\bDate\.now\s*\(|new Date\(\)|datetime\.now\s*\(", content
            ):
                errors.append(f"{relative}: implicit domain clock")
            if service == "web" and "/adaptors/http/" not in relative and re.search(r"\bfetch\s*\(", content):
                errors.append(f"{relative}: frontend HTTP calls must use the generated operation client")
            if service == "web" and re.search(r"\bnewSubscription\s*\(|\bsetInterval\s*\(", content):
                errors.append(f"{relative}: client-selected subscription or browser polling")
    if {p.name for p in (ROOT / "services").iterdir() if p.is_dir()} != {*SERVICES, "api", "web"}:
        errors.append("expected exactly three domain services, Go API and Next.js web")
    baseline = ROOT / "devops/postgres/bootstrap/0001_initial.sql"
    if (
        baseline.read_bytes()
        != (
            ROOT / "services/storefront/adaptors/persistence/postgres/migrations/generated/0001_initial.sql"
        ).read_bytes()
    ):
        errors.append("Go embedded migration differs from the shared persistence contract")
    try:
        for service, expected in metadata().items():
            if json.loads((ROOT / OUTPUTS[service]).read_text()) != expected:
                errors.append(service + ": context migration metadata drift; run pnpm generate:contracts")
    except (ValueError, KeyError, OSError) as error:
        errors.append(str(error))
    checksums = {
        str(version): sha256((ROOT / f"devops/postgres/bootstrap/{name}").read_bytes()).hexdigest()
        for version, name in ((1, "0001_initial.sql"), (2, "0002_realtime.sql"))
    }
    for folder in (
        "services/operations/src/operations/adaptors/generated",
        "services/engagement/src/adaptors/generated",
    ):
        service = "operations" if "operations" in folder else "engagement"
        if json.loads((ROOT / folder / "catalogue.json").read_text()) != catalogue(service):
            errors.append(folder + ": catalogue drift")
        if json.loads((ROOT / folder / "persistence.json").read_text()) != checksums:
            errors.append(folder + ": migration checksum drift")
    if (ROOT / "devops/postgres/bootstrap/0002_realtime.sql").read_bytes() != (
        ROOT / "services/storefront/adaptors/persistence/postgres/migrations/generated/0002_realtime.sql"
    ).read_bytes():
        errors.append("Go realtime migration drift")
    for name in ("events", "realtime", "requests"):
        if any(not path.exists() for path in sources(name).values()):
            errors.append("missing published specification: " + name)
    if json.loads((ROOT / "services/storefront/apps/topology/generated.json").read_text()) != [
        {**entry, "command": entry["consumer"] in {item["consumer"] for item in command_subscriptions()}}
        for entry in catalogue("topology")
    ]:
        errors.append("deployment topology metadata drift")
    compose = (ROOT / "devops/compose.yaml").read_text()
    if re.search(r"APP_ENV:\s*(production|staging)", compose):
        errors.append("non-development runtime composition")
    from command_boundaries import check as check_commands

    errors.extend(check_commands())
    from context_catalogue import check as check_navigation

    errors.extend(check_navigation())
    if errors:
        raise SystemExit("\n".join(errors))
    print(f"Architecture verified across {len(packages)} Go packages, Python, TypeScript and Next.js")


if __name__ == "__main__":
    check()
