"""Negative authority fixtures and catalogue coverage for broker provisioning."""

import copy
import json
import re
import unittest
from unittest.mock import patch

import dev
from broker_permissions import GENERATED, build, exact
from contract_sources import ROOT, OWNERS, catalogue


class BrokerPermissionsTests(unittest.TestCase):
    @classmethod
    def setUpClass(cls):
        cls.policy = build()

    def allows(self, user, operation, name):
        return bool(re.search(self.policy["users"]["cafe_" + user]["permissions"][operation], name))

    def test_committed_policy_matches_catalogues(self):
        self.assertEqual(json.loads((ROOT / GENERATED).read_text()), self.policy)

    def test_queue_roles_include_private_commands_retries_and_requests(self):
        for name in (
            "ref.loyalty.issue-reward",
            "ref.loyalty.issue-reward.retry",
            "ref.loyalty.issue-reward.dead",
            "ref.loyalty.issue-reward.command",
            "ref.loyalty.issue-reward.command.retry",
            "ref.loyalty.issue-reward.command.dead",
            "ref.loyalty.commands",
            "ref.loyalty.commands.dead",
            "ref.loyalty.queries",
            "ref.loyalty.queries.dead",
        ):
            self.assertTrue(self.allows("loyalty", "read", name), name)
        for owner in OWNERS:
            for queue in self.policy["queues"][owner]:
                self.assertTrue(self.allows(owner, "read", queue), queue)
                self.assertFalse(self.allows(owner, "write", queue), queue)
                self.assertFalse(self.allows(owner, "configure", queue), queue)
                for foreign in (*[other for other in OWNERS if other != owner], "api"):
                    self.assertFalse(self.allows(foreign, "read", queue), (foreign, queue))
            for name in (f"ref.{owner}.unregistered", f"ref.{owner}.delivery", f"ref.api.{owner}.replies"):
                self.assertFalse(self.allows(owner, "read", name), name)
            self.assertTrue(self.allows("api", "read", f"ref.api.{owner}.replies"))
        self.assertFalse(self.allows("api", "read", "ref.api.unknown.replies"))

    def test_resource_and_topic_permissions_deny_runtime_binding_and_spoofing(self):
        for owner in (*OWNERS, "api"):
            authority = self.policy["users"]["cafe_" + owner]
            for exchange in self.policy["exchanges"]:
                self.assertFalse(self.allows(owner, "read", exchange))
                self.assertFalse(self.allows(owner, "configure", exchange))
            self.assertFalse(self.allows(owner, "write", "amq.default"))
            for topic in authority["topics"]:
                self.assertEqual(topic["read"], "^$")
                for key in (
                    "#",
                    "domain.foreign.event",
                    "integration.foreign.event",
                    "request.foreign.command",
                    "reply.foreign",
                ):
                    self.assertIsNone(re.search(topic["write"], key))
        topics = {item["exchange"]: item["write"] for item in self.policy["users"]["cafe_menu"]["topics"]}
        self.assertRegex("domain.menu.drink-published", topics["cafe.events"])
        self.assertRegex("integration.menu.edition-published", topics["cafe.events"])
        self.assertRegex("reply.menu", topics["cafe.replies"])
        self.assertNotRegex("reply.menu.extra", topics["cafe.replies"])

    def test_topology_has_exact_resources_and_required_dead_letter_authority(self):
        queues = {name for names in self.policy["queues"].values() for name in names}
        exchanges = set(self.policy["exchanges"])
        self.assertFalse(queues & exchanges)
        for name in queues | exchanges:
            self.assertTrue(self.allows("topology", "configure", name))
            self.assertEqual(
                self.allows("topology", "read", name),
                name in exchanges or name in self.policy["dead_letter_sources"],
            )
            self.assertEqual(
                self.allows("topology", "write", name), name in queues or name.endswith(".delivery")
            )
        for permission in ("configure", "read", "write"):
            self.assertFalse(self.allows("topology", permission, "ref.menu.unregistered"))
        for topic in self.policy["users"]["cafe_topology"]["topics"]:
            self.assertEqual(topic["write"], "^$")
            self.assertNotRegex("#", topic["read"])
            self.assertNotRegex("domain.menu.unregistered", topic["read"])

    def test_colliding_queue_and_exchange_names_are_rejected(self):
        events = catalogue("topology")
        for consumer in (
            "menu.delivery",
            "menu.commands",
            "menu.drink-directory.retry",
            "menu.drink-directory",
            "loyalty.issue-reward.command",
        ):
            entries = [*events, {**events[0], "consumer": consumer}]
            with self.subTest(consumer=consumer), patch("broker_permissions.catalogue", return_value=entries):
                with self.assertRaisesRegex(ValueError, "resource name collision"):
                    build()

    def test_wildcard_consumer_identity_is_rejected(self):
        events = copy.deepcopy(catalogue("topology"))
        events[0]["consumer"] = "menu.*"
        with patch("broker_permissions.catalogue", return_value=events):
            with self.assertRaisesRegex(ValueError, "consumer identity"):
                build()

    def test_regex_names_are_literal_and_anchored(self):
        pattern = exact(["ref.owner.a+b", "ref.owner.a.b"])
        self.assertRegex("ref.owner.a+b", pattern)
        for name in ("refXownerXaXb", "ref.owner.ab", "prefix.ref.owner.a.b", "ref.owner.a.b.suffix"):
            self.assertNotRegex(name, pattern)

    def test_provisioning_uses_generated_policy_and_distinct_untagged_users(self):
        values = {item["password_key"]: name for name, item in self.policy["users"].items()}
        with patch("dev.broker_request") as request:
            dev.broker_users(values)
        self.assertEqual(request.call_count, len(values) * 5)
        for name, authority in self.policy["users"].items():
            request.assert_any_call(values, "PUT", f"users/{name}", {"password": name, "tags": ""})
            request.assert_any_call(values, "PUT", f"permissions/reference/{name}", authority["permissions"])
            for topic in authority["topics"]:
                request.assert_any_call(values, "PUT", f"topic-permissions/reference/{name}", topic)
