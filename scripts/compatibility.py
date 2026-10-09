"""Compare published HTTP and Protobuf interfaces and persisted private formats with Git history."""
import argparse
import json
import os
from pathlib import Path
import subprocess
import tempfile

ROOT = Path(__file__).resolve().parents[1]
BASELINE = 'caa3206'


def proto_snapshot(root):
    """Compile every schema using its stable logical imports, including owner-private formats."""
    import sys
    sys.path.insert(0, str(ROOT/'scripts'))
    from contract_sources import sources
    from google.protobuf import descriptor_pb2
    import grpc_tools
    result = {}
    groups = {kind: sources(kind, root) for kind in ('events', 'requests', 'realtime', 'menu_private', 'loyalty_private', 'communication_private')}
    for path in root.glob('services/**/internal_commands.proto'):
        owner = path.parts[path.parts.index('contexts')+1]
        groups[owner+'_commands'] = {f'cafe/internal/{owner}/internal_commands.proto': path}
    for group, definitions in groups.items():
        with tempfile.TemporaryDirectory() as directory:
            stage = Path(directory)
            for logical, source in definitions.items():
                target = stage/logical
                target.parent.mkdir(parents=True, exist_ok=True)
                target.write_bytes(source.read_bytes())
            output = stage/'schema.pb'
            subprocess.run([sys.executable, '-m', 'grpc_tools.protoc', '-I'+str(stage),
                '-I'+str(Path(grpc_tools.__file__).parent/'_proto'), '--include_imports',
                '--descriptor_set_out='+str(output), *definitions], check=True)
            descriptors = descriptor_pb2.FileDescriptorSet.FromString(output.read_bytes())
        def messages(prefix, items):
            for item in items:
                name = prefix+'.'+item.name
                result[name] = {'fields': {str(f.number): {
                    'name': f.name, 'json': f.json_name, 'type': f.type, 'type_name': f.type_name,
                    'label': f.label, 'optional': f.proto3_optional, 'default': f.default_value,
                    'oneof': item.oneof_decl[f.oneof_index].name if f.HasField('oneof_index') else '',
                    'options': f.options.SerializeToString().hex(),
                } for f in item.field}}
                messages(name, item.nested_type)
                enums(name, item.enum_type)
        def enums(prefix, items):
            for item in items:
                result[prefix+'.'+item.name] = {'values': {v.name: v.number for v in item.value}}
        for file in descriptors.file:
            if file.name in definitions:
                messages(file.package, file.message_type)
                enums(file.package, file.enum_type)
    return result


def proto_changes(old, new):
    errors = []
    for name, before in old.items():
        after = new.get(name)
        if after is None:
            errors.append(name+': removed message or enum')
            continue
        for category, values in before.items():
            for key, value in values.items():
                if after.get(category, {}).get(key) != value:
                    errors.append(name+': changed or removed '+category+' '+key)
            if category == 'fields':
                for key, value in after[category].items():
                    if key not in values and value['options']:
                        errors.append(name+': added field with required/custom semantics '+key)
    return errors


def snapshot(root):
    import http_contracts
    http_contracts.SOURCE = root/'contracts'
    http = {name: http_contracts.bundle(http_contracts.SOURCE/path) for name, path in http_contracts.DOCUMENTS.items()}
    fixtures = {p.relative_to(root).as_posix(): p.read_text() for p in root.glob('services/**/fixtures/*.command.hex')}
    return {'protobuf': proto_snapshot(root), 'http': http, 'fixtures': fixtures}


def checkout(ref, target):
    """Export only specifications and fixtures; never execute code from the baseline."""
    commit = subprocess.check_output(['git', 'rev-parse', '--verify', ref+'^{commit}'], cwd=ROOT, text=True).strip()
    names = subprocess.check_output(['git', 'ls-tree', '-r', '--name-only', commit, '--', 'contracts', 'services'], cwd=ROOT, text=True).splitlines()
    for name in names:
        if not (name.endswith(('.proto', '.command.hex')) or name.startswith('contracts/') and name.endswith('.json')):
            continue
        destination = target/name
        destination.parent.mkdir(parents=True, exist_ok=True)
        destination.write_bytes(subprocess.check_output(['git', 'show', commit+':'+name], cwd=ROOT))


def main():
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument('--against', default=os.environ.get('CAFE_CONTRACT_BASE') or BASELINE)
    args = parser.parse_args()
    from http_compatibility import changes
    with tempfile.TemporaryDirectory() as directory:
        baseline = Path(directory)
        checkout(args.against, baseline)
        old, new = snapshot(baseline), snapshot(ROOT)
    errors = proto_changes(old['protobuf'], new['protobuf'])
    for name, document in old['http'].items():
        errors += [name+': '+error for error in changes(document, new['http'][name])]
    for name, data in old['fixtures'].items():
        if new['fixtures'].get(name) != data:
            errors.append(name+': historical queued-command fixture changed or removed')
    if errors:
        raise SystemExit('Contract incompatibility against '+args.against+':\n'+'\n'.join(errors))
    print('Published interfaces and stored command fixtures are compatible with '+args.against)


if __name__ == '__main__':
    main()
