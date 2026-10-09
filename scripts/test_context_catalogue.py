import json
from pathlib import Path
from tempfile import TemporaryDirectory
import unittest
from unittest.mock import patch
from context_catalogue import check, describe, entries
from command_boundaries import layout_violations
from check_architecture import core_violation, import_violation, MODULE


class ContextNavigationTests(unittest.TestCase):
    def test_all_current_contexts_have_valid_navigation(self):
        self.assertEqual(check(), [])
        for context in entries():
            output = describe(context['name'])
            self.assertIn(context['source']+'/application/queries/', output)
            self.assertIn('pnpm test:focused '+context['service'], output)

    def test_stale_catalogue_fails(self):
        with TemporaryDirectory() as folder:
            root = Path(folder)
            (root/'docs').mkdir()
            (root/'docs/contexts.json').write_text(json.dumps(entries()))
            self.assertTrue(any('ownership tree' in error for error in check(root)))

    def test_python_queries_and_reactions_have_one_responsibility_per_module(self):
        pair = 'class GetTicketQuery: pass\nclass GetTicketQueryHandler: pass\n'
        path = 'contexts/preparation/application/queries/get_ticket.py'
        self.assertEqual(layout_violations(pair, path), [])
        for source, location in [(pair, 'contexts/preparation/application.py'),
                                 (pair+'class ListTicketsQuery: pass', path),
                                 ('class OrderPlacedIntegrationEventHandler: pass', path)]:
            self.assertTrue(layout_violations(source, location))

    def test_shared_adaptors_cannot_import_owned_behaviour(self):
        for service, ext, prefix in [('operations', 'py', 'operations/'), ('engagement', 'ts', 'services/engagement/src/')]:
            self.assertIsNotNone(core_violation(f'services/{service}/src/{prefix}adaptors/restore.{ext}',
                prefix+'contexts/loyalty/domain/loyalty-account'))
        self.assertIsNone(core_violation('services/engagement/src/adaptors/codec.ts',
            'services/engagement/src/contexts/loyalty/adaptors/messaging/generated/private_messages.json'))

    def test_queries_cannot_return_domain_snapshots_directly(self):
        self.assertIsNotNone(core_violation('operations/contexts/preparation/application/queries/get_ticket.py',
            'operations/contexts/preparation/domain/preparation_ticket'))
        self.assertIsNotNone(import_violation(MODULE+'/services/storefront/contexts/menu/application/queries',
            MODULE+'/services/storefront/contexts/menu/domain'))

    def test_broken_context_readme_link_fails(self):
        read_text = Path.read_text
        source = entries()[0]['source']+'/README.md'

        def content(path, *args, **kwargs):
            text = read_text(path, *args, **kwargs)
            return text+'\n[Missing](missing-query.py)\n' if path.as_posix().endswith(source) else text

        with patch.object(Path, 'read_text', content):
            self.assertTrue(any('broken README link missing-query.py' in error for error in check()))
