"""Context-owned migration manifests and generated runtime checksum metadata."""
from hashlib import sha256
import json
from pathlib import Path
import re

ROOT = Path(__file__).resolve().parents[1]
CONTEXTS = {
    'storefront': ('services/storefront/contexts', ('menu', 'ordering')),
    'operations': ('services/operations/src/operations/contexts', ('preparation', 'collection')),
    'engagement': ('services/engagement/src/contexts', ('loyalty', 'communication')),
}
LEGACY = {
    '0001_initial.sql': '1b3e16641bbbbb9ce79620970a288b502b26b3adcc82568eb6c2f5360fdc4738',
    '0002_realtime.sql': '91307774de1e1254578913bb9e5523f8bc05dd9ef497e46bd4b6b62a4d5a45e1',
}
OUTPUTS = {
    'storefront': 'services/storefront/foundation/persistence/postgres/migrations/contexts.json',
    'operations': 'services/operations/src/operations/adaptors/generated/context-persistence.json',
    'engagement': 'services/engagement/src/adaptors/generated/context-persistence.json',
}


def manifests(root=ROOT):
    for name, checksum in LEGACY.items():
        if sha256((root/'contracts/persistence'/name).read_bytes()).hexdigest() != checksum:
            raise ValueError('Historical migration is immutable: '+name)
    result = {}
    for service, (base, owners) in CONTEXTS.items():
        for owner in owners:
            directory = root/base/owner/'adaptors/persistence/migrations'
            manifest = json.loads((directory/'manifest.json').read_text())
            if manifest.get('owner') != owner:
                raise ValueError(f'{owner}: migration manifest owner mismatch')
            migrations = manifest['migrations']
            if not migrations or [item['version'] for item in migrations] != list(range(1, len(migrations)+1)):
                raise ValueError(f'{owner}: migrations must have consecutive local versions starting at 1')
            for item in migrations:
                name = item['file']
                if not re.fullmatch(r'[0-9]{4}_[a-z_]+\.sql', name):
                    raise ValueError(f'{owner}: invalid migration filename')
                source = (directory/name).read_text()
                if sha256(source.encode()).hexdigest() != item['checksum']:
                    raise ValueError(f'{owner}: migration checksum differs from manifest: {name}')
                # The administrator owns atomicity and psql commands, never an individual migration.
                if re.search(r'(?im)^\s*(BEGIN|COMMIT|ROLLBACK)\s*;|^\s*\\', source):
                    raise ValueError(f'{owner}: migration cannot control transactions or psql')
            result[owner] = (service, directory, manifest)
    return result


def metadata(root=ROOT):
    result = {service: {} for service in CONTEXTS}
    for owner, (service, _, manifest) in manifests(root).items():
        result[service][owner] = {str(item['version']): item['checksum'] for item in manifest['migrations']}
    return result


def generate(root=ROOT):
    for service, data in metadata(root).items():
        (root/OUTPUTS[service]).write_text(json.dumps(data, indent=2)+'\n')


if __name__ == '__main__':
    generate()
