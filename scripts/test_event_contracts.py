"""New events require an owned schema and consumer declaration, with no codec edits."""

import hashlib
import json
from pathlib import Path
import shutil
import tempfile
import unittest
from contract_sources import ROOT, CONTEXT_PATHS, PRIVATE, catalogue
from event_contracts import load
from generate_events import generate
from generated_files import clear_location


class EventContractTests(unittest.TestCase):
    def setUp(self):
        directory = tempfile.TemporaryDirectory()
        self.addCleanup(directory.cleanup)
        self.root = Path(directory.name)
        shutil.copytree(ROOT / "contracts", self.root / "contracts")
        for directory in CONTEXT_PATHS.values():
            destination = self.root / directory
            destination.mkdir(parents=True)
            shutil.copyfile(ROOT / directory / "subscriptions.json", destination / "subscriptions.json")
        for directory in PRIVATE.values():
            shutil.copyfile(
                ROOT / directory / "private_messages.json", self.root / directory / "private_messages.json"
            )
        for directory in PRIVATE.values():
            source = ROOT / directory / "internal_commands.proto"
            if source.exists():
                shutil.copyfile(source, self.root / directory / source.name)
        self.schema = (
            self.root / "contracts/menu/messaging/integration_events/v1/menu_integration_events.proto"
        )
        load.cache_clear()

    def test_new_file_and_fields_generate_bindings_without_an_envelope_or_registry_edit(self):
        envelope = self.root / "contracts/shared/messaging/v1/events.proto"
        before = hashlib.sha256(envelope.read_bytes()).hexdigest()
        extra = self.schema.with_name("menu_archived.proto")
        extra.write_text("""syntax = "proto3";
package cafe.v1;
import "cafe/v1/event_options.proto";
message MenuArchived {
 option (cafe.events.event) = {name:"menu.edition-archived" owner:"menu" kind:"edition" version:1};
 string edition_id = 1 [(cafe.events.rules) = {uuid:true source_identity:true}];
 string reason = 2 [(cafe.events.rules) = {non_empty:true}];
}
""")
        generate(self.root)
        self.assertEqual(hashlib.sha256(envelope.read_bytes()).hexdigest(), before)
        model = (self.root / "services/storefront/contracts/events/generated/model/menu.go").read_text()
        self.assertIn("type MenuArchivedIntegrationEvent struct", model)
        self.assertRegex(model, r"Reason\s+string")
        codec = (self.root / "services/storefront/contracts/events/protobuf/menu_generated.go").read_text()
        self.assertIn("func PublishMenuArchived(", codec)
        registry = (
            self.root / "services/storefront/contracts/events/protobuf/registry_generated.go"
        ).read_text()
        self.assertIn('"menu.edition-archived"', registry)
        self.assertFalse(
            any(item["name"] == "menu.edition-archived" for item in catalogue("topology", self.root))
        )

    def test_subscriptions_allow_independent_consumers_and_keep_producers_consumer_free(self):
        path = self.root / CONTEXT_PATHS["ordering"] / "subscriptions.json"
        definitions = json.loads(path.read_text())
        definitions.append({"consumer": "ordering.menu-audit", "event": "menu.edition-published"})
        path.write_text(json.dumps(definitions))
        topology = catalogue("topology", self.root)
        self.assertEqual(
            {item["consumer"] for item in topology if item["name"] == "menu.edition-published"},
            {"ordering.menu-directory", "ordering.menu-audit"},
        )
        self.assertTrue(all("consumer" not in item for item in catalogue(root=self.root)))

    def test_publication_subscriptions_require_private_source_and_same_owner_and_kind(self):
        path = self.root / CONTEXT_PATHS["loyalty"] / "subscriptions.json"
        base = {
            "consumer": "loyalty.publish-reward-issued",
            "event": "loyalty.reward-issued.domain",
            "publication": "loyalty.reward-issued",
        }
        for changes in (
            {"command": "issueReward"},
            {"event": "loyalty.reward-issued"},
            {"event": "loyalty.reward-earned"},
            {"publication": "collection.order-collected"},
            {"publication": "loyalty.reward-issued.domain"},
            {"publication": "missing.event"},
        ):
            with self.subTest(changes=changes):
                path.write_text(json.dumps([{**base, **changes}]))
                with self.assertRaisesRegex(ValueError, "publication subscription"):
                    catalogue("topology", self.root)

    def test_explicit_dto_names_preserve_published_wire_identity(self):
        generate(self.root)
        python = (
            self.root / "services/operations/src/operations/contracts/events/generated/preparation.py"
        ).read_text()
        typescript = (self.root / "services/engagement/src/contracts/events/generated/loyalty.ts").read_text()
        self.assertIn("class DrinksReadyIntegrationEvent", python)
        self.assertIn("type RewardIssuedIntegrationEvent", typescript)
        event = next(e for e in load(self.root)["events"] if e["name"] == "preparation.drinks-ready")
        self.assertEqual(event["version"], 1)
        self.assertIn(
            "message DrinksReady {",
            (
                self.root
                / "contracts/preparation/messaging/integration_events/v1/preparation_integration_events.proto"
            ).read_text(),
        )

    def test_generated_outputs_have_clear_paths_and_prune_stale_producers(self):
        before = {p for p in self.root.rglob("*") if p.is_file()}
        generate(self.root)
        outputs = {p for p in self.root.rglob("*") if p.is_file()} - before
        self.assertTrue(outputs)
        self.assertTrue(all(clear_location(p.relative_to(self.root)) for p in outputs))
        for directory, suffix, expected in (
            ("services/operations/src/operations/contracts/events", ".py", "ordering.py"),
            ("services/engagement/src/contracts/events", ".ts", "collection.ts"),
        ):
            base = self.root / directory
            self.assertTrue((base / "generated" / expected).is_file())
            self.assertEqual(list(base.glob("*" + suffix)), [])
            (base / "generated" / ("obsolete" + suffix)).write_text("stale output\n")
        go = self.root / "services/storefront/contracts/events"
        self.assertTrue((go / "generated/model/menu.go").is_file())
        self.assertTrue((go / "generated/names/names.go").is_file())
        self.assertFalse((go / "model").exists())
        self.assertFalse((go / "names").exists())
        (go / "generated/model/obsolete.go").write_text("stale output\n")
        handwritten = go / "protobuf/codec.go"
        handwritten.write_text("package protobuf\n")
        generate(self.root)
        self.assertFalse(list(self.root.rglob("obsolete.*")))
        self.assertEqual(handwritten.read_text(), "package protobuf\n")

    def test_invalid_schema_declarations_fail_generation(self):
        source = self.schema.read_text()
        cases = (
            ("foreign owner", source.replace('owner: "menu"', 'owner: "ordering"')),
            ("incomplete event", source.replace("version: 1", "version: 0")),
            (
                "string constraint",
                source.replace("int64 minor = 5", "int64 minor = 5").replace("minimum: 0", "uuid: true"),
            ),
            (
                "unsupported",
                source.replace("string currency = 2;", "bytes currency = 2;").replace(
                    "string currency = 2 [", "bytes currency = 2 ["
                ),
            ),
            ("inverted numeric", source.replace("minimum: 0", "minimum: 5 maximum: 1")),
            (
                "multiple source",
                source.replace(
                    'string currency = 2 [(cafe.events.rules) = { pattern: "^[A-Z]{3}$" }]',
                    "string currency = 2 [(cafe.events.rules) = { uuid:true source_identity:true }]",
                ),
            ),
        )
        for expected, candidate in cases:
            with self.subTest(expected=expected):
                self.schema.write_text(candidate)
                load.cache_clear()
                # Compiler inspection happens in its pinned subprocess.
                with self.assertRaisesRegex(ValueError, expected):
                    load(self.root)

    def test_duplicate_names_and_unknown_or_foreign_private_subscriptions_fail(self):
        path = self.root / CONTEXT_PATHS["ordering"] / "subscriptions.json"
        for definitions in (
            [{"consumer": "ordering.typo", "event": "missing.event"}],
            [{"consumer": "ordering.private", "event": "menu.drink-published"}],
            [{"consumer": "menu.foreign", "event": "menu.edition-published"}],
            [{"consumer": "ordering.duplicate", "event": "menu.edition-published"}] * 2,
        ):
            path.write_text(json.dumps(definitions))
            with self.assertRaises(ValueError):
                catalogue("topology", self.root)


if __name__ == "__main__":
    unittest.main()
