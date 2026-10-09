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

    def test_command_starters_follow_the_owner_module_layout(self):
        _, go = render('command', 'ordering', 'CancelOrder')
        self.assertIn('package commands', go['commands/cancel_order.go'])
        self.assertIn('commands/cancel_order_test.go', go)
        _, python = render('command', 'preparation', 'CancelTicket')
        self.assertIn('commands/cancel_ticket.py', python)
        self.assertIn('test_cancel_ticket.py', python)
        _, typescript = render('command', 'loyalty', 'CancelReward')
        self.assertIn("'../../../../foundation/write-repository.js'", typescript['commands/cancel-reward.ts'])
        self.assertIn('commands/cancel-reward.test.ts', typescript)

    def test_python_focused_lane_discovers_all_owned_tests(self):
        with patch('developer.run') as run:
            run_tests('operations', 'preparation')
            self.assertIn('tests/contexts/preparation', run.call_args.args[0])
            self.assertIn('tests/bdd/test_preparation.py', run.call_args.args[0])
            self.assertNotIn('tests/test_domain.py', run.call_args.args[0])

    def test_query_and_subscription_starters_follow_the_owner_layout(self):
        for context, folder, suffix in [('ordering', 'eventhandlers', 'go'),
                                        ('preparation', 'event_handlers', 'py'),
                                        ('loyalty', 'event-handlers', 'ts')]:
            _, files = render('query', context, 'FindResource')
            self.assertTrue(any(path.startswith('queries/') for path in files))
            self.assertTrue(any(path.startswith('ports/') for path in files))
            self.assertTrue(any(path.startswith(('readmodels/', 'read_models/', 'read-models/')) for path in files))
            source = next(text for path, text in files.items() if path.startswith('queries/') and 'test' not in path)
            self.assertIn('FindResourceQueryHandler', source)
            self.assertIn('FindResourceQuery', source)
            _, files = render('subscription', context, 'ResourcePublished')
            self.assertTrue(any(path.startswith(folder+'/') and path.endswith('.'+suffix) for path in files))
