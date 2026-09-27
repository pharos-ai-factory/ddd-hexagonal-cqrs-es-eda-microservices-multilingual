#!/usr/bin/env python3
"""Run real PostgreSQL/RabbitMQ evidence in an isolated disposable project."""
import argparse
import os
from pathlib import Path
import subprocess
import sys
from dev import ROOT, configure, compose, test_environment, up, wait_ready
from journey import run


def execute(env_file, keep=False):
    reports = ROOT / ".local/bdd"
    reports.mkdir(parents=True, exist_ok=True)
    values = configure(env_file, disposable=True)
    if not values["COMPOSE_PROJECT_NAME"].startswith("cafe-reference-test-"):
        raise RuntimeError("Integration fixtures require a disposable test project")
    passed = False
    previous_pause = os.environ.get("PAUSED_CONSUMERS")
    try:
        os.environ["PAUSED_CONSUMERS"] = "loyalty.issue-reward"
        up(env_file)
        # Foundation fixtures cannot compete with live application relays.
        compose(env_file, "stop", "storefront", "operations", "engagement")
        subprocess.run([sys.executable, "scripts/go.py", "test", "-race", "-tags", "integration", "-count=1", "-timeout=120s",
                        "./foundation/persistence/postgres", "./foundation/transport/amqp",
                        "./contexts/ordering/application"], cwd=ROOT,
                       env=dict(test_environment(values), BDD_REPORT_DIR=str(reports)), check=True)
        subprocess.run(["uv", "run", "--project", "services/operations", "pytest", "-q",
                        "services/operations/tests/persistence_integration.py",
                        "services/operations/tests/delivery_integration.py"],
                       cwd=ROOT, env=test_environment(values), check=True)
        subprocess.run(["pnpm", "--filter", "@cafe/engagement", "exec", "tsx", "--test",
                        "src/adaptors/postgres.integration.ts", "src/adaptors/broker.integration.ts"],
                       cwd=ROOT, env=test_environment(values), check=True)
        # Component fixtures deliberately contain partial roots. The journey
        # starts from empty business tables in this disposable project only.
        for owner in ("menu", "ordering", "preparation", "collection", "loyalty", "communication"):
            compose(env_file, "exec", "-T", "postgres", "psql", "-U", "postgres", "-d", "cafe_"+owner,
                    "-v", "ON_ERROR_STOP=1", "-c",
                    "TRUNCATE cafe.aggregates,cafe.command_receipts,cafe.consumer_receipts,cafe.projections CASCADE",
                    stdout=subprocess.DEVNULL)
        compose(env_file, "start", "storefront", "operations", "engagement")
        wait_ready(values)

        def recovery(stage, client, customer, order):
            if stage == "before_place":
                compose(env_file, "stop", "operations")
            elif stage == "after_place":
                current = client.request("STOREFRONT", "/v1/ordering/orders/"+order)
                assert current["state"]["status"] == "placed"
                compose(env_file, "start", "operations")
                wait_ready(values, ("OPERATIONS",))
            elif stage == "after_earning":
                rewards = client.request("ENGAGEMENT", "/v1/loyalty/rewards")
                assert not any(item["state"]["customerId"] == customer for item in rewards), "Reward was issued while its private consumer was paused"
                print("LoyaltyAccount earned its grant while Reward issuance was paused")
                os.environ["PAUSED_CONSUMERS"] = ""
                compose(env_file, "up", "-d", "--no-deps", "engagement")
                wait_ready(values, ("ENGAGEMENT",))

        run(values, recovery=recovery)
        subprocess.run(["node", "--import", "tsx", "node_modules/@cucumber/cucumber/bin/cucumber.js",
                        "--config", "tests/acceptance/cucumber.mjs"], cwd=ROOT,
                       env=dict(os.environ, CAFE_ENV_FILE=str(env_file.resolve())), check=True)
        subprocess.run(["node", "scripts/check_bdd_reports.mjs", "infrastructure"], cwd=ROOT, check=True)
        subprocess.run([sys.executable, "scripts/browser.py"], cwd=ROOT,
                       env=dict(os.environ, CAFE_ENV_FILE=str(env_file)), check=True)
        subprocess.run([sys.executable, "tests/infrastructure/valkey_permissions.py"], cwd=ROOT,
                       env=dict(os.environ, CAFE_ENV_FILE=str(env_file)), check=True)
        passed = True
        print("PostgreSQL, RabbitMQ, Gherkin workflows, provider recovery and browser journey passed")
    finally:
        if previous_pause is None:
            os.environ.pop("PAUSED_CONSUMERS", None)
        else:
            os.environ["PAUSED_CONSUMERS"] = previous_pause
        if not passed:
            compose(env_file, "logs", "--tail", "30", "storefront", "operations", "engagement")
        if not keep:
            compose(env_file, "down", "--volumes", "--remove-orphans")


if __name__ == "__main__":
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("--env-file", type=Path, default=ROOT/".local/integration.env")
    parser.add_argument("--keep", action="store_true")
    args = parser.parse_args()
    execute(args.env_file, args.keep)
