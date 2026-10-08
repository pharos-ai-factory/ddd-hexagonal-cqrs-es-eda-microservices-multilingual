"""Specification mutations must break real frontend uses of incompatible shapes."""
import copy
from pathlib import Path
import shutil
import subprocess
import tempfile
import unittest
from frontend_contracts import generated
from http_contracts import ROOT, SOURCE, DOCUMENTS, bundle


class FrontendContractsTests(unittest.TestCase):
    def compile(self, document, probe=""):
        with tempfile.TemporaryDirectory() as directory:
            target = Path(directory)/'web'
            shutil.copytree(ROOT/'services/web/src', target/'src')
            (target/'tsconfig.json').write_text((ROOT/'services/web/tsconfig.json').read_text())
            (target/'node_modules').symlink_to(ROOT/'services/web/node_modules', target_is_directory=True)
            (target/'src/adaptors/generated/http.ts').write_text(generated(document))
            (target/'src/contract-probe.ts').write_text(probe)
            return subprocess.run(['node', str(ROOT/'node_modules/typescript/bin/tsc'), '--noEmit',
                                   '--incremental', 'false', '--project', str(target/'tsconfig.json')],
                                  cwd=ROOT, text=True, capture_output=True)

    def test_incompatible_inputs_and_outputs_break_actual_frontend_calls(self):
        document = bundle(SOURCE/DOCUMENTS['api'])
        for component, old, new in (('menu.http_api.schemas.DrinkCommand', 'name', 'label'),
                                    ('menu.http_api.schemas.Drink', 'name', 'label')):
            with self.subTest(component=component):
                changed = copy.deepcopy(document)
                schema = changed['components']['schemas'][component]
                schema['properties'][new] = schema['properties'].pop(old)
                schema['required'] = [new if name == old else name for name in schema['required']]
                result = self.compile(changed)
                self.assertNotEqual(result.returncode, 0, result.stdout+result.stderr)
                self.assertIn('src/features/', result.stdout)

    def test_optional_input_addition_keeps_existing_frontend_compatible(self):
        document = bundle(SOURCE/DOCUMENTS['api'])
        document['components']['schemas']['menu.http_api.schemas.DrinkCommand']['properties']['note'] = {'type': 'string'}
        result = self.compile(document)
        self.assertEqual(result.returncode, 0, result.stdout+result.stderr)

    def test_required_parameters_break_real_command_and_list_callers(self):
        document = bundle(SOURCE/DOCUMENTS['api'])
        for path, method, location in (('/api/v1/menu/drinks/{id}', 'post', 'header'),
                                       ('/api/v1/menu/drinks/{id}', 'post', 'query'),
                                       ('/api/v1/menu/drinks', 'get', 'query')):
            with self.subTest(path=path, location=location):
                changed = copy.deepcopy(document)
                changed['paths'][path][method].setdefault('parameters', []).append({
                    'in': location, 'name': 'requiredReason', 'required': True, 'schema': {'type': 'string'}})
                result = self.compile(changed)
                self.assertNotEqual(result.returncode, 0, result.stdout+result.stderr)
                self.assertIn('src/features/', result.stdout)

    def test_required_query_object_cannot_be_omitted(self):
        document = bundle(SOURCE/DOCUMENTS['api'])
        document['paths']['/api/v1/menu/drinks/{id}']['get'].setdefault('parameters', []).append({
            'in': 'query', 'name': 'reason', 'required': True, 'schema': {'type': 'string'}})
        probe = "import {request} from './adaptors/http/client';\nrequest('getDrink', {path: {id: 'test'}});"
        rejected = self.compile(document, probe)
        self.assertNotEqual(rejected.returncode, 0)
        self.assertIn('query', rejected.stdout)
        accepted = self.compile(document, probe.replace("{path: {id: 'test'}}", "{path: {id: 'test'}, query: {reason: 'test'}}"))
        self.assertEqual(accepted.returncode, 0, accepted.stdout+accepted.stderr)

    def test_optional_parameters_remain_compatible(self):
        document = bundle(SOURCE/DOCUMENTS['api'])
        for location in ('header', 'query'):
            document['paths']['/api/v1/menu/drinks/{id}']['post'].setdefault('parameters', []).append({
                'in': location, 'name': 'optionalReason', 'required': False, 'schema': {'type': 'string'}})
        result = self.compile(document)
        self.assertEqual(result.returncode, 0, result.stdout+result.stderr)


if __name__ == '__main__':
    unittest.main()
