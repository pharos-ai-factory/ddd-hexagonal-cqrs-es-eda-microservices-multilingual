import unittest
from pathlib import Path
from types import SimpleNamespace
from unittest.mock import patch
import check_architecture
from check_architecture import MODULE, import_violation, core_violation

class ArchitectureTests(unittest.TestCase):
    def test_domain_cannot_import_sql_even_through_an_alias(self):
        self.assertIsNotNone(import_violation(MODULE+"/contexts/menu/domain", "github.com/jackc/pgx/v5"))
    def test_application_cannot_read_another_context(self):
        self.assertIsNotNone(import_violation(MODULE+"/contexts/ordering/application", MODULE+"/contexts/menu/adaptors/postgres"))
    def test_application_cannot_import_protobuf(self):
        self.assertIsNotNone(import_violation(MODULE+"/contexts/ordering/application", MODULE+"/contracts/events/generated/cafe/v1"))
    def test_service_cannot_import_an_unowned_context(self):
        self.assertIsNotNone(import_violation(MODULE+"/apps/storefront", MODULE+"/contexts/loyalty/application"))
    def test_framework_independent_contract_is_allowed(self):
        self.assertIsNone(import_violation(MODULE+"/contexts/ordering/application", MODULE+"/contracts/events/model"))
    def test_foundation_cannot_become_a_business_facade(self):
        self.assertIsNotNone(import_violation(MODULE+"/foundation/runtime", MODULE+"/contexts/ordering/application"))
    def test_real_service_module_cannot_escape_ownership(self):
        self.assertIsNotNone(import_violation(MODULE+"/services/storefront/contexts/menu/domain",
                                             MODULE+"/services/storefront/foundation/transport/http"))
        self.assertIsNotNone(import_violation(MODULE+"/services/api/application",
                                             MODULE+"/services/storefront/contexts/menu/domain"))
    def test_python_domain_cannot_import_framework(self):
        self.assertIsNotNone(core_violation("services/operations/src/operations/contexts/preparation/domain.py", "fastapi"))
    def test_typescript_application_cannot_import_pg_or_another_context(self):
        path = "services/engagement/src/contexts/loyalty/application/commands.ts"
        self.assertIsNotNone(core_violation(path, "pg"))
        self.assertIsNotNone(core_violation(path, "services/engagement/src/contexts/communication/domain"))
    def test_python_application_can_use_own_domain(self):
        self.assertIsNone(core_violation("services/operations/src/operations/contexts/preparation/application.py",
                                        "operations/contexts/preparation/domain"))
    def test_acceptance_uses_wire_boundaries_instead_of_service_implementation(self):
        path = "tests/acceptance/workflow.steps.ts"
        self.assertIsNotNone(core_violation(path, "services/engagement/src/contexts/loyalty/application/commands.js"))
        self.assertIsNone(core_violation(path, "contracts/events/catalogue.json"))
        self.assertIsNone(core_violation(path, "tests/acceptance/client.js"))
    def test_python_relative_imports_cannot_escape_their_context(self):
        path = check_architecture.ROOT/"services/operations/src/operations/contexts/preparation/domain.py"
        read_text = Path.read_text
        for source in ("from ..collection.domain import Pickup", "from ..collection import domain",
                       "from .. import collection"):
            with self.subTest(source=source):
                def source_text(candidate, *args, **kwargs):
                    return source if candidate == path else read_text(candidate, *args, **kwargs)
                with patch.object(check_architecture, "files", return_value=[path]), \
                     patch.object(Path, "read_text", source_text), \
                     patch.object(check_architecture.subprocess, "run", return_value=SimpleNamespace(stdout="")):
                    with self.assertRaisesRegex(SystemExit, "foreign context implementation"):
                        check_architecture.check()
    def test_python_relative_imports_preserve_allowed_dependencies(self):
        relative = "services/operations/src/operations/contexts/preparation/application.py"
        path = check_architecture.ROOT/relative
        for source in ("from .domain import PreparationTicket", "from . import domain",
                       "from ...foundation.application import Metadata"):
            with self.subTest(source=source):
                for target in check_architecture.python_imports(path, source):
                    self.assertIsNone(core_violation(relative, target))

    def test_frontend_raw_http_calls_fail_enforcement(self):
        path = check_architecture.ROOT/"services/web/src/features/realtime/useCafe.ts"
        read_text = Path.read_text
        def source_text(candidate, *args, **kwargs):
            return "export const bypass = () => fetch('/api/v1/menu/drinks');" if candidate == path else read_text(candidate, *args, **kwargs)
        with patch.object(check_architecture, "files", return_value=[path]), \
             patch.object(Path, "read_text", source_text), \
             patch.object(check_architecture.subprocess, "run", return_value=SimpleNamespace(stdout="")):
            with self.assertRaisesRegex(SystemExit, "frontend HTTP calls"):
                check_architecture.check()

if __name__ == "__main__":
    unittest.main()
