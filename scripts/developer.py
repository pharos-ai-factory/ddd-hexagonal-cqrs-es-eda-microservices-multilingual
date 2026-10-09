"""Focused development commands; full verification remains a separate hand-off gate."""
import argparse
import json
import os
from pathlib import Path
import shutil
import socket
import subprocess
import sys
from dev import ROOT, PORTS, load

CONTEXTS = {'storefront': ('menu', 'ordering'), 'operations': ('preparation', 'collection'),
            'engagement': ('loyalty', 'communication'), 'api': (), 'web': ()}


def run(args, cwd=ROOT, env=None):
    return subprocess.run(args, cwd=cwd, env=env, check=True)


def test(service, context=None):
    if context and context not in CONTEXTS[service]:
        raise ValueError('Context does not belong to service')
    folder = ROOT/'services'/service
    if service in ('storefront', 'api'):
        run([sys.executable, 'scripts/go.py', 'test', '-count=1', './contexts/'+context+'/...' if context else './...'],
            env=dict(os.environ, CAFE_GO_PROJECT='services/'+service))
    elif service == 'operations':
        selected = ['tests/bdd/test_'+context+'.py', 'tests/test_domain.py'] if context else ['tests']
        run(['uv', 'run', '--frozen', 'pytest', '-q', *selected], cwd=folder)
    else:
        files = sorted(str(p.relative_to(folder)) for p in (folder/('src/contexts/'+context if context else 'src')).rglob('*.test.ts'))
        if files:
            run(['pnpm', 'exec', 'tsx', '--test', *files], cwd=folder)
        if context:
            run(['pnpm', 'test:bdd', '--tags', '@'+context], cwd=folder)


def doctor():
    checks = {}
    for tool, args in {'docker': ['compose', 'version'], 'node': ['--version'], 'pnpm': ['--version'],
                       'uv': ['--version'], 'python3': ['--version']}.items():
        if not shutil.which(tool):
            checks[tool] = 'missing'
        else:
            result = subprocess.run([tool, *args], capture_output=True, text=True)
            checks[tool] = result.stdout.strip() if result.returncode == 0 else 'unavailable'
    docker = subprocess.run(['docker', 'info', '--format', '{{.ServerVersion}}'], capture_output=True, text=True) if shutil.which('docker') else None
    checks['dockerEngine'] = 'ready' if docker and docker.returncode == 0 else 'unavailable'
    checks['go'] = 'local' if shutil.which('go') or (ROOT/'.tools/go/bin/go').exists() else 'pinned Docker fallback'
    checks['workspaceDependencies'] = 'installed' if (ROOT/'node_modules').exists() else 'run pnpm install --frozen-lockfile'
    config = ROOT/'.local/dev.env'
    try:
        values = load(config) if config.exists() else {}
        checks['configuration'] = 'readable' if values else 'created by pnpm dev:up'
    except Exception:
        checks['configuration'] = 'invalid; check file permissions and secret references'
        values = {}
    checks['ports'] = {}
    for key, default in PORTS.items():
        with socket.socket() as sock:
            try:
                sock.bind(('127.0.0.1', int(values.get(key, default))))
                checks['ports'][key] = 'available'
            except OSError:
                checks['ports'][key] = 'in use (may be the running development stack)'
    print(json.dumps(checks, indent=2))
    return 1 if any(v in ('missing', 'unavailable') for v in checks.values() if isinstance(v, str)) else 0


def watch(services, env_file):
    from dev import up
    up(env_file)
    from secret_files import compose_environment
    from dev import OWNERS
    run(['docker', 'compose', '--env-file', str(env_file), '-f', 'devops/compose.yaml',
         '-f', 'devops/watch.yaml', 'watch', *services], env=compose_environment(load(env_file), OWNERS))


def main():
    parser = argparse.ArgumentParser(description=__doc__)
    sub = parser.add_subparsers(dest='action', required=True)
    tests = sub.add_parser('test')
    tests.add_argument('service', nargs='?', choices=CONTEXTS)
    tests.add_argument('--context')
    sub.add_parser('doctor')
    watching = sub.add_parser('watch')
    watching.add_argument('services', nargs='*')
    watching.add_argument('--env-file', type=Path, default=ROOT/'.local/dev.env')
    args = parser.parse_args()
    if args.action == 'doctor':
        raise SystemExit(doctor())
    if args.action == 'watch':
        if any(name not in ('storefront', 'operations', 'engagement', 'api', 'ingress') for name in args.services):
            parser.error('Unknown development service')
        watch(args.services, args.env_file)
    else:
        if args.context and not args.service:
            parser.error('--context requires a service')
        for service in [args.service] if args.service else CONTEXTS:
            test(service, args.context)


if __name__ == '__main__':
    main()
