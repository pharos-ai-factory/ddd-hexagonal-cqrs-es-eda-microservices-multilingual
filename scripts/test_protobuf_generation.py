"""Exercise legal schema evolution through the real compiler and API generator."""
from pathlib import Path
import json
import shutil
import subprocess
import tempfile
import unittest
from contract_sources import ROOT
from generate_requests import generate_requests
from http_contracts import bundle, SOURCE, DOCUMENTS
from typed_http_boundaries import generate, definitions


class ProtobufGenerationTests(unittest.TestCase):
    def setUp(self):
        self.directory = tempfile.TemporaryDirectory()
        self.addCleanup(self.directory.cleanup)
        self.root = Path(self.directory.name)
        shutil.copytree(ROOT/'contracts', self.root/'contracts')
        self.mapping = self.root/'services/api/adaptors/http/backend/contract_mappings.json'
        self.mapping.parent.mkdir(parents=True)
        shutil.copyfile(ROOT/self.mapping.relative_to(self.root), self.mapping)
        self.commands = self.root/'contracts/menu/messaging/commands/v1/menu_commands.proto'
        self.document = bundle(SOURCE/DOCUMENTS['api'])

    def test_multiline_options_and_comments_preserve_required_input_mapping(self):
        expected = generate(self.root, self.document)
        self.commands.write_text(self.commands.read_text().replace(
            'optional string name = 1 [(cafe.requests.v1.required_input) = true];',
            'optional string name = 1 [\n /* A required public input. */\n (cafe.requests.v1.required_input)=true\n ];'))
        self.assertTrue(definitions(self.root)[('menu', 'CreateDrink')][0]['required'])
        self.assertEqual(generate(self.root, self.document), expected)

    def test_mapping_to_an_unknown_wire_field_fails_generation(self):
        mappings = json.loads(self.mapping.read_text())
        mappings['createDrink']['fields'] = {'missingWireField': 'name'}
        self.mapping.write_text(json.dumps(mappings))
        with self.assertRaisesRegex(ValueError, 'unknown wire input'):
            generate(self.root, self.document)

    def test_contexts_can_reuse_payload_tags_and_names(self):
        # Ordering already uses tag 17 and the name create_order. Both are local.
        self.commands.write_text(self.commands.read_text()+'\nmessage CreateOrder { }\n')
        envelope = self.root/'contracts/menu/messaging/v1/menu_requests.proto'
        envelope.write_text(envelope.read_text().replace(
            'cafe.menu.requests.v1.PublishEdition publish_edition = 16;',
            'cafe.menu.requests.v1.PublishEdition publish_edition = 16;\n cafe.menu.requests.v1.CreateOrder create_order = 17;'))
        standard = subprocess.check_output(['uv', 'run', '--no-project', '--with', 'grpcio-tools==1.84.0',
            'python', '-c', "import grpc_tools; from pathlib import Path; print(Path(grpc_tools.__file__).parent/'_proto')"], text=True).strip()
        generate_requests('api', self.root, ROOT/'.tools/bin/protoc-gen-go', standard)
        output = self.root/'services/api/adaptors/messaging/generated/cafe/requests/v1'
        self.assertFalse((output/'requests.pb.go').exists())
        for owner in ('menu', 'ordering'):
            self.assertIn('Command_CreateOrder', (output/f'contexts/{owner}/{owner}_requests.pb.go').read_text())
        # Compile both owner packages and the dispatch interfaces together.
        module = ROOT/'services/api'
        for name in ('go.mod', 'go.sum'):
            shutil.copyfile(module/name, self.root/name)
        (self.root/'adaptors/messaging').mkdir(parents=True)
        shutil.copytree(output.parents[2], self.root/'adaptors/messaging/generated', dirs_exist_ok=True)
        result = subprocess.run([str(ROOT/'.tools/go/bin/go') if (ROOT/'.tools/go/bin/go').exists() else shutil.which('go'), 'test', './adaptors/messaging/generated/...'],
                                cwd=self.root, text=True, capture_output=True)
        self.assertEqual(result.returncode, 0, result.stdout+result.stderr)


if __name__ == '__main__':
    unittest.main()
