"""Prove the source contracts bundle reproducibly and cannot escape their root."""
import json
from pathlib import Path
import tempfile
import unittest
from unittest.mock import patch
import http_contracts


class HTTPContractsTests(unittest.TestCase):
    def test_committed_bundles_match_authoritative_sources(self):
        http_contracts.generate(check=True)

    def test_generated_references_are_self_contained(self):
        def visit(value):
            if isinstance(value, dict):
                if "$ref" in value:
                    self.assertTrue(value["$ref"].startswith("#/components/"))
                for child in value.values():
                    visit(child)
            elif isinstance(value, list):
                for child in value:
                    visit(child)
        for path, content in http_contracts.generated().items():
            if path.suffix == ".json":
                visit(json.loads(content))

    def test_changed_bundle_fails_verification(self):
        with tempfile.TemporaryDirectory() as directory:
            path = Path(directory) / "api.openapi.json"
            path.write_text("outdated")
            with patch.object(http_contracts, "ROOT", Path(directory)), patch.object(
                http_contracts, "generated", return_value={path: "current"}
            ):
                with self.assertRaisesRegex(ValueError, "bundle drift"):
                    http_contracts.generate(check=True)

    def test_reference_outside_contract_root_is_rejected(self):
        with tempfile.TemporaryDirectory() as directory:
            root = Path(directory) / "http"
            root.mkdir()
            path = root / "api.openapi.json"
            path.write_text(json.dumps({"paths": {"/escape": {"$ref": "../private.json#/secret"}}}))
            with patch.object(http_contracts, "SOURCE", root):
                with self.assertRaisesRegex(ValueError, "stay inside contracts"):
                    http_contracts.bundle(path)

    def test_same_named_schemas_keep_each_owners_wire_shape(self):
        with tempfile.TemporaryDirectory() as directory:
            root = Path(directory)
            paths = {}
            for owner, field_type in (("menu", "string"), ("ordering", "integer")):
                source = root / "contexts" / owner / "schemas.json"
                source.parent.mkdir(parents=True)
                source.write_text(json.dumps({"schemas": {"State": {
                    "type": "object", "properties": {"value": {"type": field_type}}
                }}}))
                paths["/" + owner] = {"get": {"responses": {"200": {"content": {
                    "application/json": {"schema": {
                        "$ref": f"contexts/{owner}/schemas.json#/schemas/State"
                    }}
                }}}}}
            path = root / "api.openapi.json"
            path.write_text(json.dumps({"paths": paths}))
            with patch.object(http_contracts, "SOURCE", root):
                document = http_contracts.bundle(path)
            references = []
            for owner, field_type in (("menu", "string"), ("ordering", "integer")):
                reference = document["paths"]["/" + owner]["get"]["responses"]["200"]["content"]["application/json"]["schema"]["$ref"]
                references.append(reference)
                schema = document["components"]["schemas"][reference.rsplit("/", 1)[1]]
                self.assertEqual(schema["properties"]["value"]["type"], field_type)
            self.assertNotEqual(*references)

    def test_ambiguous_component_names_fail_instead_of_reusing_a_schema(self):
        with tempfile.TemporaryDirectory() as directory:
            root = Path(directory)
            paths = {}
            for name in ("contexts/menu/schemas.json", "contexts.menu.schemas.json"):
                source = root / name
                source.parent.mkdir(parents=True, exist_ok=True)
                source.write_text(json.dumps({"schemas": {"State": {"type": "string"}}}))
                paths["/" + name] = {"$ref": name + "#/schemas/State"}
            path = root / "api.openapi.json"
            path.write_text(json.dumps({"paths": paths}))
            with patch.object(http_contracts, "SOURCE", root):
                with self.assertRaisesRegex(ValueError, "component name collision"):
                    http_contracts.bundle(path)


if __name__ == "__main__":
    unittest.main()
