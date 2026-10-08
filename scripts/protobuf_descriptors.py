"""Inspect compiler descriptors, including custom field options, for code generation."""
import json
from pathlib import Path
import shutil
import subprocess
import sys
import tempfile
from contract_sources import sources


def inspect(root):
    from google.protobuf import descriptor_pb2, descriptor_pool, message_factory
    import grpc_tools
    with tempfile.TemporaryDirectory() as directory:
        staging = Path(directory)
        definitions = sources('requests', root)
        for logical, source in definitions.items():
            target = staging/logical
            target.parent.mkdir(parents=True, exist_ok=True)
            shutil.copyfile(source, target)
        output = staging/'requests.pb'
        subprocess.run([sys.executable, '-m', 'grpc_tools.protoc', '-I'+str(staging),
                        '-I'+str(Path(grpc_tools.__file__).parent/'_proto'), '--include_imports',
                        '--descriptor_set_out='+str(output), *definitions], check=True)
        descriptors = descriptor_pb2.FileDescriptorSet.FromString(output.read_bytes())
    pool = descriptor_pool.DescriptorPool()
    for file in descriptors.file:
        pool.Add(file)
    options_type = message_factory.GetMessageClass(pool.FindMessageTypeByName('google.protobuf.FieldOptions'))
    required_input = pool.FindExtensionByName('cafe.requests.v1.required_input')
    result = []
    for file in descriptors.file:
        if file.name not in definitions:
            continue
        for message in file.message_type:
            fields = []
            for field in message.field:
                options = options_type.FromString(field.options.SerializeToString())
                typ = field.type_name.lstrip('.') if field.type_name else descriptor_pb2.FieldDescriptorProto.Type.Name(field.type).removeprefix('TYPE_').lower()
                fields.append({'name': field.name, 'json_name': field.json_name, 'number': field.number,
                               'type': typ, 'presence': 'repeated' if field.label == field.LABEL_REPEATED else 'optional' if field.proto3_optional else '',
                               'required': bool(options.Extensions[required_input])})
            result.append({'name': message.name, 'package': file.package, 'source': file.name, 'fields': fields})
    return result


def load(root):
    return json.loads(subprocess.check_output(['uv', 'run', '--no-project', '--with', 'grpcio-tools==1.84.0',
        'python', str(Path(__file__).resolve()), str(root)], text=True))


if __name__ == '__main__':
    print(json.dumps(inspect(Path(sys.argv[1]))))
