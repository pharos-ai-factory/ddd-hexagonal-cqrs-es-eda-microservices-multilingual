"""Reproduce every generated contract in isolation and fail on any file drift."""
from hashlib import sha256
import os
from pathlib import Path
import shutil
import subprocess
import sys
import tempfile
from check_architecture import IGNORED

ROOT = Path(__file__).resolve().parents[1]


def hashes(root):
    result = {}
    for base in ('contracts', 'services'):
        for directory, children, names in os.walk(root/base):
            children[:] = [name for name in children if name not in IGNORED]
            for name in names:
                path = Path(directory)/name
                result[path.relative_to(root).as_posix()] = sha256(path.read_bytes()).hexdigest()
    return result


def check():
    with tempfile.TemporaryDirectory() as directory:
        target = Path(directory)
        for name in ('contracts', 'services', 'scripts', 'devops'):
            shutil.copytree(ROOT/name, target/name, ignore=lambda _, names: [name for name in names if name in IGNORED])
        (target/'node_modules').symlink_to(ROOT/'node_modules', target_is_directory=True)
        if (ROOT/'.tools').is_dir():
            (target/'.tools').symlink_to(ROOT/'.tools', target_is_directory=True)
        before = hashes(target)
        subprocess.run([sys.executable, 'scripts/generate.py'], cwd=target, check=True)
        after = hashes(target)
        changed = sorted(name for name in before.keys() | after.keys() if before.get(name) != after.get(name))
        if changed:
            raise SystemExit('Generated contract drift; run pnpm generate:contracts:\n'+'\n'.join(changed))
    print('All generated contracts match their sources, including added and removed files')


if __name__ == '__main__':
    check()
