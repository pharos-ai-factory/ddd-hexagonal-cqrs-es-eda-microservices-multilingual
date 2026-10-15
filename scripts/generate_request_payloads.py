"""Generate application-facing command/query contracts and mechanical boundary mappings."""

from pathlib import Path
import subprocess
from request_payloads import load, ROOT
from contract_sources import PRIVATE
from generate_request_go import generate as go
from generate_request_python import generate as python
from generate_request_typescript import generate as typescript


def generate(root=ROOT):
    schema = load(root)
    for owner in ("shared", *PRIVATE):
        messages = (
            schema["helpers"] if owner == "shared" else [m for m in schema["payloads"] if m["owner"] == owner]
        )
        if owner in ("shared", "menu", "ordering"):
            go(root, owner, messages)
        if owner in ("shared", "preparation", "collection"):
            python(root, owner, messages)
        if owner in ("shared", "loyalty", "communication"):
            typescript(root, owner, messages)
    paths = [
        *root.glob("services/storefront/contracts/requests/model/**/*_generated.go"),
        *root.glob("services/storefront/contexts/*/adaptors/messaging/request_payloads_generated.go"),
    ]
    subprocess.run(
        ["python3", str(Path(__file__).with_name("go.py")), "gofmt", "-w", *map(str, paths)], check=True
    )


if __name__ == "__main__":
    generate()
