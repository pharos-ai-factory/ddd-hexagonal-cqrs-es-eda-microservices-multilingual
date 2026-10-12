"""Compiler setup must work without a pre-existing developer tool cache."""
from pathlib import Path
import subprocess
import tempfile
import unittest
from unittest.mock import patch
from protobuf_tools import ensure_go_plugin

VERSION = 'protoc-gen-go v1.36.12\n'


class ProtobufToolTests(unittest.TestCase):
    def test_missing_compiler_is_installed_in_the_requested_checkout(self):
        with tempfile.TemporaryDirectory() as directory:
            root = Path(directory)
            with patch('protobuf_tools.subprocess.check_output', side_effect=[FileNotFoundError(), VERSION]), \
                 patch('protobuf_tools.subprocess.run') as install:
                plugin = ensure_go_plugin(root)
            self.assertEqual(plugin, root/'.tools/bin/protoc-gen-go')
            self.assertTrue(plugin.parent.is_dir())
            install.assert_called_once()
            arguments, options = install.call_args
            self.assertEqual(arguments[0][-2:], ['install', 'google.golang.org/protobuf/cmd/protoc-gen-go@v1.36.12'])
            self.assertEqual(options['env']['GOBIN'], str(plugin.parent))
            self.assertEqual(options['cwd'], root)
            self.assertTrue(options['check'])

    def test_matching_compiler_is_reused_without_installation(self):
        with patch('protobuf_tools.subprocess.check_output', return_value=VERSION), \
             patch('protobuf_tools.subprocess.run') as install:
            ensure_go_plugin()
        install.assert_not_called()

    def test_outdated_or_unusable_compiler_is_replaced(self):
        for current in ('protoc-gen-go v1.36.0', PermissionError(),
                        subprocess.CalledProcessError(1, ['protoc-gen-go', '--version'])):
            with self.subTest(current=current), tempfile.TemporaryDirectory() as directory, \
                 patch('protobuf_tools.subprocess.check_output', side_effect=[current, VERSION]), \
                 patch('protobuf_tools.subprocess.run') as install:
                ensure_go_plugin(Path(directory))
                install.assert_called_once()

    def test_failed_installation_is_reported_before_generation(self):
        with tempfile.TemporaryDirectory() as directory, \
             patch('protobuf_tools.subprocess.check_output', side_effect=FileNotFoundError()), \
             patch('protobuf_tools.subprocess.run', side_effect=subprocess.CalledProcessError(1, ['go', 'install'])):
            with self.assertRaises(subprocess.CalledProcessError):
                ensure_go_plugin(Path(directory))

    def test_installation_must_provide_the_pinned_executable(self):
        for installed in ('protoc-gen-go v1.36.0', FileNotFoundError(), PermissionError()):
            with self.subTest(installed=installed), tempfile.TemporaryDirectory() as directory, \
                 patch('protobuf_tools.subprocess.check_output', side_effect=[FileNotFoundError(), installed]), \
                 patch('protobuf_tools.subprocess.run'):
                with self.assertRaisesRegex(RuntimeError, 'did not provide protoc-gen-go v1.36.12'):
                    ensure_go_plugin(Path(directory))


if __name__ == '__main__':
    unittest.main()
