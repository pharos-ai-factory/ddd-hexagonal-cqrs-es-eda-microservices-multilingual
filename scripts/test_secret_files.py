"""Exercise persistence permissions, migration and developer file indirection."""
from pathlib import Path
import stat
import tempfile
import unittest
from unittest.mock import patch

import dev
from secret_files import atomic_private_write, compose_environment


class SecretFilesTest(unittest.TestCase):
    def test_configure_preserves_credentials_and_repairs_permissions(self):
        with tempfile.TemporaryDirectory() as directory:
            path = Path(directory) / "dev.env"
            original = dev.configure(path)
            path.chmod(0o644)
            self.assertEqual(original, dev.configure(path))
            self.assertEqual(stat.S_IMODE(path.stat().st_mode), 0o600)

    def test_symlinks_cannot_overwrite_another_file(self):
        with tempfile.TemporaryDirectory() as directory:
            target = Path(directory) / "target"
            target.write_text("preserve")
            link = Path(directory) / "dev.env"
            link.symlink_to(target)
            with self.assertRaises(ValueError):
                atomic_private_write(link, "replacement")
            self.assertEqual(target.read_text(), "preserve")

    def test_failed_atomic_write_preserves_original(self):
        with tempfile.TemporaryDirectory() as directory:
            path = Path(directory) / "dev.env"
            atomic_private_write(path, "original")
            with patch("secret_files.os.replace", side_effect=OSError("test failure")):
                with self.assertRaises(OSError):
                    atomic_private_write(path, "replacement")
            self.assertEqual(path.read_text(), "original")
            self.assertEqual(list(path.parent.iterdir()), [path])

    def test_external_credentials_are_never_copied_to_configuration(self):
        with tempfile.TemporaryDirectory() as directory:
            path = Path(directory) / "dev.env"
            key = Path(directory) / "developer.key"
            atomic_private_write(key, "billable-provider-key\n")
            atomic_private_write(path, "DELIVERY_KEY_FILE=developer.key\n")
            values = dev.configure(path)
            self.assertEqual(values["DELIVERY_KEY"], "billable-provider-key")
            self.assertNotIn("billable-provider-key", path.read_text())
            self.assertIn("DELIVERY_KEY_FILE=developer.key", path.read_text())
            self.assertNotIn("DELIVERY_KEY_FILE", values)
            self.assertEqual(dev.load(path)["DELIVERY_KEY"], "billable-provider-key")
            key.chmod(0o644)
            with self.assertRaisesRegex(ValueError, "private regular file"):
                dev.load(path)

    def test_external_file_errors_are_redacted(self):
        with tempfile.TemporaryDirectory() as directory:
            path = Path(directory) / "dev.env"
            for body in ("KEY_FILE=secret-path\n", "KEY=value\nKEY_FILE=secret-path\n"):
                atomic_private_write(path, body)
                with self.assertRaises(ValueError) as failure:
                    dev.load(path)
                self.assertNotIn("secret-path", str(failure.exception))
                self.assertNotIn("value", str(failure.exception))

    def test_derived_urls_encode_passwords(self):
        values = {"MENU_DB_PASSWORD": "a/b@c", "MENU_BROKER_PASSWORD": "x:y", "BROKER_PASSWORD": "p@ss", "API_BROKER_PASSWORD": "r:p"}
        environment = compose_environment(values, ["menu"])
        self.assertIn(":a%2Fb%40c@postgres:", environment["MENU_DATABASE_URL"])
        self.assertIn(":x%3Ay@rabbitmq:", environment["MENU_BROKER_URL"])
