"""Selection, dependency closure and isolated historical cross-language evidence."""

import json
import os
from pathlib import Path
import shutil
import subprocess
import tempfile
import unittest
from unittest.mock import patch
from contract_sources import ROOT
from contract_selection import declarations, owners
from protobuf_selection import plan
from protobuf_tools import ensure_go_plugin
from generate_selected import generate_selected
import http_contracts


class ContractSelectionTests(unittest.TestCase):
    def test_categories_keep_browser_and_api_boundaries_explicit(self):
        self.assertEqual(owners("storefront", "events"), ("menu", "ordering"))
        self.assertEqual(owners("operations", "events"), ("ordering", "preparation", "collection"))
        self.assertEqual(owners("engagement", "events"), ("collection", "loyalty"))
        self.assertEqual(owners("web", "events"), ())
        self.assertEqual(owners("web", "requests"), ())
        self.assertEqual(len(owners("web", "realtime")), 6)
        self.assertEqual(owners("api", "events"), ())
        self.assertEqual(owners("api", "realtime"), ())
        self.assertEqual(len(owners("api", "requests")), 6)
        self.assertEqual(declarations()["openapi"]["web"]["consumes"], ["api"])

    def test_invalid_selection_fails_closed(self):
        with tempfile.TemporaryDirectory() as directory:
            root = Path(directory)
            shutil.copytree(ROOT / "contracts", root / "contracts")
            original = declarations(root)
            for change in (
                lambda c: c["events"]["operations"]["consumes"].append("typo"),
                lambda c: c["events"]["operations"]["consumes"].append("ordering"),
                lambda c: c["events"].update(typo={"produces": [], "consumes": []}),
                lambda c: c["events"].update(web={"produces": [], "consumes": ["menu"]}),
                lambda c: c["requests"].update(web={"produces": [], "consumes": ["menu"]}),
                lambda c: c["events"]["operations"].update(consume=[]),
                lambda c: c.pop("realtime"),
            ):
                candidate = json.loads(json.dumps(original))
                change(candidate)
                (root / "contracts/generation.json").write_text(json.dumps(candidate))
                with self.assertRaises(ValueError):
                    declarations(root)

    def test_selected_descriptor_preserves_dependencies_and_wire_identities(self):
        script = """
from google.protobuf import descriptor_pb2, descriptor_pool
from protobuf_selection import compile_descriptors, select
from contract_selection import roots
full = compile_descriptors('events')
selected = select(full, 'events', roots('events', ['menu']))
files = {f.name: f for f in selected.file}
assert set(files) == {'cafe/v1/events.proto', 'cafe/v1/event_options.proto', 'google/protobuf/descriptor.proto', 'cafe/v1/contexts/menu/menu_integration_events.proto'}
event = files['cafe/v1/events.proto'].message_type[0]
assert not event.oneof_decl
assert next(f for f in event.field if f.name == 'payload').type == 12
original = next(f for f in full.file if f.name == 'cafe/v1/events.proto').message_type[0]
for field in event.field:
    assert field == next(f for f in original.field if f.name == field.name)
assert all('/contexts/' not in dep for dep in files['cafe/v1/events.proto'].dependency)
# A selected payload can depend transitively on an unselected schema.
menu = next(f for f in full.file if '/menu/' in f.name)
ordering = next(f for f in full.file if '/ordering/' in f.name)
menu.dependency.append(ordering.name)
menu.message_type[-1].field.add(name='source', number=50, type=11, type_name='.cafe.v1.OrderPlaced', label=1)
selected = select(full, 'events', roots('events', ['menu']))
assert ordering.name in {f.name for f in selected.file}
pool = descriptor_pool.DescriptorPool()
for file in selected.file:
    # The synthetic dependency changes topological order: add it first.
    if file.name == menu.name:
        pool.Add(ordering)
    pool.Add(file)
assert pool.FindMessageTypeByName('cafe.v1.MenuPublished').fields_by_name['source'].message_type.name == 'OrderPlaced'
assert 'order_placed' not in pool.FindMessageTypeByName('cafe.v1.Event').fields_by_name
assert not select(full, 'events', []).file
requests = select(compile_descriptors('requests'), 'requests', roots('requests', ['collection']))
request_files = {file.name for file in requests.file}
assert 'cafe/requests/v1/common.proto' in request_files
assert 'cafe/requests/v1/validation.proto' in request_files
assert 'google/protobuf/descriptor.proto' in request_files
assert all('/collection/' in name for name in request_files if '/contexts/' in name)
"""
        result = subprocess.run(
            ["uv", "run", "--no-project", "--with", "grpcio-tools==1.84.0", "python", "-c", script],
            cwd=ROOT,
            env=dict(os.environ, PYTHONPATH=str(ROOT / "scripts")),
            capture_output=True,
            text=True,
            check=False,
        )
        self.assertEqual(result.returncode, 0, result.stdout + result.stderr)

    def test_runtime_outputs_exclude_unrelated_payloads(self):
        self.assertFalse((ROOT / "services/web/src/adaptors/generated/events.json").exists())
        self.assertFalse((ROOT / "services/web/src/adaptors/generated/requests.json").exists())
        for owner in ("preparation", "collection", "loyalty"):
            self.assertFalse(
                (
                    ROOT
                    / f"services/storefront/contracts/events/generated/cafe/v1/{owner}_integration_events.pb.go"
                ).exists()
            )
        for owner in ("menu", "loyalty"):
            self.assertFalse(
                (
                    ROOT
                    / f"services/operations/src/operations/adaptors/generated/cafe/v1/contexts/{owner}/{owner}_integration_events_pb2.py"
                ).exists()
            )
        for service, expected in (
            ("engagement", {"account", "reward", "notification"}),
            ("web", {"drink", "edition", "order", "ticket", "pickup", "account", "reward", "notification"}),
        ):
            document = json.loads(
                (ROOT / f"services/{service}/src/adaptors/generated/realtime.json").read_text()
            )
            namespace = document["nested"]["cafe"]["nested"]["realtime"]["nested"]["v1"]["nested"]
            self.assertEqual(set(namespace["Publication"]["oneofs"]["snapshot"]["oneof"]), expected)
            if service == "engagement":
                self.assertNotIn("Drink", namespace)
                self.assertNotIn("Ticket", namespace)

    def test_openapi_selection_and_stale_output_cleanup(self):
        with tempfile.TemporaryDirectory() as directory:
            root = Path(directory)
            shutil.copytree(ROOT / "contracts", root / "contracts")
            config = declarations(root)
            config["openapi"] = {"web": {"produces": [], "consumes": ["api"]}}
            (root / "contracts/generation.json").write_text(json.dumps(config))
            stale = root / http_contracts.OUTPUTS["storefront"] / "storefront.openapi.json"
            stale.parent.mkdir(parents=True)
            stale.write_text("{}")
            with (
                patch.object(http_contracts, "ROOT", root),
                patch.object(http_contracts, "SOURCE", root / "contracts"),
            ):
                with self.assertRaisesRegex(ValueError, "Stale OpenAPI"):
                    http_contracts.generate(check=True)
                http_contracts.generate()
                self.assertFalse(stale.exists())
                self.assertTrue((root / http_contracts.FRONTEND_OUTPUT).exists())
                self.assertFalse((root / http_contracts.MAPPING_OUTPUT).exists())
                http_contracts.generate(check=True)

    def test_regeneration_is_reproducible_and_prunes_removed_targets(self):
        with tempfile.TemporaryDirectory() as directory:
            root = Path(directory)
            shutil.copytree(ROOT / "contracts", root / "contracts")
            (root / "scripts").mkdir()
            shutil.copyfile(ROOT / "scripts/descriptor_json.mjs", root / "scripts/descriptor_json.mjs")
            (root / "node_modules").symlink_to(ROOT / "node_modules", target_is_directory=True)
            plugin = ensure_go_plugin(ROOT)
            generate_selected(plugin, root)

            def contents():
                return {
                    p.relative_to(root): p.read_bytes() for p in (root / "services").rglob("*") if p.is_file()
                }

            before = contents()
            self.assertFalse((root / "services/web/src/adaptors/generated/events.json").exists())
            generate_selected(plugin, root)
            self.assertEqual(before, contents())
            marker = root / "services/operations/src/operations/adaptors/generated/catalogue.json"
            marker.write_text("preserve metadata")
            config = declarations(root)
            config["events"].pop("operations")
            config["realtime"].pop("storefront")
            (root / "contracts/generation.json").write_text(json.dumps(config))
            generate_selected(plugin, root)
            self.assertFalse(list((marker.parent / "cafe/v1").rglob("*_pb2.*")))
            self.assertFalse(
                list((root / "services/storefront/contracts/realtime/generated").rglob("*.pb.go"))
            )
            self.assertTrue(list((root / "services/storefront/contracts/events/generated").rglob("*.pb.go")))
            self.assertTrue(list((marker.parent / "cafe/realtime").rglob("*_pb2.py")))
            self.assertEqual(marker.read_text(), "preserve metadata")

    def test_reward_fixture_uses_dedicated_go_and_python_outputs(self):
        # Full historical coverage must not require a foreign runtime contract.
        plugin = ensure_go_plugin(ROOT)
        with tempfile.TemporaryDirectory() as directory:
            root = Path(directory)
            descriptor = root / "reward.pb"
            names = plan("events", ["loyalty"], descriptor)
            python, go = root / "python", root / "go"
            python.mkdir()
            go.mkdir()
            arguments = [
                "uv",
                "run",
                "--no-project",
                "--with",
                "grpcio-tools==1.84.0",
                "python",
                "-m",
                "grpc_tools.protoc",
                "--descriptor_set_in=" + str(descriptor),
                "--python_out=" + str(python),
                "--plugin=protoc-gen-go=" + str(plugin),
                "--go_out=" + str(go),
                "--go_opt=paths=source_relative",
            ]
            arguments += ["--go_opt=M" + name + "=compatibility/wire;wire" for name in names]
            subprocess.run([*arguments, *names], check=True, cwd=ROOT, capture_output=True)
            wire = go / "wire"
            wire.mkdir()
            for path in list(go.rglob("*.pb.go")):
                path.rename(wire / path.name)
            (go / "go.mod").write_text(
                "module compatibility\n\ngo 1.27.1\n\nrequire google.golang.org/protobuf v1.36.12\n"
            )
            shutil.copyfile(ROOT / "services/storefront/go.sum", go / "go.sum")
            fixture = ROOT / "contracts/loyalty/messaging/integration_events/v1/fixtures/reward-issued.v1.hex"
            (go / "wire/compatibility_test.go").write_text("""package wire
import ("encoding/hex"; "os"; "strings"; "bytes"; "testing"; "google.golang.org/protobuf/proto")
func TestHistoricalReward(t *testing.T) {
 text, err := os.ReadFile(os.Getenv("REWARD_FIXTURE")); if err != nil { t.Fatal(err) }
 body, err := hex.DecodeString(strings.TrimSpace(string(text))); if err != nil { t.Fatal(err) }
 event := new(Event); if err = proto.Unmarshal(body, event); err != nil { t.Fatal(err) }
 payload := new(RewardIssued); if err=proto.Unmarshal(event.Payload,payload); err!=nil {t.Fatal(err)}
 if payload.GetBenefit() != "one free drink" { t.Fatal(event) }
 encoded, err := proto.MarshalOptions{Deterministic:true}.Marshal(event)
 if err != nil || !bytes.Equal(encoded, body) { t.Fatal("historical bytes changed", err) }
}
""")
            tool = (
                str(ROOT / ".tools/go/bin/go") if (ROOT / ".tools/go/bin/go").exists() else shutil.which("go")
            )
            result = subprocess.run(
                [tool, "test", "./wire"],
                cwd=go,
                env=dict(os.environ, REWARD_FIXTURE=str(fixture)),
                text=True,
                capture_output=True,
                check=False,
            )
            self.assertEqual(result.returncode, 0, result.stdout + result.stderr)
            script = """from pathlib import Path
from cafe.v1.events_pb2 import Event
import sys
body = bytes.fromhex(Path(sys.argv[1]).read_text())
event = Event.FromString(body)
from cafe.v1.contexts.loyalty.loyalty_integration_events_pb2 import RewardIssued
assert RewardIssued.FromString(event.payload).expires_at == '2026-10-03T12:00:00Z'
assert event.SerializeToString(deterministic=True) == body
"""
            result = subprocess.run(
                [
                    "uv",
                    "run",
                    "--no-project",
                    "--with",
                    "grpcio-tools==1.84.0",
                    "python",
                    "-c",
                    script,
                    str(fixture),
                ],
                cwd=root,
                env=dict(os.environ, PYTHONPATH=str(python)),
                text=True,
                capture_output=True,
                check=False,
            )
            self.assertEqual(result.returncode, 0, result.stdout + result.stderr)


if __name__ == "__main__":
    unittest.main()
