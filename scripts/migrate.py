#!/usr/bin/env python3
"""Apply one context's checksummed migrations using the development administrator."""
import argparse
from hashlib import sha256
from pathlib import Path
import subprocess
from migration_catalogue import ROOT, manifests


def literal(value):
    return "'" + str(value).replace("'", "''") + "'"


def migration_sql(owner, root=ROOT):
    entries = manifests(root)
    if owner not in entries:
        raise ValueError('Unknown migration owner')
    _, directory, manifest = entries[owner]
    initial = (root/'devops/postgres/bootstrap/0001_initial.sql').read_bytes()
    realtime = (root/'devops/postgres/bootstrap/0002_realtime.sql').read_text()
    initial_hash, realtime_hash = sha256(initial).hexdigest(), sha256(realtime.encode()).hexdigest()
    # Legacy schemas stay frozen. Existing volumes may predate realtime support.
    statements = [r'\set ON_ERROR_STOP on', f"""DO $$ BEGIN
 IF current_database() <> 'cafe_{owner}' THEN RAISE EXCEPTION 'Database owner mismatch'; END IF;
 IF NOT EXISTS (SELECT FROM cafe.schema_version WHERE version=1 AND checksum='{initial_hash}')
 THEN RAISE EXCEPTION 'Legacy migration checksum mismatch'; END IF;
 END $$;""", f"\\set role cafe_{owner}", f"\\set checksum {realtime_hash}", realtime,
        'BEGIN;', "SELECT pg_advisory_xact_lock(hashtextextended('cafe.context-migrations',0));",
        """CREATE TABLE IF NOT EXISTS cafe.context_identity(
 singleton boolean PRIMARY KEY CHECK(singleton), owner text NOT NULL);
 CREATE TABLE IF NOT EXISTS cafe.context_migrations(
 version integer PRIMARY KEY CHECK(version>0), checksum text NOT NULL,
 applied_at timestamptz NOT NULL DEFAULT clock_timestamp());""",
        f"INSERT INTO cafe.context_identity VALUES(true,'{owner}') ON CONFLICT DO NOTHING;",
        f"""DO $$ BEGIN
 IF NOT EXISTS (SELECT FROM cafe.context_identity WHERE singleton AND owner='{owner}')
 THEN RAISE EXCEPTION 'Context identity mismatch'; END IF;
 END $$;""",
        f'GRANT SELECT ON cafe.context_identity,cafe.context_migrations TO cafe_{owner};']
    expected = ','.join(f"({item['version']},{literal(item['checksum'])})" for item in manifest['migrations'])
    statements.append(f"""DO $$ BEGIN
 IF EXISTS (SELECT FROM cafe.context_migrations actual LEFT JOIN (VALUES {expected}) expected(version,checksum)
 ON actual.version=expected.version AND actual.checksum=expected.checksum WHERE expected.version IS NULL)
 THEN RAISE EXCEPTION 'Applied context migration is unknown or its checksum changed'; END IF;
 END $$;""")
    for item in manifest['migrations']:
        version = item['version']
        statements += [f'SELECT NOT EXISTS (SELECT FROM cafe.context_migrations WHERE version={version}) AS apply_migration \\gset',
                       r'\if :apply_migration', (directory/item['file']).read_text(),
                       f"INSERT INTO cafe.context_migrations(version,checksum) VALUES({version},{literal(item['checksum'])});",
                       r'\endif']
    statements.append('COMMIT;')
    return '\n'.join(statements)+'\n'


def migrate(env_file, owners=None):
    # Imported lazily so catalogue validation and SQL rendering need no Docker configuration.
    from dev import compose
    for owner in owners or manifests():
        compose(env_file, 'exec', '-T', 'postgres', 'psql', '-X', '-U', 'postgres', '-d', 'cafe_'+owner,
                input=migration_sql(owner), text=True, stdout=subprocess.DEVNULL)


if __name__ == '__main__':
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument('--env-file', type=Path, default=ROOT/'.local/dev.env')
    parser.add_argument('--owner', choices=tuple(manifests()))
    args = parser.parse_args()
    migrate(args.env_file, [args.owner] if args.owner else None)
