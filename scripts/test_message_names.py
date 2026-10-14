"""Protect generated event/subscription scope and private command source authority."""

import json
from pathlib import Path
import shutil
import tempfile
import unittest

from contract_sources import ROOT, PRIVATE
from generate_message_names import message_names, render


class MessageNameTests(unittest.TestCase):
    def setUp(self):
        directory = tempfile.TemporaryDirectory()
        self.addCleanup(directory.cleanup)
        self.root = Path(directory.name)
        shutil.copytree(ROOT / "contracts", self.root / "contracts")
        for path in PRIVATE.values():
            destination = self.root / path
            destination.mkdir(parents=True)
            for name in ("private_messages.json", "subscriptions.json", "internal_commands.proto"):
                source = ROOT / path / name
                if source.exists():
                    shutil.copyfile(source, destination / name)

    def test_generated_roles_retain_private_scope_and_wire_values(self):
        catalogue = message_names("operations", self.root)
        self.assertEqual(catalogue["preparation"]["PrivateCommand"], [("ACCEPT_ORDER", "accept_order")])
        self.assertIn(
            ("DRINKS_READY", "preparation.drinks-ready.domain"), catalogue["preparation"]["DomainEvent"]
        )
        self.assertIn(("ORDER_PLACED", "ordering.order-placed"), catalogue["ordering"]["IntegrationEvent"])
        self.assertNotIn("DomainEvent", catalogue["ordering"])
        self.assertNotIn("Subscription", catalogue["ordering"])
        self.assertNotIn("loyalty", catalogue)
        code = render(catalogue, "python")
        self.assertIn("class PreparationSubscription(StrEnum)", code)
        self.assertIn("PUBLISH_DRINKS_READY = 'preparation.publish-drinks-ready'", code)

    def test_new_subscription_is_derived_from_the_owner_manifest(self):
        path = self.root / PRIVATE["preparation"] / "subscriptions.json"
        definitions = json.loads(path.read_text())
        definitions.append(
            {
                "consumer": "preparation.audit-ready",
                "event": "preparation.drinks-ready.domain",
                "publication": "preparation.drinks-ready",
            }
        )
        path.write_text(json.dumps(definitions))
        catalogue = message_names("operations", self.root)
        self.assertIn(("AUDIT_READY", "preparation.audit-ready"), catalogue["preparation"]["Subscription"])

    def test_private_command_manifest_must_match_the_compiled_envelope(self):
        path = self.root / PRIVATE["preparation"] / "internal_commands.proto"
        path.write_text(path.read_text().replace(" accept_order =", " renamed_accept ="))
        with self.assertRaisesRegex(ValueError, "unknown private command: preparation"):
            message_names("operations", self.root)
