"""Bundle authoritative OpenAPI documents into each owning Go module."""

import argparse
import copy
import json
import shutil
from typed_http_boundaries import OUTPUT as MAPPING_OUTPUT, generate as mapping_generated
from orval_contracts import (
    LEGACY_OUTPUT as LEGACY_FRONTEND_OUTPUT,
    OUTPUT as FRONTEND_OUTPUT,
    check as check_frontend,
    generate as generate_frontend,
)
from pathlib import Path
from contract_selection import declarations

ROOT = Path(__file__).resolve().parents[1]
SOURCE = ROOT / "contracts"
DOCUMENTS = {
    "api": "services/api/http_api/api.openapi.json",
    "storefront": "services/storefront/http_api/storefront.openapi.json",
    "gateway": "services/realtime_gateway/http_api/gateway.openapi.json",
    "provider": "services/notification_provider/http_api/provider.openapi.json",
}
OUTPUTS = {
    "api": "services/api/adaptors/openapi/generated",
    "gateway": "services/api/adaptors/openapi/generated",
    "storefront": "services/storefront/adaptors/http/openapi/generated",
    "provider": "services/storefront/adaptors/http/openapi/generated",
}


def bundle(path: Path) -> dict:
    document = json.loads(path.read_text())
    components = document.setdefault("components", {})
    imported = {}

    def resolve(value, source):
        if isinstance(value, list):
            return [resolve(item, source) for item in value]
        if not isinstance(value, dict):
            return value
        if "$ref" not in value:
            return {key: resolve(item, source) for key, item in value.items()}
        if len(value) != 1:
            raise ValueError("Reference siblings are unsupported: " + str(source))
        filename, fragment = value["$ref"].split("#", 1)
        target = (source.parent / filename).resolve() if filename else source
        if not target.is_relative_to(SOURCE.resolve()):
            raise ValueError("HTTP references must stay inside contracts")
        parts = fragment.removeprefix("/").split("/")
        content = json.loads(target.read_text())
        for part in parts:
            content = content[part.replace("~1", "/").replace("~0", "~")]
        if parts[0] in {"schemas", "parameters", "responses"}:
            section = parts[0]
            relative = target.relative_to(SOURCE.resolve())
            namespace = ".".join((*relative.parts[:-1], target.stem))
            name = namespace + "." + parts[1]
            key = (section, name)
            origin = (target, fragment)
            if key in imported and imported[key] != origin:
                raise ValueError("HTTP component name collision: " + name)
            if key not in imported:
                imported[key] = origin
                components.setdefault(section, {})[name] = resolve(content, target)
            return {"$ref": f"#/components/{section}/{name}"}
        return resolve(content, target)

    # Resolve paths first; imported reusable components are then attached once.
    document["paths"] = resolve(copy.deepcopy(document["paths"]), path.resolve())
    return document


def generated() -> dict[Path, str]:
    usage = declarations(ROOT)["openapi"]
    supported = {
        "api": {"produces": {"api", "gateway"}, "consumes": set()},
        "storefront": {"produces": {"storefront", "provider"}, "consumes": set()},
        "web": {"produces": set(), "consumes": {"api"}},
    }
    for service, roles in usage.items():
        for role, names in roles.items():
            if set(names) - supported.get(service, {}).get(role, set()):
                raise ValueError(f"No OpenAPI generation target for {service}.{role}: {names}")
    produced = {name for roles in usage.values() for name in roles["produces"]}
    outputs = {
        ROOT / directory / (name + ".openapi.json"): json.dumps(
            bundle(SOURCE / DOCUMENTS[name]), indent=2, ensure_ascii=False
        )
        + "\n"
        for name, directory in OUTPUTS.items()
        if name in produced
    }
    if "api" not in produced:
        return outputs
    mapping = mapping_generated(ROOT, bundle(SOURCE / DOCUMENTS["api"]))
    import subprocess

    gofmt = ROOT / ".tools/go/bin/gofmt"
    if not gofmt.exists():
        import shutil

        gofmt = Path(shutil.which("gofmt") or "gofmt")
    mapping = subprocess.check_output([str(gofmt)], input=mapping, text=True)
    outputs[ROOT / MAPPING_OUTPUT] = mapping
    return outputs


def generate(check=False):
    expected_outputs = generated()
    usage = declarations(ROOT)["openapi"]
    frontend = "api" in usage.get("web", {}).get("consumes", [])
    frontend_document = bundle(SOURCE / DOCUMENTS["api"]) if frontend else None
    managed = {ROOT / MAPPING_OUTPUT}
    for directory in set(OUTPUTS.values()):
        managed.update((ROOT / directory).glob("*.openapi.json"))
    for path in managed - expected_outputs.keys():
        if path.exists():
            if check:
                raise ValueError(f"Stale OpenAPI output: {path.relative_to(ROOT)}")
            path.unlink()
    for path, expected in expected_outputs.items():
        if check:
            if not path.exists() or path.read_text() != expected:
                raise ValueError(f"OpenAPI bundle drift: {path.relative_to(ROOT)}; run pnpm generate:http")
        else:
            path.parent.mkdir(parents=True, exist_ok=True)
            path.write_text(expected)
    legacy = ROOT / LEGACY_FRONTEND_OUTPUT
    if check:
        if frontend_document is not None:
            check_frontend(ROOT, frontend_document)
        elif (ROOT / FRONTEND_OUTPUT).exists() or legacy.exists():
            raise ValueError("Stale frontend OpenAPI output")
    elif frontend_document is not None:
        if legacy.exists():
            legacy.unlink()
        generate_frontend(ROOT, frontend_document, ROOT / FRONTEND_OUTPUT)
    else:
        if legacy.exists():
            legacy.unlink()
        if (ROOT / FRONTEND_OUTPUT).exists():
            shutil.rmtree(ROOT / FRONTEND_OUTPUT)


if __name__ == "__main__":
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("--check", action="store_true")
    generate(parser.parse_args().check)
