#!/usr/bin/env python3
"""Run the same HTTP behaviour contract against any language implementation."""
import argparse
import json
from pathlib import Path
import time
import urllib.error
import urllib.request
import uuid
from dev import ROOT, load

class Client:
    def __init__(self, values):
        self.values = values
        self.correlation = str(uuid.uuid4())

    def request(self, service, path, body=None, version=None, key=None, allowed=(200,)):
        headers = {"Authorization": "Bearer "+self.values["API_KEY"],
                   "X-Correlation-ID": self.correlation}
        if body is not None:
            headers.update({"Content-Type": "application/json", "Idempotency-Key": key or str(uuid.uuid4()),
                            "If-Match": str(version)})
        request = urllib.request.Request(f'http://127.0.0.1:{self.values["API_PORT"]}/api{path}',
                                        data=None if body is None else json.dumps(body).encode(), headers=headers)
        try:
            response = urllib.request.urlopen(request, timeout=10)
        except urllib.error.HTTPError as error:
            response = error
        with response:
            result = json.load(response)
            if response.status not in allowed:
                raise AssertionError(f"{path}: HTTP {response.status}: {result}")
            return result

    def await_command(self, service, path, body, version):
        deadline = time.monotonic()+30
        while True:
            result = self.request(service, path, body, version, allowed=(200, 422))
            rejection = result.get("rejection")
            if not rejection:
                return result
            if rejection["code"] not in ("menu_pending", "drink_revision_pending") or time.monotonic()>deadline:
                raise AssertionError(result)
            # A rejected command remains rejected. A new attempt uses a new identity.
            time.sleep(0.1)

    def find(self, service, resource, predicate, timeout=30):
        deadline = time.monotonic()+timeout
        while True:
            for item in self.request(service, resource):
                if predicate(item["state"]):
                    return item
            if time.monotonic()>deadline:
                raise AssertionError(f"No matching state arrived at {resource}")
            time.sleep(0.1)

    def provider(self, path, body=None):
        request = urllib.request.Request(f'http://127.0.0.1:{self.values["DELIVERY_PORT"]}{path}',
                                        data=None if body is None else json.dumps(body).encode(),
                                        headers={"Authorization": "Bearer "+self.values["DELIVERY_KEY"], "Content-Type": "application/json"})
        with urllib.request.urlopen(request, timeout=5) as response:
            return json.load(response)

def run(values, recovery=None):
    client = Client(values)
    for service, resource in (("STOREFRONT", "/v1/ordering/orders"),
                              ("OPERATIONS", "/v1/preparation/tickets"),
                              ("ENGAGEMENT", "/v1/loyalty/rewards")):
        assert client.request(service, resource+"/invalid", allowed=(400,))["code"] == "invalid_id"
    customer, drink, edition = (str(uuid.uuid4()) for _ in range(3))
    prefix = "/v1/menu"
    client.request("STOREFRONT", f"{prefix}/drinks/{drink}", {"name": "Cappuccino"}, 0)
    client.request("STOREFRONT", f"{prefix}/drinks/{drink}/publish", {}, 1)
    client.request("STOREFRONT", f"{prefix}/editions/{edition}", {"currency": "EUR"}, 0)
    client.await_command("STOREFRONT", f"{prefix}/editions/{edition}/offers",
                         {"code": "C1", "drinkId": drink, "drinkRevision": 1, "minor": 300}, 1)
    client.request("STOREFRONT", f"{prefix}/editions/{edition}/publish", {}, 2)
    rejected = client.request("STOREFRONT", f"{prefix}/editions/{edition}/prices",
                              {"code": "C1", "minor": 350}, 3, allowed=(422,))
    assert rejected["rejection"]["code"] == "edition_already_published"
    # Acceptance followed by a lost response must still produce one provider effect.
    client.provider("/failures", {"lostResponses": 1})
    order_ids = []
    for number in range(3):
        order, line = str(uuid.uuid4()), str(uuid.uuid4())
        order_ids.append(order)
        base = f"/v1/ordering/orders/{order}"
        client.await_command("STOREFRONT", base, {"customerId": customer, "editionId": edition}, 0)
        client.request("STOREFRONT", base+"/lines", {"lineId": line, "editionId": edition, "offerCode": "C1", "quantity": 2}, 1)
        command_id = str(uuid.uuid4())
        if recovery and number == 0:
            recovery("before_place", client, customer, order)
        first = client.request("STOREFRONT", base+"/place", {}, 2, command_id)
        repeated = client.request("STOREFRONT", base+"/place", {}, 2, command_id)
        assert repeated == first, "Command replay changed the recorded outcome"
        if recovery and number == 0:
            recovery("after_place", client, customer, order)
        ticket = client.find("OPERATIONS", "/v1/preparation/tickets", lambda state: state["orderId"] == order)
        ticket_path = "/v1/preparation/tickets/"+ticket["state"]["id"]
        client.request("OPERATIONS", ticket_path+"/start", {}, ticket["version"])
        client.request("OPERATIONS", ticket_path+"/complete", {}, ticket["version"]+1)
        pickup = client.find("OPERATIONS", "/v1/collection/pickups", lambda state: state["orderId"] == order)
        pickup_path = "/v1/collection/pickups/"+pickup["state"]["id"]
        wrong = client.request("OPERATIONS", pickup_path+"/collect", {"code": "WRONG1"}, pickup["version"], allowed=(422,))
        assert wrong["rejection"]["code"] == "incorrect_collection_code"
        client.request("OPERATIONS", pickup_path+"/collect", {"code": pickup["state"]["code"]}, pickup["version"])
    account = client.find("ENGAGEMENT", "/v1/loyalty/accounts", lambda state: state["id"] == customer and state["collections"] == 3)
    assert account["state"]["stampBalance"] == 0 and account["state"]["grantsEarned"] == 1
    if recovery:
        recovery("after_earning", client, customer, order_ids[-1])
    reward = client.find("ENGAGEMENT", "/v1/loyalty/rewards", lambda state: state["customerId"] == customer)
    assert reward["state"]["grantId"] == account["state"]["lastGrant"]["id"]
    deadline = time.monotonic()+30
    while True:
        notices = [item for item in client.request("ENGAGEMENT", "/v1/communication/notifications")
                   if item["state"]["recipient"] == customer]
        if len(notices) == 4 and all(item["state"]["status"] == "sent" for item in notices):
            break
        if time.monotonic()>deadline:
            raise AssertionError(f"Expected four delivered notices: {notices}")
        time.sleep(0.1)
    accepted = [item for item in client.provider("/messages") if item["message"]["recipient"] == customer]
    assert len(accepted) == 4, "A lost acceptance response produced a duplicate provider effect"
    result = {"customerId": customer, "editionId": edition, "orderIds": order_ids,
              "rewardId": reward["state"]["id"], "notifications": len(accepted), "correlationId": client.correlation}
    print(json.dumps(result, indent=2))
    return result

if __name__ == "__main__":
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("--env-file", type=Path, default=ROOT/".local/dev.env")
    run(load(parser.parse_args().env_file))
