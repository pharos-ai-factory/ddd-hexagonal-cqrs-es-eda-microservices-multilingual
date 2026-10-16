"""Recover a committed command after its RabbitMQ reply cannot be routed."""

import json
import os
from pathlib import Path
import subprocess
import sys
import urllib.error
import urllib.request
import uuid
import time
from concurrent.futures import ThreadPoolExecutor
from urllib.parse import quote

ROOT = Path(__file__).resolve().parents[2]
sys.path.insert(0, str(ROOT / "scripts"))
from dev import broker_request, compose, load


def main():
    env_file = Path(os.environ["CAFE_ENV_FILE"])
    values = load(env_file)
    if not values["COMPOSE_PROJECT_NAME"].startswith("cafe-reference-test-"):
        raise RuntimeError("Request recovery requires disposable infrastructure")
    drink, create_id, publish_id, correlation = (str(uuid.uuid4()) for _ in range(4))

    def command(action, body, version, identity, expected):
        request = urllib.request.Request(
            f"http://127.0.0.1:{values['API_PORT']}/api/v1/menu/drinks/{drink}{action}",
            data=json.dumps(body).encode(),
            method="POST",
            headers={
                "Authorization": "Bearer " + values["API_KEY"],
                "Content-Type": "application/json",
                "Idempotency-Key": identity,
                "If-Match": str(version),
                "X-Correlation-ID": correlation,
            },
        )
        try:
            response = urllib.request.urlopen(request, timeout=20)
        except urllib.error.HTTPError as error:
            response = error
        with response:
            result = json.load(response)
            assert response.status == expected, f"Expected {expected}, received {response.status}: {result}"
            return result

    def evidence():
        query = f"""SELECT json_build_object(
            'version',(SELECT version FROM cafe.aggregates WHERE kind='drink' AND id='{drink}'),
            'receipts',(SELECT count(*) FROM cafe.command_receipts WHERE aggregate_id='{drink}' AND command_id='{publish_id}'),
            'events',(SELECT count(*) FROM cafe.outbox_events WHERE aggregate_id='{drink}'),
            'publications',(SELECT count(*) FROM cafe.realtime_publications WHERE aggregate_id='{drink}'))"""
        result = compose(
            env_file,
            "exec",
            "-T",
            "postgres",
            "psql",
            "-U",
            "postgres",
            "-d",
            "cafe_menu",
            "-At",
            "-v",
            "ON_ERROR_STOP=1",
            "-c",
            query,
            stdout=subprocess.PIPE,
            text=True,
        )
        return json.loads(result.stdout)

    command("", {"name": "Reply recovery"}, 0, create_id, 200)
    # Hold the command advisory lock while an independent query reads committed state.
    label = "queue-isolation-" + drink

    def sql(source):
        result = compose(
            env_file,
            "exec",
            "-T",
            "postgres",
            "psql",
            "-U",
            "postgres",
            "-d",
            "cafe_menu",
            "-At",
            "-v",
            "ON_ERROR_STOP=1",
            "-c",
            source,
            stdout=subprocess.PIPE,
            stderr=subprocess.PIPE,
            text=True,
        )
        return result.stdout.strip()

    with ThreadPoolExecutor(max_workers=2) as pool:
        held = pool.submit(
            sql,
            f"SELECT set_config('application_name','{label}',false); SELECT pg_advisory_lock(hashtextextended('drink:{drink}',0)); SELECT pg_sleep(30);",
        )
        try:
            deadline = time.monotonic() + 5
            while (
                sql(
                    f"SELECT EXISTS(SELECT FROM pg_stat_activity a JOIN pg_locks l ON a.pid=l.pid WHERE a.application_name='{label}' AND l.locktype='advisory' AND l.granted)"
                )
                != "t"
            ):
                assert time.monotonic() < deadline, "Command fixture lock did not acquire"
                time.sleep(0.1)
            blocked = pool.submit(command, "", {"name": "Queue isolation"}, 1, str(uuid.uuid4()), 422)
            while sql("SELECT EXISTS(SELECT FROM pg_locks WHERE locktype='advisory' AND NOT granted)") != "t":
                assert time.monotonic() < deadline, "Command consumer did not wait on the owner transaction"
                time.sleep(0.1)
            request = urllib.request.Request(
                f"http://127.0.0.1:{values['API_PORT']}/api/v1/menu/drinks/{drink}",
                headers={"Authorization": "Bearer " + values["API_KEY"]},
            )
            with urllib.request.urlopen(request, timeout=3) as response:
                assert json.load(response)["version"] == 1
            assert not blocked.done(), "Command fixture completed before query isolation was proved"
        finally:
            sql(f"SELECT pg_terminate_backend(pid) FROM pg_stat_activity WHERE application_name='{label}'")
            try:
                held.result(timeout=5)
            except subprocess.CalledProcessError:
                pass
        blocked.result(timeout=15)
    print("A blocked command transaction leaves its owner's query consumer responsive")
    # Disconnect the originating API after its command reaches the owner but before
    # commit. Its exclusive queue disappears; a retry must use the new connection.
    label = "lost-command-reply-" + drink
    with ThreadPoolExecutor(max_workers=2) as pool:
        held = pool.submit(
            sql,
            f"SELECT set_config('application_name','{label}',false); SELECT pg_advisory_lock(hashtextextended('drink:{drink}',0)); SELECT pg_sleep(30);",
        )
        try:
            deadline = time.monotonic() + 5
            while (
                sql(
                    f"SELECT EXISTS(SELECT FROM pg_stat_activity a JOIN pg_locks l ON a.pid=l.pid WHERE a.application_name='{label}' AND l.locktype='advisory' AND l.granted)"
                )
                != "t"
            ):
                assert time.monotonic() < deadline, "Lost-reply fixture lock did not acquire"
                time.sleep(0.1)
            uncertain = pool.submit(command, "/publish", {}, 1, publish_id, 503)
            while sql("SELECT EXISTS(SELECT FROM pg_locks WHERE locktype='advisory' AND NOT granted)") != "t":
                assert time.monotonic() < deadline, "Command did not reach its owner"
                time.sleep(0.1)
            connections = json.loads(broker_request(values, "GET", "connections", {}))
            api_connections = [item for item in connections if item["user"] == "cafe_api"]
            assert len(api_connections) == 1, "Recovery requires one originating API connection"
            broker_request(values, "DELETE", "connections/" + quote(api_connections[0]["name"], safe=""), {})
            assert uncertain.result(timeout=15) == {"code": "temporarily_unavailable"}
        finally:
            sql(f"SELECT pg_terminate_backend(pid) FROM pg_stat_activity WHERE application_name='{label}'")
            try:
                held.result(timeout=5)
            except subprocess.CalledProcessError:
                pass
    deadline = time.monotonic() + 10
    pending = None
    while not pending:
        raw = sql(
            f"SELECT json_build_object('id',r.id,'body',encode(r.body,'hex'),'route',r.reply_route,'attempts',d.attempts) FROM cafe.command_replies r JOIN cafe.command_reply_dispatches d ON d.event_id=r.id WHERE d.completed_at IS NULL AND position(convert_to('{drink}','UTF8') in r.body)>0 ORDER BY r.created_at DESC LIMIT 1"
        )
        pending = json.loads(raw) if raw else None
        if not pending or pending["attempts"] == 0:
            pending = None
            assert time.monotonic() < deadline, "Committed reply did not survive its disconnected destination"
            time.sleep(0.1)
    assert pending["route"].startswith("reply.menu."), "Reply lost its originating connection"
    deadline = time.monotonic() + 5
    while True:
        queue = json.loads(broker_request(values, "GET", "queues/reference/ref.menu.commands", {}))
        if queue.get("messages_ready") == 0 and queue.get("messages_unacknowledged") == 0:
            break
        assert time.monotonic() < deadline, "Command acknowledgement waited for reply delivery"
        time.sleep(0.2)
    assert evidence() == {"version": 2, "receipts": 1, "events": 1, "publications": 2}
    from dev import wait_ready

    wait_ready(values, ("API",))
    recovered = command("/publish", {}, 1, publish_id, 200)
    assert recovered == {"aggregateId": drink, "version": 2, "status": "published"}
    assert command("/publish", {}, 1, publish_id, 200) == recovered
    deadline = time.monotonic() + 40
    while (
        sql(
            f"SELECT completed_at IS NOT NULL FROM cafe.command_reply_dispatches WHERE event_id='{pending['id']}'"
        )
        != "t"
    ):
        assert time.monotonic() < deadline, "Disconnected reply did not expire"
        time.sleep(0.1)
    assert (
        sql(f"SELECT last_error FROM cafe.command_reply_dispatches WHERE event_id='{pending['id']}'")
        == "expired"
    )
    assert sql(f"SELECT reply_route FROM cafe.command_replies WHERE id='{pending['id']}'") == pending["route"]
    assert (
        sql(f"SELECT encode(body,'hex') FROM cafe.command_replies WHERE id='{pending['id']}'")
        == pending["body"]
    ), "Recovered reply bytes changed"
    assert evidence() == {"version": 2, "receipts": 1, "events": 1, "publications": 2}, (
        "Reply recovery duplicated a committed transition"
    )
    print(
        "RabbitMQ reply loss preserves uncertainty; identical retries recover one committed outcome and one event"
    )


if __name__ == "__main__":
    main()
