#!/usr/bin/env python3
"""Check service ownership and inward dependencies in all implementation languages."""
import ast
from hashlib import sha256
import json
import os
from pathlib import Path
import re
import subprocess
import sys

ROOT = Path(__file__).resolve().parents[1]
MODULE = 'github.com/pharos-ai-factory/ddd-hexagonal-cqrs-es-eda-microservices-multilingual'
SERVICES = {'storefront': {'menu', 'ordering'}, 'operations': {'preparation', 'collection'},
            'engagement': {'loyalty', 'communication'}}
IGNORED = {'.git', '.tools', '.local', 'node_modules', '__pycache__', '.venv', '.next',
           'dist', 'test-results', 'playwright-report', '.pytest_cache', '.ruff_cache', '.mypy_cache'}
GENERATED = {
    'services/storefront/contracts/events/generated/cafe/v1/events.pb.go',
    'services/storefront/contracts/realtime/generated/cafe/realtime/v1/realtime.pb.go',
    'services/operations/src/operations/adaptors/generated/cafe/v1/events_pb2.py',
    'services/operations/src/operations/adaptors/generated/cafe/realtime/v1/realtime_pb2.py',
    'services/operations/src/operations/adaptors/generated/cafe/v1/events_pb2.pyi',
    'services/operations/src/operations/adaptors/generated/cafe/realtime/v1/realtime_pb2.pyi',
    'services/engagement/src/adaptors/generated/events.json',
    'services/engagement/src/adaptors/generated/realtime.json',
    'services/web/src/adaptors/generated/realtime.json',
    'pnpm-lock.yaml', 'services/operations/uv.lock',
}


def import_violation(package, imported):
    path = package.removeprefix(MODULE+'/')
    target = imported.removeprefix(MODULE+'/')
    internal = imported.startswith(MODULE+'/')
    if path.startswith('services/'):
        service = path.split('/')[1]
        if internal and not target.startswith('services/'+service+'/'):
            return 'cross-service source dependency'
        path = '/'.join(path.split('/')[2:])
        if internal:
            target = '/'.join(target.split('/')[2:])
    else:
        service = 'storefront'
    if path.startswith('contexts/'):
        owner, ring = path.split('/')[1:3]
        if internal and target.startswith('contexts/') and target.split('/')[1] != owner:
            return 'foreign context implementation'
        if ring in {'domain', 'application'}:
            allowed = [f'contexts/{owner}/domain', 'foundation/domain']
            if ring == 'application':
                allowed += [f'contexts/{owner}/application', 'foundation/application', 'contracts/events/model']
            if internal and not any(target == p or target.startswith(p+'/') for p in allowed):
                return 'outward core dependency'
            if not internal and ('.' in imported.split('/')[0] or imported.split('/')[0] in {'net', 'os', 'database', 'log'}):
                return 'infrastructure in core'
    if path.startswith('foundation/') and internal and not target.startswith('foundation/'):
        return 'business dependency in technical foundation'
    if path.startswith('apps/') and target.startswith('contexts/'):
        if target.split('/')[1] not in SERVICES.get(service, set()):
            return 'service imports a context it does not own'
        if path.startswith('apps/support'):
            return 'shared composition contains context behaviour'
    if service == 'api' and path.startswith('application/') and internal and not target.startswith('application/'):
        return 'outward API application dependency'
    return None


def core_violation(path, target):
    """Normalised service-local Python/TypeScript module paths."""
    if path.startswith('tests/acceptance/') and target.startswith('services/'):
        return 'cross-service acceptance must use published boundaries'
    parts = path.split('/')
    if 'foundation' in parts and '/contexts/' in '/'+target:
        return 'business dependency in technical foundation'
    if 'contexts' not in parts:
        return None
    offset = parts.index('contexts')
    owner, ring = parts[offset+1], parts[offset+2].split('.')[0]
    if 'contexts/' in target and target.split('contexts/')[1].split('/')[0] != owner:
        return 'foreign context implementation'
    if ring in {'domain', 'application'}:
        if any(p in target.split('/') for p in ('adaptors', 'apps', 'generated')):
            return 'outward core dependency'
        if ring == 'domain' and any(p in target.split('/') for p in ('application', 'contracts')):
            return 'outward domain dependency'
        if target.startswith(('node:', 'pg', 'amqplib', 'protobuf', 'react', 'next', 'fastapi', 'pika', 'psycopg', 'requests', 'http', 'socket', 'os')):
            return 'infrastructure in core'
    return None


def files():
    for directory, directories, names in os.walk(ROOT):
        directories[:] = [name for name in directories if name not in IGNORED]
        for name in names:
            yield Path(directory)/name


def python_imports(path, content):
    for node in ast.walk(ast.parse(content)):
        if isinstance(node, ast.Import):
            yield from (item.name.replace('.', '/') for item in node.names)
        elif isinstance(node, ast.ImportFrom):
            if node.level:
                base = path.parent
                for _ in range(node.level - 1):
                    base = base.parent
                target = base.relative_to(ROOT).as_posix()
                if node.module:
                    target += '/' + node.module.replace('.', '/')
            else:
                target = (node.module or '').replace('.', '/')
            yield target
            # ImportFrom may name a child module, including `from .. import collection`.
            for item in node.names:
                if item.name != '*':
                    yield target + '/' + item.name.replace('.', '/')


def check():
    errors, packages = [], []
    for service in ('storefront', 'api'):
        result = subprocess.run([sys.executable, 'scripts/go.py', 'list', '-json', './...'], cwd=ROOT,
            env=dict(os.environ, CAFE_GO_PROJECT='services/'+service), capture_output=True, text=True, check=True)
        decoder, text = json.JSONDecoder(), result.stdout.strip()
        while text:
            package, offset = decoder.raw_decode(text)
            packages.append(package)
            text = text[offset:].lstrip()
    for package in packages:
        for imported in package.get('Imports', []):
            reason = import_violation(package['ImportPath'], imported)
            if reason:
                errors.append(f'{package["ImportPath"]}: {reason}: {imported}')
    for path in files():
        relative = path.relative_to(ROOT).as_posix()
        if path.suffix not in {'.go', '.py', '.pyi', '.md', '.sql', '.yaml', '.proto', '.ts', '.tsx', '.json', '.mjs', '.css', '.feature'}:
            continue
        content = path.read_text()
        if len(content.splitlines()) > 450 and relative not in GENERATED:
            errors.append(f'{relative} exceeds 450 lines without a responsibility review')
        if relative in GENERATED or path.name.endswith(('_test.go', '.test.ts', '.spec.ts')) or path.name.startswith('test_'):
            continue
        if relative.startswith('tests/acceptance/') and path.suffix == '.ts':
            for target in re.findall(r'(?:from\s+|import\s*)[\'\"]([^\'\"]+)', content):
                normal = (path.parent/target).resolve().relative_to(ROOT).as_posix() if target.startswith('.') else target
                if reason := core_violation(relative, normal):
                    errors.append(f'{relative}: {reason}: {target}')
        if relative.startswith('services/'):
            service = relative.split('/')[1]
            if '/contexts/' in relative and path.name != '__init__.py':
                owner = relative.split('/contexts/')[1].split('/')[0]
                if owner not in SERVICES.get(service, set()):
                    errors.append(f'{relative}: service does not own context {owner}')
            if path.suffix in {'.py', '.pyi'}:
                for target in python_imports(path, content):
                    if target.startswith('services/') and target.split('/')[1] != service:
                        errors.append(f'{relative}: cross-service source dependency')
                    if reason := core_violation(relative, target):
                        errors.append(f'{relative}: {reason}: {target}')
            if path.suffix in {'.ts', '.tsx'}:
                for target in re.findall(r'(?:from\s+|import\s*)[\'\"]([^\'\"]+)', content):
                    normal = (path.parent/target).resolve().relative_to(ROOT).as_posix() if target.startswith('.') else target
                    if normal.startswith('services/') and normal.split('/')[1] != service:
                        errors.append(f'{relative}: cross-service source dependency')
                    if reason := core_violation(relative, normal):
                        errors.append(f'{relative}: {reason}: {target}')
            if re.search(r'/domain(?:/|\.py)', relative) and re.search(r'\btime\.Now\s*\(|\bDate\.now\s*\(|new Date\(\)|datetime\.now\s*\(', content):
                errors.append(f'{relative}: implicit domain clock')
            if service == 'web' and re.search(r'\bnewSubscription\s*\(|\bsetInterval\s*\(', content):
                errors.append(f'{relative}: client-selected subscription or browser polling')
    if {p.name for p in (ROOT/'services').iterdir() if p.is_dir()} != {*SERVICES, 'api', 'web'}:
        errors.append('expected exactly three domain services, Go API and Next.js web')
    baseline = ROOT/'contracts/persistence/0001_initial.sql'
    if baseline.read_bytes() != (ROOT/'services/storefront/foundation/persistence/postgres/migrations/0001_initial.sql').read_bytes():
        errors.append('Go embedded migration differs from the shared persistence contract')
    catalogue = json.loads((ROOT/'contracts/events/catalogue.json').read_text())
    go_catalogue = (ROOT/'services/storefront/contracts/events/model/catalogue.go').read_text()
    parsed = [{'name': name, 'owner': owner, 'visibility': 'domain' if visibility == 'Private' else 'integration',
               'consumer': consumer, 'kind': kind} for name, owner, visibility, consumer, kind in re.findall(
        r'\{"([^"]+)", "([^"]+)", application\.(Private|Public), \[\]string\{"([^"]+)"\}, "([^"]+)"\}', go_catalogue)]
    if parsed != catalogue:
        errors.append('Go event catalogue differs from the canonical wire catalogue')
    checksums = {str(version): sha256((ROOT/f'contracts/persistence/{name}').read_bytes()).hexdigest()
                 for version, name in ((1, '0001_initial.sql'), (2, '0002_realtime.sql'))}
    for folder in ('services/operations/src/operations/adaptors/generated',
                   'services/engagement/src/adaptors/generated'):
        if json.loads((ROOT/folder/'catalogue.json').read_text()) != catalogue:
            errors.append(folder+': catalogue drift')
        if json.loads((ROOT/folder/'persistence.json').read_text()) != checksums:
            errors.append(folder+': migration checksum drift')
    if (ROOT/'contracts/persistence/0002_realtime.sql').read_bytes() != (
            ROOT/'services/storefront/foundation/persistence/postgres/migrations/0002_realtime.sql').read_bytes():
        errors.append('Go realtime migration drift')
    for name in ('events', 'realtime'):
        if not (ROOT/f'contracts/{name}/proto').is_dir():
            errors.append('missing separate wire contract: '+name)
    compose = (ROOT/'devops/compose.yaml').read_text()
    if re.search(r'APP_ENV:\s*(production|staging)', compose):
        errors.append('non-development runtime composition')
    if errors:
        raise SystemExit('\n'.join(errors))
    print(f'Architecture verified across {len(packages)} Go packages, Python, TypeScript and Next.js')


if __name__ == '__main__':
    check()
