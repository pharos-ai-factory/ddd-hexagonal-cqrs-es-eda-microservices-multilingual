#!/usr/bin/env python3
"""Run the real browser lane with pinned Chromium and system libraries in Docker."""
import json
import os
from pathlib import Path
import subprocess
import sys

ROOT = Path(__file__).resolve().parents[1]
security = json.loads(subprocess.check_output(
    ['docker', 'info', '--format', '{{json .SecurityOptions}}'], text=True))
command = ['docker', 'run', '--rm', '--network', 'host', '--ipc', 'host']
if 'name=rootless' not in security:
    command += ['--user', f'{os.getuid()}:{os.getgid()}']
env_path = Path(os.environ.get('CAFE_ENV_FILE', '.local/dev.env')).resolve().relative_to(ROOT)
command += ['-e', 'CAFE_ENV_FILE='+str(env_path), '-v', str(ROOT)+':/work', '-w', '/work',
            'mcr.microsoft.com/playwright:v1.63.0-noble',
            'node', 'node_modules/@playwright/test/cli.js', 'test', *sys.argv[1:]]
subprocess.run(command, cwd=ROOT, check=True)
