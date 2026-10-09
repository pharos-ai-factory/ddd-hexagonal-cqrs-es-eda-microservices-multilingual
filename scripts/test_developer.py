from pathlib import Path
import unittest
from unittest.mock import patch
from developer import test as run_tests
from inspect_workflow import envelope_identity
from scaffold import render


class DeveloperTests(unittest.TestCase):
    def test_wrong_owner_cannot_select_foreign_context(self):
        with self.assertRaises(ValueError): run_tests('operations', 'loyalty')

    def test_focused_go_selection(self):
        with patch('developer.run') as run:
            run_tests('storefront', 'ordering')
            self.assertIn('./contexts/ordering/...', run.call_args.args[0])

    def test_inspection_reads_historical_headers_without_exposing_content(self):
        paths = list(Path('services').glob('**/fixtures/*.command.hex'))
        self.assertEqual(len(paths), 7)
        for path in paths:
            identity = envelope_identity(bytes.fromhex(path.read_text().strip()))
            self.assertEqual(set(identity), {'commandId', 'consumer', 'eventId', 'target', 'correlationId'})
        with self.assertRaises(ValueError): envelope_identity(b'\x0a\xff')

    def test_all_scaffold_roles_and_languages_have_registration_and_a_test(self):
        for context in ('ordering', 'preparation', 'loyalty'):
            for kind in ('command', 'query', 'subscription'):
                _, files = render(kind, context, 'ExampleUseCase')
                self.assertIn('registration.txt', files)
                self.assertTrue(any('test' in name for name in files))
        with self.assertRaises(ValueError): render('command', 'ordering', '../escape')
