"""Runtime realtime credentials must not administer the session authority."""
import os
import subprocess
import sys
import unittest
from pathlib import Path
from uuid import uuid4

ROOT = Path(__file__).resolve().parents[2]
sys.path.insert(0, str(ROOT/"scripts"))
from dev import compose, load


class RealtimePermissions(unittest.TestCase):
    @classmethod
    def setUpClass(cls):
        cls.env_file = Path(os.environ["CAFE_ENV_FILE"])
        values = load(cls.env_file)
        if not values["COMPOSE_PROJECT_NAME"].startswith("cafe-reference-test-"):
            raise RuntimeError("Permission probes require the disposable integration project")

    def command(self, *arguments: str) -> str:
        # Read the existing credential inside its container; never print it.
        result = compose(self.env_file, "exec", "-T", "realtime-history", "/bin/sh", "-c",
                         'export REDISCLI_AUTH=$(cat "$REALTIME_REDIS_PASSWORD_FILE"); '
                         'exec valkey-cli --raw --user realtime --no-auth-warning "$@"', "valkey-cli", *arguments,
                         stdout=subprocess.PIPE, stderr=subprocess.STDOUT, text=True)
        return result.stdout.strip()

    def test_realtime_cannot_write_session_keys(self):
        key = "cafe:auth:permission-probe:"+str(uuid4())
        self.assertIn("NOPERM", self.command("SET", key, "probe"))

    def test_realtime_cannot_grant_itself_session_access(self):
        identity = "disabled-permission-probe-"+str(uuid4())
        result = self.command("ACL", "SETUSER", identity, "off")
        try:
            self.assertIn("NOPERM", result,
                          "The realtime login can administer ACLs and override its own key restrictions")
        finally:
            if result == "OK":
                self.command("ACL", "DELUSER", identity)

    def test_scripts_cannot_write_session_keys(self):
        key = "cafe:auth:permission-probe:"+str(uuid4())
        self.assertIn("NOPERM", self.command("EVAL", "return redis.call('HSET', KEYS[1], 'probe', 'value')", "1", key))

    def test_realtime_can_use_only_its_history_keys(self):
        key = "cafe:realtime.permission-probe:"+str(uuid4())
        try:
            self.assertEqual(self.command("HSET", key, "offset", "1"), "1")
            self.assertEqual(self.command("HGET", key, "offset"), "1")
        finally:
            self.command("DEL", key)

    def test_realtime_identity_cannot_authenticate_to_sessions(self):
        result = compose(self.env_file, "exec", "-T", "realtime-history", "/bin/sh", "-c",
                         'export REDISCLI_AUTH=$(cat "$REALTIME_REDIS_PASSWORD_FILE"); '
                         'exec valkey-cli -h valkey --raw --user realtime --no-auth-warning PING',
                         stdout=subprocess.PIPE, stderr=subprocess.STDOUT, text=True)
        self.assertIn("WRONGPASS", result.stdout)


if __name__ == "__main__":
    unittest.main()
