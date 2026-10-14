"""The explicit reference redesign leaves unrelated and later changes checked."""

from copy import deepcopy
import unittest
from compatibility import proto_changes
from event_redesign import accepted_baseline


def field(name, typ=9, oneof="", type_name="", options=""):
    return dict(name=name, type=typ, oneof=oneof, type_name=type_name, options=options)


class EventRedesignTests(unittest.TestCase):
    def test_only_the_retired_payload_union_and_added_rules_are_accepted(self):
        old = {
            "cafe.v1.Event": {
                "fields": {
                    "1": field("id"),
                    "22": field("order_placed", 11, "payload", ".cafe.v1.OrderPlaced"),
                }
            },
            "cafe.v1.OrderPlaced": {"fields": {"1": field("order_id")}},
        }
        new = deepcopy(old)
        del new["cafe.v1.Event"]["fields"]["22"]
        new["cafe.v1.Event"]["fields"]["12"] = field("payload", 12)
        new["cafe.v1.OrderPlaced"]["fields"]["1"]["options"] = "declared constraints"
        self.assertEqual(proto_changes(accepted_baseline(old, new), new), [])
        new["cafe.v1.Event"]["fields"]["1"]["name"] = "different"
        self.assertTrue(proto_changes(accepted_baseline(old, new), new))

    def test_modern_baselines_keep_full_comparison(self):
        old = {"cafe.v1.Event": {"fields": {"12": field("payload", 12)}}}
        new = deepcopy(old)
        new["cafe.v1.Event"]["fields"]["12"]["type"] = 9
        self.assertEqual(accepted_baseline(old, new), old)
        self.assertTrue(proto_changes(old, new))
