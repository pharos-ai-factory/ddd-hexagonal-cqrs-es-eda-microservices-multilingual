"""Negative fixtures preserve the shared adaptor responsibility boundaries."""

import unittest
from adaptor_layout import import_violation, layout_violation

PREFIX = "services/operations/src/operations/adaptors/"


class AdaptorLayoutTests(unittest.TestCase):
    def test_flat_transport_modules_are_rejected(self):
        for name in ("delivery.py", "codec.py", "requests.py", "internal_commands.py"):
            with self.subTest(name=name):
                self.assertIsNotNone(layout_violation(PREFIX + name))
        for name in ("command_execution.py", "diagnostics.py", "messaging/event_codec.py"):
            self.assertIsNone(layout_violation(PREFIX + name))
        self.assertIsNone(layout_violation("services/engagement/src/adaptors/messaging/event-codec.ts"))

    def test_network_delivery_and_database_access_have_separate_owners(self):
        for role, target in (
            ("persistence", "pika"),
            ("persistence", "urllib/request"),
            ("messaging", "psycopg/Connection"),
            ("realtime", "psycopg_pool"),
        ):
            with self.subTest(role=role, target=target):
                self.assertIsNotNone(import_violation(PREFIX + role + "/adaptor.py", target))
        self.assertIsNone(import_violation(PREFIX + "messaging/rabbitmq_consumer.py", "pika"))
        self.assertIsNone(import_violation(PREFIX + "persistence/postgres_database.py", "psycopg"))
        self.assertIsNone(
            import_violation(
                PREFIX + "messaging/rabbitmq_outbox_relay.py",
                "operations/adaptors/persistence/dispatch_leases",
            )
        )

    def test_every_service_uses_responsibility_packages(self):
        for prefix in (
            "services/engagement/src/adaptors/",
            "services/storefront/adaptors/",
            "services/api/adaptors/",
            "services/web/src/adaptors/",
        ):
            with self.subTest(prefix=prefix):
                self.assertIsNotNone(layout_violation(prefix + "codec.ts"))
                self.assertIsNone(
                    layout_violation(prefix + "messaging/request_codec.go")
                    if "/web/" not in prefix
                    else layout_violation(prefix + "http/operation-client.ts")
                )

    def test_context_roots_expose_boundary_roles(self):
        for prefix in (
            "services/storefront/contexts/menu/adaptors/",
            "services/operations/src/operations/contexts/preparation/adaptors/",
            "services/engagement/src/contexts/communication/adaptors/",
        ):
            self.assertIsNotNone(layout_violation(prefix + "query_endpoints.py"))
            self.assertIsNone(layout_violation(prefix + "messaging/ticket_query_endpoints.py"))
            self.assertIsNone(layout_violation(prefix + "delivery/http-notification-delivery.ts"))

    def test_typescript_and_go_keep_delivery_out_of_persistence(self):
        for path, target in (
            ("services/engagement/src/adaptors/persistence/outbox.ts", "amqplib"),
            ("services/engagement/src/adaptors/messaging/relay.ts", "pg"),
            ("services/storefront/adaptors/persistence/postgres/", "net/http"),
            ("services/storefront/adaptors/messaging/workers/", "github.com/jackc/pgx/v5"),
        ):
            self.assertIsNotNone(import_violation(path, target))

    def test_shared_go_adaptors_cannot_acquire_owner_policy(self):
        self.assertIsNotNone(
            import_violation(
                "services/storefront/adaptors/messaging/requests/",
                "services/storefront/contexts/menu/adaptors/messaging",
            )
        )
