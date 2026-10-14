"""Handwritten shared codecs cannot grow business event registries again."""

import unittest
from event_boundaries import SHARED_CODECS, violation


class EventBoundaryTests(unittest.TestCase):
    def test_shared_event_registries_fail_in_every_language(self):
        for path in SHARED_CODECS:
            with self.subTest(path=path):
                self.assertIsNotNone(violation(path, '"ordering.order-placed"', ["ordering.order-placed"]))

    def test_context_mappings_and_technical_checks_have_their_own_roles(self):
        self.assertIsNone(
            violation(
                "services/operations/src/operations/contexts/preparation/adaptors/messaging/order_placed_payload.py",
                '"ordering.order-placed"',
                ["ordering.order-placed"],
            )
        )
        for path in SHARED_CODECS:
            self.assertIsNone(violation(path, '"Unknown event contract or owner"', ["ordering.order-placed"]))
