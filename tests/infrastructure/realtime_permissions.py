"""Prove a running publisher cannot cross the gateway's authority boundary."""
import os
from pathlib import Path
import sys

ROOT = Path(__file__).resolve().parents[2]
sys.path.insert(0, str(ROOT / "scripts"))
from dev import compose, load


PROBE = '''
import json
import os
import urllib.error
import urllib.request
from operations.foundation.secrets import secret

url = os.environ["REALTIME_GATEWAY_URL"]
for owner in ("preparation", "collection"):
    key = secret(owner.upper()+"_REALTIME_KEY")
    probes = [
        ("/api/publish", {"channel": "cafe:"+owner, "b64data": "AA==", "idempotency_key": "probe"}, 400),
        ("/api/publish", {"channel": "cafe:menu", "b64data": "AA==", "idempotency_key": "probe"}, 403),
        ("/api/disconnect", {"user": "permission-probe"}, 403),
        ("/api/info", {}, 404),
        ("/api/history", {"channel": "cafe:menu"}, 404),
    ]
    for path, body, expected in probes:
        request = urllib.request.Request(url+path, data=json.dumps(body).encode(), method="POST",
            headers={"Content-Type": "application/json", "X-API-Key": key})
        try:
            with urllib.request.urlopen(request, timeout=5) as response:
                status = response.status
        except urllib.error.HTTPError as error:
            status = error.code
        assert status == expected, f"{owner} {path}: expected {expected}, received {status}"
print("Realtime gateway validates own publications and denies foreign publication and administration")
'''


def main():
    env_file = Path(os.environ["CAFE_ENV_FILE"])
    if not load(env_file)["COMPOSE_PROJECT_NAME"].startswith("cafe-reference-test-"):
        raise RuntimeError("Permission probes require the disposable integration project")
    # Secrets are read inside the authorised workload and never returned to the host.
    # Invalid own bytes prove authentication reached validation without publishing.
    compose(env_file, "exec", "-T", "operations", "/app/.venv/bin/python", "-",
            input=PROBE, text=True)


if __name__ == "__main__":
    main()
