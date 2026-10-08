"""Generate frontend HTTP operation types from the assembled OpenAPI boundary."""
import json
from pathlib import Path

OUTPUT = Path('services/web/src/adaptors/generated/http.ts')


def schema_type(schema: dict) -> str:
    if '$ref' in schema:
        name = schema['$ref'].removeprefix('#/components/schemas/')
        result = 'Schemas['+json.dumps(name)+']'
    elif 'enum' in schema:
        result = ' | '.join(json.dumps(value) for value in schema['enum'])
    elif 'oneOf' in schema or 'anyOf' in schema:
        result = ' | '.join('('+schema_type(item)+')' for item in schema.get('oneOf', schema.get('anyOf')))
    elif 'allOf' in schema:
        result = ' & '.join('('+schema_type(item)+')' for item in schema['allOf'])
    elif schema.get('type') == 'array':
        result = 'Array<'+schema_type(schema['items'])+'>'
    elif schema.get('type') == 'object' or 'properties' in schema:
        required = schema.get('required', [])
        properties = [json.dumps(name)+('' if name in required else '?')+': '+schema_type(value)
                      for name, value in schema.get('properties', {}).items()]
        additional = schema.get('additionalProperties', False)
        if additional:
            properties.append('[key: string]: '+(schema_type(additional) if isinstance(additional, dict) else 'unknown'))
        result = '{'+'; '.join(properties)+'}' if properties else 'Record<string, never>'
    else:
        try:
            result = {'string': 'string', 'integer': 'number', 'number': 'number',
                      'boolean': 'boolean', None: 'unknown'}[schema.get('type')]
        except KeyError as error:
            raise ValueError('Unsupported frontend schema: '+str(schema)) from error
    return '('+result+') | null' if schema.get('nullable') else result


def generated(document: dict) -> str:
    lines = ['// Code generated from the API OpenAPI specification. DO NOT EDIT.',
             '// Source: contracts/services/api/http_api/api.openapi.json',
             'export type Schemas = {']
    lines += ['  '+json.dumps(name)+': '+schema_type(value)+';'
              for name, value in sorted(document.get('components', {}).get('schemas', {}).items())]
    lines += ['};', 'export type Operations = {']
    routes, commands, lists = {}, [], []
    for path, item in sorted(document['paths'].items()):
        for method, operation in sorted(item.items()):
            if not isinstance(operation, dict) or 'operationId' not in operation:
                continue
            name = operation['operationId']
            if name in routes:
                raise ValueError('Duplicate HTTP operation ID: '+name)
            parameters = {'path': [], 'query': [], 'headers': []}
            required_parameters = set()
            resolved = {}
            for parameter in item.get('parameters', [])+operation.get('parameters', []):
                if '$ref' in parameter:
                    parameter = document['components']['parameters'][parameter['$ref'].rsplit('/', 1)[1]]
                # Fetch supplies Origin for the same-origin POSTs used by the browser.
                if method == 'post' and parameter['in'] == 'header' and parameter['name'].lower() == 'origin':
                    continue
                location = 'headers' if parameter['in'] == 'header' else parameter['in']
                if location not in parameters:
                    raise ValueError('Unsupported frontend parameter location: '+location)
                resolved[(location, parameter['name'])] = parameter
            for (location, _), parameter in resolved.items():
                if parameter.get('required'):
                    required_parameters.add(location)
                parameters[location].append(json.dumps(parameter['name'])+
                        ('' if parameter.get('required') else '?')+': '+schema_type(parameter['schema']))
            inputs = []
            for kind, values in parameters.items():
                inputs.append(kind+('' if kind in required_parameters else '?')+': '+
                              ('{'+'; '.join(values)+'}' if values else 'never'))
            if operation.get('requestBody'):
                body = operation['requestBody']
                inputs.append('body'+('' if body.get('required') else '?')+': '+
                              schema_type(body['content']['application/json']['schema']))
            else:
                inputs.append('body?: never')
            responses = []
            for status, response in sorted(operation['responses'].items()):
                if '$ref' in response:
                    response = document['components']['responses'][response['$ref'].rsplit('/', 1)[1]]
                content = response.get('content', {}).get('application/json')
                responses.append(status+': '+(schema_type(content['schema']) if content else 'undefined'))
            lines.append('  '+json.dumps(name)+': {input: {'+'; '.join(inputs)+'}; responses: {'+'; '.join(responses)+'}};')
            routes[name] = {'method': method.upper(), 'path': path}
            if method == 'post' and item.get('x-owner'):
                commands.append(name)
            if method == 'get' and name.startswith('list') and item.get('x-owner'):
                lists.append(name)
    lines += ['};', 'export const routes = '+json.dumps(routes, indent=2)+' as const;',
              'export type Operation = keyof Operations;',
              'export type Input<K extends Operation> = Operations[K]["input"];',
              'export type Success<K extends Operation> = Operations[K]["responses"][200];',
              'export type CommandOperation = '+' | '.join(json.dumps(name) for name in commands)+';',
              'export type ListOperation = '+' | '.join(json.dumps(name) for name in lists)+';',
              'export type CommandBody<K extends CommandOperation> = Input<K>["body"];',
              'export type Intersection<U> = (U extends unknown ? (value: U) => void : never) extends (value: infer I) => void ? I : never;',
              'export type CommandParameters = Intersection<{[K in CommandOperation]: Omit<Input<K>, "body">}[CommandOperation]>;',
              'export type SendCommand = <K extends CommandOperation>(operation: K, id: string, body: CommandBody<NoInfer<K>>, version: number) => Promise<boolean>;']
    return '\n'.join(lines)+'\n'
