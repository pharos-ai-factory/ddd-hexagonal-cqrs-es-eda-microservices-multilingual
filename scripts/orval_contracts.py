"""Generate and compare the browser's Orval HTTP client."""

import json
import os
from pathlib import Path
import subprocess
import tempfile

OUTPUT = Path("services/web/src/adaptors/generated/http")
LEGACY_OUTPUT = Path("services/web/src/adaptors/generated/http.ts")
CONFIG = Path("services/web/orval.config.ts")


def generate(root: Path, document: dict, output: Path) -> None:
    """Generate the complete Orval tree from an assembled OpenAPI document."""
    binary = root / "node_modules/.bin/orval"
    config = root / CONFIG
    if not binary.exists():
        raise ValueError("Orval is unavailable; run pnpm install --frozen-lockfile")
    if not config.exists():
        raise ValueError(f"Orval configuration is unavailable: {CONFIG}")
    with tempfile.TemporaryDirectory() as directory:
        source = Path(directory) / "api.openapi.json"
        source.write_text(json.dumps(document, indent=2, ensure_ascii=False) + "\n")
        environment = dict(
            os.environ,
            CAFE_ORVAL_INPUT=str(source),
            CAFE_ORVAL_OUTPUT=str(output),
        )
        subprocess.run(
            [str(binary), "--config", str(config)],
            cwd=root,
            env=environment,
            check=True,
        )
    for path in output.rglob("*.ts"):
        lines = (line.rstrip() for line in path.read_text().splitlines())
        path.write_text("\n".join(lines).rstrip() + "\n")


def files(directory: Path) -> dict[Path, bytes]:
    """Return the generated file tree keyed relative to its output directory."""
    if not directory.exists():
        return {}
    return {path.relative_to(directory): path.read_bytes() for path in directory.rglob("*") if path.is_file()}


def check(root: Path, document: dict) -> None:
    """Regenerate in isolation and reject added, removed or changed Orval output."""
    actual = root / OUTPUT
    with tempfile.TemporaryDirectory() as directory:
        expected = Path(directory) / "http"
        generate(root, document, expected)
        expected_files = files(expected)
    actual_files = files(actual)
    changed = sorted(
        str(path)
        for path in expected_files.keys() | actual_files.keys()
        if expected_files.get(path) != actual_files.get(path)
    )
    if changed:
        raise ValueError("Frontend OpenAPI client drift; run pnpm generate:http:\n" + "\n".join(changed))
    if (root / LEGACY_OUTPUT).exists():
        raise ValueError(f"Stale frontend OpenAPI output: {LEGACY_OUTPUT}")
