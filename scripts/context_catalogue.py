"""Navigate business contexts and verify their developer entry points."""
import argparse
import json
from pathlib import Path
import re

ROOT = Path(__file__).resolve().parents[1]


def entries(root=ROOT):
    return json.loads((root/'docs/contexts.json').read_text())


def services(root=ROOT):
    result = {'storefront': [], 'operations': [], 'engagement': [], 'api': [], 'web': []}
    for entry in entries(root):
        result[entry['service']].append(entry['name'])
    return result


def check(root=ROOT):
    errors, seen = [], set()
    declared = entries(root)
    detected = {path.relative_to(root).as_posix() for pattern in (
        'services/storefront/contexts/*', 'services/operations/src/operations/contexts/*',
        'services/engagement/src/contexts/*') for path in root.glob(pattern)
        if path.is_dir() and path.name != '__pycache__'}
    if detected != {entry['source'] for entry in declared}:
        errors.append('context catalogue differs from the source ownership tree')
    for entry in declared:
        name, service = entry['name'], entry['service']
        if name in seen:
            errors.append('duplicate context: '+name)
        seen.add(name)
        expected = {'storefront': 'contexts', 'operations': 'src/operations/contexts',
                    'engagement': 'src/contexts'}.get(service)
        if not expected or entry['source'] != f'services/{service}/{expected}/{name}':
            errors.append('invalid context owner path: '+name)
        source = root/entry['source']
        readme = source/'README.md'
        required = [readme, source/'domain', source/'application/commands',
                    source/'application/queries', source/'application/ports',
                    root/'contracts'/name, root/'specifications'/name]
        required += [root/'services/web/src/features'/feature for feature in entry['features']]
        for path in required:
            if not path.exists():
                errors.append(f'{name}: missing navigation target {path.relative_to(root)}')
        if readme.exists():
            for target in re.findall(r'\[[^\]]+\]\(([^)]+)\)', readme.read_text()):
                if not target.startswith(('https://', '#')) and not (source/target.split('#')[0]).exists():
                    errors.append(f'{name}: broken README link {target}')
            command = f'pnpm test:focused {service} --context {name}'
            if command not in readme.read_text():
                errors.append(f'{name}: README lacks its focused test command')
    return errors


def describe(name, root=ROOT):
    entry = next(item for item in entries(root) if item['name'] == name)
    source = root/entry['source']
    lines = [f"{name.title()} — {entry['purpose']}", '', 'Start: '+entry['source']+'/README.md']
    for role in ('commands', 'queries', 'event_handlers', 'event-handlers', 'projections'):
        folder = source/'application'/role
        if folder.exists():
            files = [p.relative_to(root).as_posix() for p in sorted(folder.iterdir())
                     if p.suffix in ('.go', '.py', '.ts') and not p.name.startswith('__')
                     and not p.name.endswith(('_test.go', '.test.ts'))]
            lines.extend(['', role.replace('_', ' ').replace('-', ' ').title()+':', *('  '+p for p in files)])
    lines.extend(['', 'Contracts: contracts/'+name, 'Scenarios: specifications/'+name,
                  '', f"Test: pnpm test:focused {entry['service']} --context {name}"])
    return '\n'.join(lines)


def main():
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument('context', nargs='?', choices=[e['name'] for e in entries()])
    parser.add_argument('--check', action='store_true')
    args = parser.parse_args()
    if args.check:
        errors = check()
        if errors:
            raise SystemExit('\n'.join(errors))
        print('Context navigation verified')
    elif args.context:
        print(describe(args.context))
    else:
        for entry in entries():
            print(f"{entry['name']:15} {entry['service']:12} {entry['purpose']}")


if __name__ == '__main__':
    main()
