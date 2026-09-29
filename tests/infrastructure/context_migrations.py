"""Prove owner migration upgrades and failed administrative work against PostgreSQL."""
from hashlib import sha256
import json
import os
from pathlib import Path
import shutil
import subprocess
import sys
import tempfile
import unittest
from uuid import uuid4

ROOT = Path(__file__).resolve().parents[2]
sys.path.insert(0, str(ROOT/'scripts'))
from dev import compose, load
from migration_catalogue import CONTEXTS, manifests
from migrate import migration_sql


class ContextMigrations(unittest.TestCase):
    @classmethod
    def setUpClass(cls):
        cls.env_file = Path(os.environ['CAFE_ENV_FILE'])
        if not load(cls.env_file)['COMPOSE_PROJECT_NAME'].startswith('cafe-reference-test-'):
            raise RuntimeError('Migration probes require the disposable integration project')

    def sql(self, source, owner='ordering'):
        return compose(self.env_file, 'exec', '-T', 'postgres', 'psql', '-X', '-At', '-v', 'ON_ERROR_STOP=1',
                       '-U', 'postgres', '-d', 'cafe_'+owner, input=source, text=True,
                       stdout=subprocess.PIPE, stderr=subprocess.PIPE).stdout.strip()

    def copy_sources(self):
        temp = tempfile.TemporaryDirectory()
        self.addCleanup(temp.cleanup)
        root = Path(temp.name)
        shutil.copytree(ROOT/'contracts/persistence', root/'contracts/persistence')
        for _, (base, owners) in CONTEXTS.items():
            for owner in owners:
                path = Path(base)/owner/'adaptors/persistence/migrations'
                shutil.copytree(ROOT/path, root/path)
        return root

    def append(self, root, source):
        directory = manifests(root)['ordering'][1]
        manifest = json.loads((directory/'manifest.json').read_text())
        (directory/'0002_probe.sql').write_text(source)
        manifest['migrations'].append({'version': 2, 'file': '0002_probe.sql', 'checksum': sha256(source.encode()).hexdigest()})
        (directory/'manifest.json').write_text(json.dumps(manifest))

    def test_upgrade_existing_volume_preserves_state_and_is_repeatable(self):
        identity = str(uuid4())
        self.addCleanup(self.sql, f"DELETE FROM cafe.aggregates WHERE kind='migration_probe' AND id='{identity}';")
        self.sql(f"""BEGIN;
 SELECT set_config('cafe.command_target','migration_probe:{identity}',true);
 INSERT INTO cafe.aggregates(kind,id,version,state) VALUES('migration_probe','{identity}',1,'{{"kept":true}}');
 COMMIT;
 DROP TABLE cafe.context_migrations,cafe.context_identity;
 DROP INDEX cafe.ordering_order_status;""")
        self.addCleanup(self.sql, migration_sql('ordering'))
        self.sql(migration_sql('ordering'))
        applied = self.sql('SELECT applied_at FROM cafe.context_migrations WHERE version=1;')
        self.sql(migration_sql('ordering'))
        self.assertEqual(self.sql('SELECT applied_at FROM cafe.context_migrations WHERE version=1;'), applied)
        self.assertEqual(self.sql(f"SELECT state->>'kept' FROM cafe.aggregates WHERE id='{identity}';"), 'true')

    def test_changed_applied_checksum_fails_before_new_migration(self):
        root = self.copy_sources()
        self.append(root, 'CREATE TABLE cafe.migration_probe(id int);\n')
        checksum = manifests()['ordering'][2]['migrations'][0]['checksum']
        self.addCleanup(self.sql, f"UPDATE cafe.context_migrations SET checksum='{checksum}' WHERE version=1;")
        self.sql("UPDATE cafe.context_migrations SET checksum='tampered' WHERE version=1;")
        with self.assertRaises(subprocess.CalledProcessError) as failure:
            self.sql(migration_sql('ordering', root))
        self.assertIn('checksum changed', failure.exception.stderr)
        self.assertEqual(self.sql("SELECT to_regclass('cafe.migration_probe') IS NULL;"), 't')
        self.assertEqual(self.sql('SELECT count(*) FROM cafe.context_migrations;'), '1')

    def test_failing_migration_rolls_back_schema_and_ledger(self):
        root = self.copy_sources()
        self.append(root, 'CREATE TABLE cafe.migration_probe(id int);\nSELECT 1/0;\n')
        with self.assertRaises(subprocess.CalledProcessError):
            self.sql(migration_sql('ordering', root))
        self.assertEqual(self.sql("SELECT to_regclass('cafe.migration_probe') IS NULL;"), 't')
        self.assertEqual(self.sql('SELECT count(*) FROM cafe.context_migrations;'), '1')

    def test_context_evolves_without_advancing_sibling(self):
        root = self.copy_sources()
        self.append(root, 'CREATE TABLE cafe.migration_probe(id int);\n')
        self.addCleanup(self.sql, 'DROP TABLE IF EXISTS cafe.migration_probe; DELETE FROM cafe.context_migrations WHERE version=2;')
        self.sql(migration_sql('ordering', root))
        self.assertEqual(self.sql('SELECT max(version) FROM cafe.context_migrations;'), '2')
        self.assertEqual(self.sql('SELECT max(version) FROM cafe.context_migrations;', 'menu'), '1')
        self.sql(migration_sql('menu'), 'menu')

    def test_identity_mismatch_is_rejected(self):
        self.addCleanup(self.sql, "UPDATE cafe.context_identity SET owner='ordering';")
        self.sql("UPDATE cafe.context_identity SET owner='menu';")
        with self.assertRaises(subprocess.CalledProcessError) as failure:
            self.sql(migration_sql('ordering'))
        self.assertIn('Context identity mismatch', failure.exception.stderr)

    def test_wrong_database_is_rejected(self):
        with self.assertRaises(subprocess.CalledProcessError) as failure:
            self.sql(migration_sql('ordering'), 'menu')
        self.assertIn('Database owner mismatch', failure.exception.stderr)

    def test_runtime_cannot_rewrite_migration_authority(self):
        for query in ('UPDATE cafe.context_identity SET owner=owner;',
                      "UPDATE cafe.context_migrations SET checksum='tampered';"):
            with self.assertRaises(subprocess.CalledProcessError):
                self.sql('SET ROLE cafe_ordering; '+query)


if __name__ == '__main__':
    unittest.main()
