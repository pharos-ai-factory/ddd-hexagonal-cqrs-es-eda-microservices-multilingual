"""Format explicit handwritten paths, or the changed files when no paths are supplied."""
import argparse
from pathlib import Path
import subprocess
import sys
from developer import ROOT, run


def main():
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument('--check', action='store_true')
    parser.add_argument('paths', nargs='*')
    args = parser.parse_args()
    names = args.paths or subprocess.check_output(['git', 'diff', '--name-only', 'HEAD'], cwd=ROOT, text=True).splitlines()
    groups = {'go': [], 'python': [], 'prettier': []}
    for name in names:
        path = (ROOT/name).resolve()
        if not path.is_file() or not path.is_relative_to(ROOT) or 'generated' in path.parts:
            continue
        if path.suffix == '.go': groups['go'].append(str(path))
        elif path.suffix == '.py': groups['python'].append(str(path))
        elif path.suffix in ('.ts', '.tsx', '.mjs', '.json', '.yaml', '.yml', '.css', '.md') and not path.name.endswith('lock.yaml'):
            groups['prettier'].append(str(path))
    if groups['go']:
        command = [sys.executable, 'scripts/go.py', 'gofmt', '-l' if args.check else '-w', *groups['go']]
        if args.check:
            result = subprocess.check_output(command, cwd=ROOT, text=True)
            if result.strip(): raise SystemExit('Run pnpm format for:\n'+result)
        else: run(command)
    if groups['python']:
        run(['uv', 'run', '--project', 'services/operations', '--frozen', 'ruff', 'format', *(['--check'] if args.check else []), *groups['python']])
    if groups['prettier']:
        run(['pnpm', 'exec', 'prettier', '--check' if args.check else '--write', *groups['prettier']])


if __name__ == '__main__': main()
