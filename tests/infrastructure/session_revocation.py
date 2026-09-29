"""Exercise the real API/session worker against isolated Valkey and Centrifugo."""
from contextlib import ExitStack
import json
import os
from pathlib import Path
import secrets
import subprocess
import sys
import tempfile
import time
import urllib.error
import urllib.request

ROOT = Path(__file__).resolve().parents[2]
sys.path.insert(0, str(ROOT/"scripts"))
from dev import compose, free_port, load


def eventually(check, description):
    deadline = time.monotonic()+15
    while not check():
        if time.monotonic() >= deadline:
            raise RuntimeError(description)
        time.sleep(0.05)


def healthy(url):
    try:
        with urllib.request.urlopen(url, timeout=1) as response:
            return response.status == 200
    except (OSError, urllib.error.URLError):
        return False


def stop(process):
    process.terminate()
    try:
        process.wait(timeout=6)
    except subprocess.TimeoutExpired:
        process.kill()
        process.wait()


def run():
    env_file = Path(os.environ["CAFE_ENV_FILE"])
    source = load(env_file)
    if not source["COMPOSE_PROJECT_NAME"].startswith("cafe-reference-test-"):
        raise RuntimeError("Revocation tests require the disposable integration project")
    # Use the same pinned images as the repository composition.
    configuration = compose(env_file, "config", "--format", "json", stdout=subprocess.PIPE, text=True)
    images = json.loads(configuration.stdout)["services"]
    folder = ROOT/".local"
    binary = folder/"revocation-fixture"
    subprocess.run([sys.executable, "scripts/go.py", "build", "-o", "../../.local/revocation-fixture",
                    "./tests/revocation"], cwd=ROOT,
                   env=dict(os.environ, CAFE_GO_PROJECT="services/api"), check=True)
    with ExitStack() as cleanup:
        directory = Path(cleanup.enter_context(tempfile.TemporaryDirectory(prefix="revocation-", dir=folder)))
        api, realtime, admin, valkey = (free_port() for _ in range(4))
        values = {key: secrets.token_hex(24) for key in (
            "SESSION_PASSWORD", "REALTIME_REDIS_PASSWORD", "CENTRIFUGO_API_KEY",
            "OPERATOR_PASSWORD", "API_KEY", "CONNECT_PROXY_SECRET")}
        values.update(API_PORT=str(api), REALTIME_PORT=str(realtime), REALTIME_ADMIN_PORT=str(admin),
            CAFE_DISPOSABLE_PROJECT=source["COMPOSE_PROJECT_NAME"], LISTEN_ADDR=f"127.0.0.1:{api}",
            WEB_ORIGIN=f"http://127.0.0.1:{api}", VALKEY_ADDRESS=f"127.0.0.1:{valkey}",
            CENTRIFUGO_API_URL=f"http://127.0.0.1:{admin}")
        environment = dict(os.environ, **values)
        fixture_env = directory/"fixture.env"
        fixture_env.write_text("".join(f"{key}={value}\n" for key, value in values.items()))
        fixture_env.chmod(0o600)
        config = json.loads((ROOT/"devops/centrifugo/config.json").read_text())
        config["http_server"] = {"port": realtime, "internal_port": admin}
        # Authentication is under test; this separate broker needs no shared history.
        config.pop("engine")
        config["client"]["proxy"]["connect"]["endpoint"] = f"http://127.0.0.1:{api}/api/realtime/connect"
        # The deliberate hold must outlive the fence even on a busy test host.
        # Cancellation is separately rejected by the browser evidence.
        config["client"]["proxy"]["connect"]["timeout"] = "10s"
        config["client"]["proxy"]["refresh"]["endpoint"] = f"http://127.0.0.1:{api}/api/realtime/refresh"
        config_file = directory/"centrifugo.json"
        config_file.write_text(json.dumps(config))
        environment.update(CENTRIFUGO_HTTP_API_KEY=values["CENTRIFUGO_API_KEY"],
            CENTRIFUGO_VAR_CONNECT_PROXY_SECRET=values["CONNECT_PROXY_SECRET"],
            CENTRIFUGO_CLIENT_ALLOWED_ORIGINS=values["WEB_ORIGIN"])
        label = secrets.token_hex(6)
        for name, arguments in (
            ("valkey", ["-p", f"127.0.0.1:{valkey}:6379", "--tmpfs", "/data",
                "-e", "SESSION_PASSWORD", "-e", "REALTIME_REDIS_PASSWORD",
                "-v", str(ROOT/"devops/valkey/start.sh")+":/reference/start.sh:ro",
                "-v", str(ROOT/"devops/read-secret.sh")+":/reference/read-secret.sh:ro",
                "--entrypoint", "/bin/sh", images["valkey"]["image"], "/reference/start.sh"]),
            ("centrifugo", ["--network", "host", "-e", "CENTRIFUGO_HTTP_API_KEY",
                "-e", "CENTRIFUGO_VAR_CONNECT_PROXY_SECRET", "-e", "CENTRIFUGO_CLIENT_ALLOWED_ORIGINS",
                "-v", str(config_file)+":/etc/centrifugo/config.json:ro", images["centrifugo"]["image"],
                "centrifugo", "--config=/etc/centrifugo/config.json"]),
        ):
            container = "cafe-revocation-"+name+"-"+label
            subprocess.run(["docker", "run", "--rm", "-d", "--name", container, *arguments],
                           env=environment, check=True, stdout=subprocess.DEVNULL)
            cleanup.callback(subprocess.run, ["docker", "rm", "-f", container],
                             stdout=subprocess.DEVNULL, stderr=subprocess.DEVNULL, check=False)
        def valkey_ready():
            result = subprocess.run(["docker", "exec", "cafe-revocation-valkey-"+label, "/bin/sh", "-c",
                'valkey-cli --user sessions --pass "$SESSION_PASSWORD" --no-auth-warning ping'],
                capture_output=True, text=True)
            return result.returncode == 0 and result.stdout.strip() == "PONG"
        eventually(valkey_ready, "Fixture Valkey did not become ready")
        eventually(lambda: healthy(f"http://127.0.0.1:{admin}/health"), "Fixture Centrifugo did not become ready")
        log = cleanup.enter_context((folder/"revocation-fixture.log").open("w"))
        host = subprocess.Popen([str(binary)], env=environment, stdout=log, stderr=subprocess.STDOUT)
        cleanup.callback(stop, host)
        eventually(lambda: healthy(f"http://127.0.0.1:{api}/healthz"), "Fixture API did not become ready")
        subprocess.run([sys.executable, "scripts/browser.py", "--config", "tests/infrastructure/playwright.config.ts"],
                       cwd=ROOT, env=dict(os.environ, CAFE_ENV_FILE=str(fixture_env)), check=True)


if __name__ == "__main__":
    run()
