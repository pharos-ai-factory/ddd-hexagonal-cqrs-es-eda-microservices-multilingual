"""Published source inventory excludes owner-private delivery implementation."""
import re
from pathlib import Path
import unittest
from contract_sources import ROOT, PRIVATE, catalogue, sources
from request_contracts import OWNED, envelopes


class ContractSourcesTests(unittest.TestCase):
    def test_published_sources_live_in_contracts_and_have_no_private_payload(self):
        for kind in ("events", "requests", "realtime"):
            for path in sources(kind).values():
                self.assertTrue(path.is_relative_to(ROOT/"contracts"))
                self.assertTrue(path.exists())
        self.assertFalse(any("domain_events" in path.name for path in sources("events").values()))
        envelope = (ROOT/"contracts/shared/messaging/v1/events.proto").read_text()
        self.assertNotIn("_domain_events.proto", envelope)
        self.assertNotIn("DrinkPublished drink_published", envelope)
        self.assertNotIn("RewardEarned reward_earned", envelope)
        self.assertNotIn("NotificationRequested notification_requested", envelope)

    def test_private_metadata_is_added_only_to_its_runtime(self):
        self.assertTrue(all(item["visibility"] == "integration" for item in catalogue()))
        for service, owners in (("storefront", {"menu"}), ("operations", set()),
                               ("engagement", {"loyalty", "communication"})):
            private = [item for item in catalogue(service) if item["visibility"] == "domain"]
            self.assertEqual({item["owner"] for item in private}, owners)
            self.assertTrue(all(item["consumer"].startswith(item["owner"]+".") for item in private))
        for owner, directory in PRIVATE.items():
            self.assertEqual(list(sources(owner+"_private").values()), [ROOT/directory/"private_messages.proto"])
            self.assertFalse(Path(directory).is_relative_to(Path("contracts")))

    def test_request_packages_are_owned_and_independently_importable(self):
        self.assertFalse((ROOT/"contracts/shared/messaging/v1/requests.proto").exists())
        for owner in OWNED["api"]:
            logical = f"cafe/requests/v1/contexts/{owner}/{owner}_requests.proto"
            envelope = sources("requests")[logical].read_text()
            self.assertIn(f"package cafe.{owner}.requests.v1;", envelope)
            for imported in re.findall(r'import "([^"]+)";', envelope):
                self.assertTrue("/contexts/" not in imported or f"/contexts/{owner}/" in imported)
        for service, owners in OWNED.items():
            derived = envelopes(owners, "test")
            for owner in OWNED["api"]:
                self.assertEqual(f'/contexts/{owner}"' in derived, owner in owners)
                if service not in ("api", "engagement") and owner not in owners:
                    base = ROOT/("services/storefront/contracts/requests/generated/cafe/requests/v1/contexts" if service == "storefront" else "services/operations/src/operations/adaptors/generated/cafe/requests/v1/contexts")
                    self.assertFalse(list((base/owner).glob("*.go" if service == "storefront" else "*_pb2.*")))


if __name__ == "__main__":
    unittest.main()
