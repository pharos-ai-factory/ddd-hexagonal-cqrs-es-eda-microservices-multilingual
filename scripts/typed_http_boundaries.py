"""Compile explicit API ACL mappings into typed Go request/reply conversions."""
import json
from protobuf_descriptors import load
from pathlib import Path

OUTPUT = Path('services/api/adaptors/http/backend/generated/mapping.go')
MODULE = 'github.com/pharos-ai-factory/ddd-hexagonal-cqrs-es-eda-microservices-multilingual/services/api/adaptors/messaging/generated/cafe/requests/v1'


def title(value):
    return ''.join(word[:1].upper()+word[1:] for word in value.split('_'))


def camel(value):
    result = title(value)
    return result[:1].lower()+result[1:]


def definitions(root):
    result = {}
    for message in load(root):
        parts = message['package'].split('.')
        if len(parts) != 4 or not message['source'].endswith(('_commands.proto', '_queries.proto', '_replies.proto')):
            continue
        owner = parts[1]
        for field in message['fields']:
            field['type'] = field['type'].removeprefix(message['package']+'.')
        result[(owner, message['name'])] = message['fields']
    return result


def generate(root, document):
    mappings = json.loads((root/'services/api/adaptors/http/backend/contract_mappings.json').read_text())
    messages = definitions(root)
    lines = ['// Code generated from OpenAPI, Protobuf and explicit ACL mappings. DO NOT EDIT.',
             'package generated', 'import ("encoding/json"; "fmt"; "google.golang.org/protobuf/proto";', 'pb "'+MODULE+'";']
    for owner in ('menu', 'ordering', 'preparation', 'collection', 'loyalty', 'communication'):
        lines.append(owner+' "'+MODULE+'/contexts/'+owner+'";')
    lines.append(')')
    operations = {}
    for path, item in document['paths'].items():
        owner = item.get('x-owner')
        if not owner:
            continue
        for method, operation in item.items():
            if isinstance(operation, dict) and 'operationId' in operation:
                operations[operation['operationId']] = (owner, method, operation)
    if set(operations) != set(mappings):
        raise ValueError('OpenAPI operations need explicit ACL mappings: '+str(set(operations)^set(mappings)))
    lines.append('func Supported(operation string) bool { switch operation {')
    lines.append('case '+','.join(json.dumps(name) for name in sorted(operations))+': return true }; return false }')
    lines.append('func Command(operation string, data []byte, metadata *pb.CommandMetadata) (pb.CommandEnvelope,error) { switch operation {')
    go_types = {'string': 'string', 'int64': 'int64', 'uint64': 'uint64', 'int32': 'int32', 'uint32': 'uint32', 'bool': 'bool'}
    for name, (owner, method, operation) in operations.items():
        if method != 'post':
            continue
        binding = mappings[name]
        schema = operation['requestBody']['content']['application/json']['schema']
        if '$ref' in schema:
            schema = document['components']['schemas'][schema['$ref'].rsplit('/', 1)[1]]
        fields = messages[(owner, binding['message'])]
        if set(binding['fields']) - {field['json_name'] for field in fields}:
            raise ValueError(name+': mapping names an unknown wire input')
        if set(schema.get('properties', {})) != set(binding['fields'].values()):
            raise ValueError(name+': every HTTP input needs an explicit wire mapping')
        lines += ['case '+json.dumps(name)+':', 'var input struct {']
        assignment = []
        for field in fields:
            if field['presence'] == 'repeated' or field['type'] not in go_types:
                raise ValueError(name+': unsupported command field: '+field['name'])
            wire_name = camel(field['name'])
            http_name = binding['fields'].get(wire_name)
            if http_name is None:
                if field['required']:
                    raise ValueError(name+': required wire input has no HTTP mapping: '+wire_name)
                continue
            prop = schema['properties'][http_name]
            expected = {'string':'string', 'bool':'boolean'}.get(field['type'], 'integer')
            if prop['type'] != expected:
                raise ValueError(name+': HTTP and wire scalar types need an explicit conversion: '+http_name)
            required = http_name in schema.get('required', [])
            if field['required'] != required:
                raise ValueError(name+': HTTP and wire input presence disagree: '+http_name)
            typ = go_types[field['type']]
            pointer = field['presence'] == 'optional'
            dto_type = '*'+typ if pointer and not required else typ
            go_name = title(field['name'])
            lines.append(go_name+' '+dto_type+' `json:"'+http_name+'"`;')
            assignment.append(go_name+': '+('&' if pointer and required else '')+'input.'+go_name)
        lines += ['}', 'if err:=json.Unmarshal(data,&input);err!=nil{return nil,err}',
                  'return &'+owner+'.Command{Metadata:metadata,Payload:&'+owner+'.Command_'+binding['message']+'{'+binding['message']+':&'+owner+'.'+binding['message']+'{'+','.join(assignment)+'}}},nil']
    lines += ['}; return nil,fmt.Errorf("unknown command operation") }',
              'func Query(operation,id string,page *pb.PageRequest)(proto.Message,error){switch operation {']
    for name, (owner, method, _) in operations.items():
        if method != 'get':
            continue
        message = mappings[name]['message']
        value = 'Id:id' if name.startswith('get') else 'Page:page'
        lines.append('case '+json.dumps(name)+':return &'+owner+'.Query{Payload:&'+owner+'.Query_'+message+'{'+message+':&'+owner+'.'+message+'{'+value+'}}},nil')
    lines += ['};return nil,fmt.Errorf("unknown query operation")}',
              'func QueryReply(request pb.Request,reply pb.Reply)(any,bool,*string,error){switch request:=request.(type){']
    for owner in ('menu', 'ordering', 'preparation', 'collection', 'loyalty', 'communication'):
        lines += ['case *'+owner+'.Request:',
                  'reply,ok:=reply.(*'+owner+'.Reply);if !ok||request.GetQuery()==nil{return nil,false,nil,fmt.Errorf("reply owner disagrees with query")}',
                  'switch value:=request.GetQuery().Payload.(type){']
        for name, (operation_owner, method, _) in operations.items():
            if method != 'get' or owner != operation_owner:
                continue
            message, output = mappings[name]['message'], mappings[name]['reply']
            loaded = title(output)
            lines += ['case *'+owner+'.Query_'+message+':',
                      'result,ok:=reply.Payload.(*'+owner+'.Reply_'+loaded+');if !ok||result.'+loaded+'==nil{return nil,false,nil,fmt.Errorf("reply disagrees with query") }']
            if name.startswith('get'):
                lines += ['item:=result.'+loaded+';if !item.Exists||item.State==nil||item.State.Id!=value.'+message+'.Id{return nil,false,nil,fmt.Errorf("query identity mismatch")}',
                          'object,err:=map'+owner+'Loaded'+loaded+'(item);return object,false,nil,err']
            else:
                lines += ['items:=make([]any,0,len(result.'+loaded+'.Items));',
                          'for _,item:=range result.'+loaded+'.Items { object,err:=map'+owner+'Loaded'+title(output.rstrip('s'))+'(item);if err!=nil{return nil,false,nil,err};items=append(items,object) }',
                          'paged:=value.'+message+'.Page!=nil;if result.'+loaded+'.Paged!=paged{return nil,false,nil,fmt.Errorf("page shape mismatch")}',
                          'return items,paged,result.'+loaded+'.NextId,nil']
        lines.append('}')
    lines += ['};return nil,false,nil,fmt.Errorf("unknown query payload") }']
    reply_messages = {}
    for message in load(root):
        if message['source'].endswith('_replies.proto'):
            reply_messages.setdefault(message['package'].split('.')[1], set()).add(message['name'])
    # Each DTO converter names every field and nested conversion explicitly.
    for (owner, name), fields in messages.items():
        source = root/f'contracts/{owner}/messaging/queries/v1/{owner}_replies.proto'
        if not source.exists() or name not in reply_messages[owner]:
            continue
        function = 'map'+owner+name
        lines += ['func '+function+'(value *'+owner+'.'+name+')(map[string]any,error){',
                  'if value==nil{return nil,fmt.Errorf("missing '+name+' DTO")};result:=map[string]any{}']
        for field in fields:
            typ, prop, go_name = field['type'], camel(field['name']), title(field['name'])
            if typ.startswith('cafe.requests.') or prop in ('paged','nextId'):
                continue
            if field['presence'] == 'repeated':
                lines += ['items'+go_name+':=make([]any,0,len(value.'+go_name+'));',
                          'for _,child:=range value.'+go_name+'{item,err:=map'+owner+typ+'(child);if err!=nil{return nil,err};items'+go_name+'=append(items'+go_name+',item)}',
                          'result['+json.dumps(prop)+']=items'+go_name]
            elif typ not in go_types:
                schema = document['components']['schemas'].get(owner+'.http_api.schemas.'+name)
                optional = field['presence'] == 'optional' or (schema is not None and prop not in schema.get('required', []))
                if optional:
                    lines.append('if value.'+go_name+'!=nil {')
                lines += ['item'+go_name+',err:=map'+owner+typ+'(value.'+go_name+');if err!=nil{return nil,err}',
                          'result['+json.dumps(prop)+']=item'+go_name]
                if optional:
                    lines.append('}')
            elif field['presence'] == 'optional':
                lines.append('if value.'+go_name+'!=nil{result['+json.dumps(prop)+']=*value.'+go_name+'}')
            else:
                lines.append('result['+json.dumps(prop)+']=value.'+go_name)
        if name.startswith('Loaded'):
            lines.append('if !value.Exists||value.State==nil{return nil,fmt.Errorf("invalid loaded DTO")}')
        lines += ['return result,nil}']
    return '\n'.join(lines)+'\n'
