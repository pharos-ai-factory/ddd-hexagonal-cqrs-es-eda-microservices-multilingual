"""Initialise the pinned Go Protobuf compiler for generation and its tests."""
import os
from pathlib import Path
import subprocess
import sys

ROOT = Path(__file__).resolve().parents[1]
GO_PLUGIN_VERSION = 'v1.36.12'


def ensure_go_plugin(root=ROOT):
    """Return an executable compiler of the required version, installing it when needed."""
    plugin = root/'.tools/bin/protoc-gen-go'
    expected = 'protoc-gen-go '+GO_PLUGIN_VERSION

    def installed_version():
        try:
            return subprocess.check_output([str(plugin), '--version'], text=True).strip()
        except (OSError, subprocess.CalledProcessError):
            return None

    if installed_version() != expected:
        plugin.parent.mkdir(parents=True, exist_ok=True)
        subprocess.run([sys.executable, str(root/'scripts/go.py'), 'install',
                        'google.golang.org/protobuf/cmd/protoc-gen-go@'+GO_PLUGIN_VERSION],
                       cwd=root, env=dict(os.environ, GOBIN=str(plugin.parent)), check=True)
        if installed_version() != expected:
            raise RuntimeError(f'Go Protobuf compiler installation did not provide {expected}: {plugin}')
    return plugin
