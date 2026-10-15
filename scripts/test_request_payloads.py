"""Protect schema authority, input presence and public/private delivery scope."""

import json
from pathlib import Path
import shutil
import tempfile
import unittest
from contract_sources import ROOT, PRIVATE
from request_payloads import load
from generate_request_go import generate as go
from generate_request_python import generate as python
from generate_request_typescript import generate as typescript
from broker_permissions import build


class RequestPayloadTests(unittest.TestCase):
    def setUp(self):
        directory = tempfile.TemporaryDirectory()
        self.addCleanup(directory.cleanup)
        self.root = Path(directory.name)
        shutil.copytree(ROOT / "contracts", self.root / "contracts")
        for directory in PRIVATE.values():
            target = self.root / directory
            target.mkdir(parents=True)
            for name in ("internal_commands.proto", "private_messages.json", "subscriptions.json"):
                source = ROOT / directory / name
                if source.exists():
                    shutil.copyfile(source, target / name)

    def schema(self):
        load.cache_clear()
        return load(self.root)

    def test_public_and_private_commands_and_queries_share_schema_authority(self):
        payloads = self.schema()["payloads"]
        by_name = {m["name"]: m for m in payloads}
        self.assertEqual(by_name["PlaceOrderCommand"]["visibility"], "PUBLIC")
        self.assertEqual(by_name["RecordCollectionCommand"]["visibility"], "PRIVATE")
        self.assertEqual(by_name["GetOrderQuery"]["role"], "Query")
        quantity = by_name["ChangeQuantityCommand"]["fields"][1]
        self.assertTrue(quantity["presence"])
        self.assertTrue(quantity["required"])
        self.assertFalse(quantity["optional"])
        self.assertTrue(by_name["ListOrdersQuery"]["fields"][0]["optional"])

    def test_schema_additions_generate_payload_and_conversion_in_every_language(self):
        cases = [
            ("ordering", "ordering_commands.proto", "PlaceOrder"),
            ("preparation", "preparation_commands.proto", "CompletePreparation"),
            ("loyalty", "loyalty_commands.proto", "RedeemReward"),
        ]
        for owner, filename, name in cases:
            p = self.root / f"contracts/{owner}/messaging/commands/v1/{filename}"
            p.write_text(
                p.read_text().replace(
                    "message " + name + " {", "message " + name + " {\n optional int32 attempt = 99;"
                )
            )
        data = self.schema()
        for owner, render in [("ordering", go), ("preparation", python), ("loyalty", typescript)]:
            render(self.root, owner, [m for m in data["payloads"] if m["owner"] == owner])
        plain = (
            self.root / "services/storefront/contracts/requests/generated/model/ordering/payloads.go"
        ).read_text()
        self.assertIn("Attempt *int", plain)
        self.assertNotIn("protobuf", plain)
        go_codec = (self.root / PRIVATE["ordering"] / "request_payloads_generated.go").read_text()
        self.assertIn("if value.Attempt != nil", go_codec)
        py = (
            self.root / "services/operations/src/operations/contracts/requests/generated/preparation.py"
        ).read_text()
        self.assertIn("attempt: NotRequired[int]", py)
        ts = (self.root / "services/engagement/src/contracts/requests/generated/loyalty.ts").read_text()
        self.assertIn("attempt?: number | undefined", ts)
        self.assertIn(
            "parseGetRewardQuery",
            (self.root / PRIVATE["loyalty"] / "generated/request-payloads.ts").read_text(),
        )

    def test_query_fields_are_generated_without_a_handwritten_input(self):
        p = self.root / "contracts/preparation/messaging/queries/v1/preparation_queries.proto"
        p.write_text(
            p.read_text().replace(
                "message GetTicket {", "message GetTicket {\n optional bool include_history = 99;"
            )
        )
        data = self.schema()
        python(self.root, "preparation", [m for m in data["payloads"] if m["owner"] == "preparation"])
        model = (
            self.root / "services/operations/src/operations/contracts/requests/generated/preparation.py"
        ).read_text()
        self.assertIn("include_history: bool | None = None", model)
        parser = (self.root / PRIVATE["preparation"] / "generated/request_payloads.py").read_text()
        self.assertIn('value.HasField("include_history")', parser)

    def test_owner_and_visibility_must_agree_with_the_delivery_protocol(self):
        path = self.root / "contracts/ordering/messaging/commands/v1/ordering_commands.proto"
        original = path.read_text()
        for old, new in [
            ("visibility: PUBLIC", "visibility: PRIVATE"),
            ('owner: "ordering"', 'owner: "menu"'),
            ("visibility: PUBLIC", "visibility: VISIBILITY_UNSPECIFIED"),
        ]:
            with self.subTest(new=new):
                path.write_text(original.replace(old, new, 1))
                with self.assertRaisesRegex(ValueError, "owner/visibility"):
                    self.schema()
        path.write_text(
            original.replace(
                'option (cafe.requests.v1.contract) = { owner: "ordering" visibility: PUBLIC response: "cafe.requests.v1.Outcome" };',
                "",
                1,
            )
        )
        with self.assertRaisesRegex(ValueError, "must declare"):
            self.schema()

    def test_unbound_and_unsupported_payload_shapes_fail_generation(self):
        path = self.root / "contracts/ordering/messaging/commands/v1/ordering_commands.proto"
        original = path.read_text()
        path.write_text(
            original
            + '\nmessage Abandoned { option (cafe.requests.v1.contract) = { owner: "ordering" visibility: PUBLIC }; }\n'
        )
        with self.assertRaisesRegex(ValueError, "delivery envelope"):
            self.schema()
        path.write_text(
            original.replace("message PlaceOrder {", "message PlaceOrder { repeated string tags = 99;")
        )
        with self.assertRaisesRegex(ValueError, "unsupported request field"):
            self.schema()

    def test_user_oneofs_cannot_masquerade_as_optional_presence(self):
        path = self.root / "contracts/ordering/messaging/commands/v1/ordering_commands.proto"
        path.write_text(
            path.read_text().replace(
                "message PlaceOrder {", "message PlaceOrder { oneof _choice { string code = 99; }"
            )
        )
        with self.assertRaisesRegex(ValueError, "oneofs require"):
            self.schema()

    def test_private_visibility_is_checked_before_broker_permissions_are_generated(self):
        path = self.root / PRIVATE["ordering"] / "internal_commands.proto"
        path.write_text(path.read_text().replace("visibility: PRIVATE", "visibility: PUBLIC"))
        load.cache_clear()
        with self.assertRaisesRegex(ValueError, "owner/visibility"):
            build(self.root)

    def test_subscription_must_select_an_owner_private_command(self):
        path = self.root / PRIVATE["ordering"] / "subscriptions.json"
        subscriptions = json.loads(path.read_text())
        next(s for s in subscriptions if "command" in s)["command"] = "placeOrder"
        path.write_text(json.dumps(subscriptions))
        with self.assertRaisesRegex(ValueError, "owner-private"):
            build(self.root)
