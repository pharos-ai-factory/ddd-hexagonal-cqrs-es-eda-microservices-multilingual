"""Compile only the request packages consumed by each independent service."""
from pathlib import Path
import shutil
import subprocess
import tempfile
from contract_sources import sources
from request_contracts import OWNED, envelopes


def generate_requests(service, root, plugin, standard_protos):
    owners = OWNED[service]
    definitions = {name: path for name, path in sources('requests', root).items()
                   if '/contexts/' not in name or name.split('/contexts/')[1].split('/')[0] in owners}
    go_relative = ('services/api/adaptors/messaging/generated' if service == 'api' else
                   'services/storefront/contracts/requests/generated')
    output = root/go_relative if service != 'operations' else root/'services/operations/src/operations/adaptors/generated'
    output.mkdir(parents=True, exist_ok=True)
    if service != 'operations':
        for path in output.rglob('*.go'):
            path.unlink()
    with tempfile.TemporaryDirectory() as directory:
        staging = Path(directory)
        for name, source in definitions.items():
            target = staging/name
            target.parent.mkdir(parents=True, exist_ok=True)
            shutil.copyfile(source, target)
        inputs = list(definitions)
        arguments = ['uv', 'run', '--no-project', '--with', 'grpcio-tools==1.84.0', 'python', '-m',
                     'grpc_tools.protoc', '-I'+str(staging), '-I'+standard_protos]
        if service == 'operations':
            arguments += ['--python_out='+str(output), '--pyi_out='+str(output)]
        else:
            package = 'github.com/pharos-ai-factory/ddd-hexagonal-cqrs-es-eda-microservices-multilingual/'+go_relative
            arguments += ['--plugin=protoc-gen-go='+str(plugin), '--go_out='+str(output), '--go_opt=module='+package]
            if service == 'api':
                for name in inputs:
                    owner = name.split('/contexts/')[1].split('/')[0] if '/contexts/' in name else None
                    arguments.append('--go_opt=M'+name+'='+package+'/'+str(Path(name).parent)+('/shared' if Path(name).stem in ('common','validation') else '')+';'+
                                     (owner+'v1' if owner else ('sharedv1' if Path(name).stem in ('common','validation') else 'requestsv1')))
        subprocess.run([*arguments, *inputs], cwd=root, check=True)
    if service != 'operations':
        target = output/'cafe/requests/v1/envelopes.go'
        target.write_text(envelopes(owners, package))
        subprocess.run(['python3', str(Path(__file__).with_name('go.py')), 'gofmt', '-w', str(target)], cwd=root, check=True)
