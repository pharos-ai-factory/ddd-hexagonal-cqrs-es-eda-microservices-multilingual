"""Generate response models and schema-bound query registrations for each owner."""

import subprocess
from request_payloads import ROOT, load
from generate_response_go import generate as go
from generate_response_python import generate as python
from generate_response_typescript import generate as typescript


def generate(root=ROOT):
    schema = load(root)
    catalogue = {m["type"]: m for m in schema["responses"]}
    for owner in ("shared", "menu", "ordering", "preparation", "collection", "loyalty", "communication"):
        messages = [m for m in schema["responses"] if m["owner"] == owner]
        requests = [r for r in schema["payloads"] if r["owner"] == owner]
        for owners, render in [
            (("shared", "menu", "ordering"), go),
            (("shared", "preparation", "collection"), python),
            (("shared", "loyalty", "communication"), typescript),
        ]:
            if owner in owners:
                render(root, owner, messages, requests, catalogue)
    paths = [
        *root.glob("services/storefront/contracts/requests/generated/model/*/responses.go"),
        *root.glob("services/storefront/adaptors/messaging/generated/responses/*.go"),
        *root.glob("services/storefront/contexts/*/adaptors/messaging/generated/queryregistration/*.go"),
        *root.glob("services/storefront/**/request_responses_generated.go"),
    ]
    subprocess.run(["python3", str(ROOT / "scripts/go.py"), "gofmt", "-w", *map(str, paths)], check=True)


if __name__ == "__main__":
    generate()
