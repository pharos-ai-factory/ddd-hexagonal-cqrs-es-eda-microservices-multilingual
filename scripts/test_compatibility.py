import copy
import unittest
from compatibility import proto_changes
from http_compatibility import changes, schema_changes
from http_retirements import compatible_changes, RETIRED_STOREFRONT_OPERATIONS


def document():
    return {
        "paths": {
            "/orders": {
                "post": {
                    "operationId": "create",
                    "requestBody": {
                        "content": {
                            "application/json": {
                                "schema": {"type": "object", "properties": {"name": {"type": "string"}}}
                            }
                        }
                    },
                    "responses": {
                        "200": {
                            "content": {
                                "application/json": {
                                    "schema": {
                                        "type": "object",
                                        "required": ["id"],
                                        "properties": {"id": {"type": "string"}},
                                    }
                                }
                            }
                        }
                    },
                }
            }
        }
    }


class CompatibilityTests(unittest.TestCase):
    def test_only_the_explicitly_retired_owner_operations_are_exempt(self):
        self.assertEqual(len(RETIRED_STOREFRONT_OPERATIONS), 17)
        for operation in RETIRED_STOREFRONT_OPERATIONS:
            method, path = operation.split(" ", 1)
            old = {"paths": {path: {method.lower(): {"responses": {}}}}}
            removed = {"paths": {}}
            self.assertEqual(compatible_changes("storefront", old, removed), [])
            self.assertTrue(compatible_changes("api", old, removed))
            changed = copy.deepcopy(old)
            changed["paths"][path][method.lower()]["security"] = []
            self.assertTrue(compatible_changes("storefront", old, changed))
        for path in ("/healthz", "/diagnostics", "/future-operation"):
            old = {"paths": {path: {"get": {"responses": {}}}}}
            self.assertTrue(compatible_changes("storefront", old, {"paths": {}}))

    def test_required_request_addition_fails_but_optional_addition_passes(self):
        old, new = document(), document()
        schema = new["paths"]["/orders"]["post"]["requestBody"]["content"]["application/json"]["schema"]
        schema["properties"]["note"] = {"type": "string"}
        self.assertEqual(changes(old, new), [])
        schema["required"] = ["note"]
        self.assertTrue(changes(old, new))

    def test_removed_response_guarantee_and_operation_fail(self):
        old, new = document(), document()
        new["paths"]["/orders"]["post"]["responses"]["200"]["content"]["application/json"]["schema"][
            "required"
        ] = []
        self.assertTrue(changes(old, new))
        self.assertTrue(changes(old, {"paths": {}}))

    def test_request_and_response_types_and_required_headers_fail(self):
        for category, entry in [("requestBody", None), ("responses", "200")]:
            old, new = document(), document()
            body = new["paths"]["/orders"]["post"][category]
            if entry:
                body = body[entry]
            body["content"]["application/json"]["schema"]["type"] = "array"
            self.assertTrue(changes(old, new))
        old, new = document(), document()
        new["paths"]["/orders"]["post"]["parameters"] = [
            {"in": "header", "name": "x-new", "required": True, "schema": {"type": "string"}}
        ]
        self.assertTrue(changes(old, new))

    def test_protobuf_field_reuse_removal_and_new_required_semantics_fail(self):
        old = {"cafe.Example": {"fields": {"1": {"name": "id", "options": ""}}}}
        for new in (
            {},
            {"cafe.Example": {"fields": {}}},
            {"cafe.Example": {"fields": {"1": {"name": "other", "options": ""}}}},
        ):
            self.assertTrue(proto_changes(old, new))
        new = copy.deepcopy(old)
        new["cafe.Example"]["fields"]["2"] = {"name": "optional", "options": ""}
        self.assertEqual(proto_changes(old, new), [])
        new["cafe.Example"]["fields"]["2"]["options"] = "required"
        self.assertTrue(proto_changes(old, new))


class PaginatedResponseCompatibilityTests(unittest.TestCase):
    def schema(self):
        item = {
            "type": "object",
            "required": ["id", "status"],
            "properties": {
                "id": {"type": "string"},
                "status": {"type": "string", "enum": ["draft", "placed"]},
            },
        }
        return {
            "oneOf": [
                {"type": "array", "items": copy.deepcopy(item)},
                {
                    "type": "object",
                    "required": ["items"],
                    "properties": {"items": {"type": "array", "items": copy.deepcopy(item)}},
                },
            ]
        }

    def test_optional_response_field_is_checked_inside_both_list_shapes(self):
        old, new = self.schema(), self.schema()
        for item in (new["oneOf"][0]["items"], new["oneOf"][1]["properties"]["items"]["items"]):
            item["properties"]["collectionRecorded"] = {"type": "boolean"}
        self.assertEqual(schema_changes(old, new, response=True), [])

    def test_incompatible_fields_in_either_branch_still_fail(self):
        for index in (0, 1):
            for mutation in ("removed", "enum", "type", "guarantee"):
                with self.subTest(index=index, mutation=mutation):
                    old, new = self.schema(), self.schema()
                    item = (
                        new["oneOf"][0]["items"]
                        if index == 0
                        else new["oneOf"][1]["properties"]["items"]["items"]
                    )
                    if mutation == "removed":
                        del item["properties"]["id"]
                    elif mutation == "enum":
                        item["properties"]["status"]["enum"].append("collected")
                    elif mutation == "type":
                        item["properties"]["id"]["type"] = "integer"
                    else:
                        item["required"] = []
                    self.assertTrue(schema_changes(old, new, response=True))

    def test_other_union_changes_remain_conservative(self):
        old, new = self.schema(), self.schema()
        new["oneOf"].append({"type": "string"})
        self.assertTrue(schema_changes(old, new, response=True))
        old = {"oneOf": [{"type": "object", "required": ["a"]}, {"type": "object", "required": ["b"]}]}
        new = copy.deepcopy(old)
        new["oneOf"][0]["properties"] = {"optional": {"type": "boolean"}}
        self.assertTrue(schema_changes(old, new, response=True))
