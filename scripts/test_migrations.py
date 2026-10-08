"""Migration manifests stay inside their owner and reject silent history edits."""
from hashlib import sha256
import json
from pathlib import Path
import shutil
import tempfile
import unittest
from migration_catalogue import CONTEXTS, ROOT, manifests, metadata
from migrate import migration_sql


class MigrationTests(unittest.TestCase):
    def setUp(self):
        self.temp = tempfile.TemporaryDirectory()
        self.addCleanup(self.temp.cleanup)
        self.root = Path(self.temp.name)
        shutil.copytree(ROOT/'devops/postgres/bootstrap', self.root/'devops/postgres/bootstrap')
        for _, (base, owners) in CONTEXTS.items():
            for owner in owners:
                path = Path(base)/owner/'adaptors/persistence/migrations'
                shutil.copytree(ROOT/path, self.root/path)

    def owner(self, owner='ordering'):
        return manifests(self.root)[owner][1]

    def test_owner_versions_evolve_independently(self):
        directory = self.owner()
        manifest = json.loads((directory/'manifest.json').read_text())
        source = 'CREATE INDEX test_next ON cafe.aggregates(version);\n'
        (directory/'0003_next.sql').write_text(source)
        manifest['migrations'].append({'version': 3, 'file': '0003_next.sql', 'checksum': sha256(source.encode()).hexdigest()})
        (directory/'manifest.json').write_text(json.dumps(manifest))
        generated = metadata(self.root)
        self.assertEqual(len(generated['storefront']['ordering']), 3)
        self.assertEqual(len(generated['storefront']['menu']), 2)
        self.assertEqual(set(generated['engagement']), {'loyalty', 'communication'})
        self.assertIn('test_next', migration_sql('ordering', self.root))

    def test_changed_source_needs_reviewed_manifest(self):
        (self.owner()/'0001_lifecycle_indexes.sql').write_text('SELECT 1;')
        with self.assertRaisesRegex(ValueError, 'checksum'):
            manifests(self.root)

    def test_legacy_history_is_frozen(self):
        (self.root/'devops/postgres/bootstrap/0001_initial.sql').write_text('SELECT 1;')
        with self.assertRaisesRegex(ValueError, 'immutable'):
            manifests(self.root)

    def test_manifest_cannot_import_foreign_source(self):
        directory = self.owner()
        manifest = json.loads((directory/'manifest.json').read_text())
        manifest['migrations'][0]['file'] = '../foreign.sql'
        (directory/'manifest.json').write_text(json.dumps(manifest))
        with self.assertRaisesRegex(ValueError, 'filename'):
            manifests(self.root)

    def test_manifest_owner_cannot_change(self):
        directory = self.owner()
        manifest = json.loads((directory/'manifest.json').read_text())
        manifest['owner'] = 'loyalty'
        (directory/'manifest.json').write_text(json.dumps(manifest))
        with self.assertRaisesRegex(ValueError, 'owner mismatch'):
            manifests(self.root)

    def test_migration_cannot_commit_before_ledger(self):
        directory = self.owner()
        manifest = json.loads((directory/'manifest.json').read_text())
        source = 'COMMIT;\n'
        (directory/manifest['migrations'][0]['file']).write_text(source)
        manifest['migrations'][0]['checksum'] = sha256(source.encode()).hexdigest()
        (directory/'manifest.json').write_text(json.dumps(manifest))
        with self.assertRaisesRegex(ValueError, 'transactions'):
            manifests(self.root)


if __name__ == '__main__':
    unittest.main()
