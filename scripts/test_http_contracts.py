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
        for content in http_contracts.generated().values():
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
                with self.assertRaisesRegex(ValueError, "stay inside contracts/http"):
                    http_contracts.bundle(path)


if __name__ == "__main__":
    unittest.main()
