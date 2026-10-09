"""Live diagnostic authentication and durable backlog response contracts."""
import json
import os
from pathlib import Path
import sys
import urllib.error
import urllib.request

ROOT = Path(__file__).resolve().parents[2]
sys.path.insert(0, str(ROOT / "scripts"))
from dev import load


def request(url, key=None):
    headers = {"Authorization": "Bearer "+key} if key else {}
    try:
        response = urllib.request.urlopen(urllib.request.Request(url, headers=headers), timeout=10)
    except urllib.error.HTTPError as error:
        response = error
    with response:
        return response.status, json.load(response)


def main():
    values = load(Path(os.environ["CAFE_ENV_FILE"]))
    if not values["COMPOSE_PROJECT_NAME"].startswith("cafe-reference-test-"):
        raise RuntimeError("Diagnostic probes require a disposable integration project")
    for service, owners in (("STOREFRONT", {"menu", "ordering"}),
                            ("OPERATIONS", {"preparation", "collection"}),
                            ("ENGAGEMENT", {"loyalty", "communication"}), ("API", set())):
        url = "http://127.0.0.1:"+values[service+"_PORT"]
        key = values[service+"_API_KEY"] if owners else values["API_KEY"]
        assert request(url+"/healthz")[0] == 200
        assert request(url+"/diagnostics")[0] == 401
        assert request(url+"/diagnostics", "incorrect")[0] == 401
        status, state = request(url+"/diagnostics", key)
        assert status == 200, service+" diagnostics unavailable"
        if owners:
            assert set(state["databases"]) == owners
            for metrics in state["databases"].values():
                for kind in ("outbox", "realtime"):
                    assert metrics[kind]["pending"] >= 0
                    assert metrics[kind]["oldestAgeSeconds"] >= 0
                    assert metrics[kind]["failed"] >= 0
            assert "process lifetime" in state["counterScope"]
        else:
            assert state["pendingRevocations"] >= 0
            assert state["expiredSessionsAwaitingRevocation"] >= 0
        assert key not in json.dumps(state), "Diagnostic response leaked its credential"
        if owners:
            assert_business_http_absent(url, key)
    print("All runtimes protect diagnostics and expose owned durable backlog metrics")


def assert_business_http_absent(url, key):
    """Probe the running composition with valid owner credentials and every public business path."""
    document = json.loads((ROOT/'contracts/services/api/http_api/api.openapi.json').read_text())
    paths = [path.removeprefix('/api').replace('{id}', '11111111-1111-4111-8111-111111111111')
             for path in document['paths'] if path.startswith('/api/v1/')]
    assert paths
    for path in paths:
        for method in ('GET', 'POST'):
            probe = urllib.request.Request(url+path, method=method,
                data=b'{}' if method == 'POST' else None,
                headers={'Authorization': 'Bearer '+key, 'Content-Type': 'application/json'})
            try:
                response = urllib.request.urlopen(probe, timeout=10)
            except urllib.error.HTTPError as error:
                response = error
            with response:
                assert response.status == 404, method+' '+url+path+' exposed an owner business route'


if __name__ == "__main__":
    main()
