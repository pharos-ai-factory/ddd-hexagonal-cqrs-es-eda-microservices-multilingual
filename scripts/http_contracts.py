"""Bundle authoritative OpenAPI documents into each owning Go module."""
import argparse
import copy
import json
from pathlib import Path

ROOT = Path(__file__).resolve().parents[1]
SOURCE = ROOT / "contracts/http"
OUTPUTS = {
    "api": "services/api/adaptors/openapi/generated",
    "gateway": "services/api/adaptors/openapi/generated",
    "storefront": "services/storefront/foundation/transport/openapi/generated",
    "provider": "services/storefront/foundation/transport/openapi/generated",
}


def bundle(path: Path) -> dict:
    document = json.loads(path.read_text())
    components = document.setdefault("components", {})
    imported = set()

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
            raise ValueError("HTTP references must stay inside contracts/http")
        parts = fragment.removeprefix("/").split("/")
        content = json.loads(target.read_text())
        for part in parts:
            content = content[part.replace("~1", "/").replace("~0", "~")]
        if parts[0] in {"schemas", "parameters", "responses"}:
            section = parts[0]
            name = target.stem + "-" + parts[1]
            key = (section, name)
            if key not in imported:
                imported.add(key)
                components.setdefault(section, {})[name] = resolve(content, target)
            return {"$ref": f"#/components/{section}/{name}"}
        return resolve(content, target)

    # Resolve paths first; imported reusable components are then attached once.
    document["paths"] = resolve(copy.deepcopy(document["paths"]), path.resolve())
    return document


def generated() -> dict[Path, str]:
    return {
        ROOT / directory / (name + ".openapi.json"):
        json.dumps(bundle(SOURCE / (name + ".openapi.json")), indent=2, ensure_ascii=False) + "\n"
        for name, directory in OUTPUTS.items()
    }


def generate(check=False):
    for path, expected in generated().items():
        if check:
            if not path.exists() or path.read_text() != expected:
                raise ValueError(f"OpenAPI bundle drift: {path.relative_to(ROOT)}; run pnpm generate:http")
        else:
            path.parent.mkdir(parents=True, exist_ok=True)
            path.write_text(expected)


if __name__ == "__main__":
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("--check", action="store_true")
    generate(parser.parse_args().check)
