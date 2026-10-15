"""Protect response schema authority, evolution and generated query signatures."""

import unittest
import test_request_payloads as inputs
from generate_response_payloads import generate
from response_contract_policy import violation


class ResponsePayloadTests(unittest.TestCase):
    setUp = inputs.RequestPayloadTests.setUp
    schema = inputs.RequestPayloadTests.schema

    def test_every_request_declares_its_response(self):
        data = self.schema()
        requests = {r["name"]: r for r in data["payloads"]}
        self.assertEqual(requests["PlaceOrderCommand"]["response"], "cafe.requests.v1.Outcome")
        self.assertEqual(requests["RecordCollectionCommand"]["response"], "cafe.requests.v1.Outcome")
        self.assertEqual(requests["GetOrderQuery"]["response_field"], "order")
        self.assertEqual(requests["ListOrdersQuery"]["response_field"], "orders")
        names = {m["name"] for m in data["responses"]}
        self.assertTrue({"Outcome", "Rejection", "RequestError", "Selection", "Orders"} <= names)

    def test_generated_response_packages_have_one_location(self):
        generate(self.root)
        base = self.root / "services/storefront"
        for owner in ("shared", "menu", "ordering"):
            self.assertTrue((base / f"contracts/requests/generated/model/{owner}/responses.go").is_file())
        self.assertTrue((base / "adaptors/messaging/generated/responses/responses.go").is_file())
        for owner in ("menu", "ordering"):
            messaging = base / f"contexts/{owner}/adaptors/messaging"
            self.assertTrue((messaging / "generated/queryregistration/responses.go").is_file())
            self.assertTrue((messaging / "request_responses_generated.go").is_file())
            self.assertFalse((messaging / "queryregistration").exists())
        self.assertFalse((base / "contracts/requests/model").exists())
        self.assertFalse((base / "adaptors/messaging/responses").exists())

    def test_missing_foreign_and_wrong_response_associations_fail(self):
        path = self.root / "contracts/ordering/messaging/queries/v1/ordering_queries.proto"
        original = path.read_text()
        for replacement in ["", "cafe.menu.requests.v1.LoadedDrink", "cafe.ordering.requests.v1.Absent"]:
            with self.subTest(response=replacement):
                path.write_text(original.replace("cafe.ordering.requests.v1.LoadedOrder", replacement))
                with self.assertRaisesRegex(ValueError, "response"):
                    self.schema()
        path.write_text(original)
        path = self.root / "contracts/ordering/messaging/commands/v1/ordering_commands.proto"
        path.write_text(
            path.read_text().replace("cafe.requests.v1.Outcome", "cafe.ordering.requests.v1.LoadedOrder", 1)
        )
        with self.assertRaisesRegex(ValueError, "executor requires Outcome"):
            self.schema()

    def test_added_nested_repeated_and_optional_fields_reach_models_and_encoders(self):
        for owner, name in [("ordering", "Order"), ("preparation", "Ticket"), ("loyalty", "Reward")]:
            path = self.root / f"contracts/{owner}/messaging/queries/v1/{owner}_replies.proto"
            text = path.read_text().replace(
                "message " + name + " {",
                "message " + name + " {\n repeated ReplyDetail details = 98;\n optional bool reviewed = 99;",
            )
            path.write_text(text + "\nmessage ReplyDetail { string note = 1; }\n")
        self.schema()
        generate(self.root)
        cases = [
            (
                "services/storefront/contracts/requests/generated/model/ordering/responses.go",
                "Reviewed *bool",
            ),
            (
                "services/storefront/contexts/ordering/adaptors/messaging/request_responses_generated.go",
                "ToReplyDetail(item)",
            ),
            (
                "services/operations/src/operations/contracts/requests/generated/preparation_responses.py",
                "details: list[ReplyDetail]",
            ),
            (
                "services/operations/src/operations/contexts/preparation/adaptors/messaging/generated/response_payloads.py",
                "to_reply_detail(item)",
            ),
            (
                "services/engagement/src/contracts/requests/generated/loyalty-responses.ts",
                "reviewed?: boolean",
            ),
            (
                "services/engagement/src/contexts/loyalty/adaptors/messaging/generated/response-payloads.ts",
                "toReplyDetail(item)",
            ),
        ]
        for file, expected in cases:
            text = (self.root / file).read_text()
            self.assertIn(" ".join(expected.split()), " ".join(text.split()))
        self.assertIn(
            "Promise<GetRewardResponse>",
            (
                self.root
                / "services/engagement/src/contexts/loyalty/adaptors/messaging/generated/response-payloads.ts"
            ).read_text(),
        )

    def test_unsupported_response_shapes_fail_before_generation(self):
        path = self.root / "contracts/ordering/messaging/queries/v1/ordering_replies.proto"
        original = path.read_text()
        for extra in ["map<string, string> metadata = 99;", "oneof result { string a = 98; string b = 99; }"]:
            path.write_text(original.replace("message Order {", "message Order { " + extra))
            with self.assertRaisesRegex(ValueError, "require a generator extension"):
                self.schema()

    def test_policy_rejects_manual_response_copies_and_unbound_composition(self):
        for path, text in [
            (
                "services/storefront/contexts/ordering/application/readmodels/order.go",
                "type OrderView struct { ID string }",
            ),
            (
                "services/operations/src/operations/foundation/application.py",
                "class Outcome(TypedDict):\n    status: str",
            ),
            (
                "services/engagement/src/contexts/loyalty/application/queries/get.ts",
                "export type GetResponse = { status: string };",
            ),
            (
                "services/engagement/src/apps/composition/loyalty.ts",
                "registry.query(name, parse, handler, reply);",
            ),
        ]:
            self.assertIsNotNone(violation(path, text))
        self.assertIsNone(
            violation(
                "services/storefront/contexts/ordering/application/readmodels/order.go",
                "type OrderView = model.Order",
            )
        )
        self.assertIsNone(
            violation(
                "services/engagement/src/apps/composition/loyalty.ts",
                "responses.registerGetReward(registry, parse, handler);",
            )
        )

    def test_go_generated_binding_rejects_another_response(self):
        import subprocess
        from contract_sources import ROOT

        source = self.root / "binding_test.go"
        module = "github.com/pharos-ai-factory/ddd-hexagonal-cqrs-es-eda-microservices-multilingual/services/storefront"
        source.write_text(f'''package binding
import (
 "context"
 rpc "{module}/contexts/ordering/adaptors/messaging"
 responses "{module}/contexts/ordering/adaptors/messaging/generated/queryregistration"
 model "{module}/contracts/requests/generated/model/ordering"
 requests "{module}/adaptors/messaging/requests"
)
func bind(registry *requests.RabbitMQRequestRegistry) {{
 responses.RegisterGetOrder(registry, rpc.GetOrder, func(context.Context, model.GetOrderQuery) (model.GetOrderResponse, error) {{ return model.GetOrderResponse{{}}, nil }})
}}
''')
        args = ["python3", "scripts/go.py", "test", str(source)]
        result = subprocess.run(args, cwd=ROOT, capture_output=True, text=True)
        self.assertEqual(result.returncode, 0, result.stdout + result.stderr)
        source.write_text(source.read_text().replace("model.GetOrderResponse", "model.ListOrdersResponse"))
        result = subprocess.run(args, cwd=ROOT, capture_output=True, text=True)
        self.assertNotEqual(result.returncode, 0)
        self.assertIn("cannot use", result.stdout + result.stderr)
        self.assertIn("RegisterGetOrder", result.stdout + result.stderr)
