"""Create reviewable use-case starters without overwriting source or inventing domain policy."""
import argparse
import json
from pathlib import Path
import re
from developer import ROOT, CONTEXTS


def render(kind, context, name):
    service = next(key for key, owners in CONTEXTS.items() if context in owners)
    snake = re.sub(r'(?<!^)(?=[A-Z])', '_', name).lower()
    if not re.fullmatch(r'[A-Z][A-Za-z0-9]{1,60}', name):
        raise ValueError('Use a PascalCase business name, such as CancelOrder')
    role = {'command': 'CommandHandler', 'query': 'QueryHandler', 'subscription': 'IntegrationEventHandler'}[kind]
    if service == 'engagement':
        port = {'command': 'AggregateCommandPort<S>', 'query': 'QueryPort<S>', 'subscription': 'DurableCommandPort<C>'}[kind]
        generic = '<C extends object>' if kind == 'subscription' else '<S>'
        body = {'command': "return this.port.execute(metadata, () => { throw new Error('Define the aggregate transition'); });",
                'query': 'return this.port.get(id);', 'subscription': 'return this.port.enqueue(metadata, command);'}[kind]
        args = {'command': 'metadata: Metadata, command: '+name+'Command', 'query': 'id: string',
                'subscription': 'metadata: Metadata, command: C'}[kind]
        method = 'handle' if kind == 'subscription' else 'execute'
        imports = {'command': 'AggregateCommandPort, Metadata', 'query': 'QueryPort', 'subscription': 'DurableCommandPort, Metadata'}[kind]
        source = f"""import type {{{imports}}} from '../../../foundation/application.js';
/** Define the owner-local inputs for {name}. */
export type {name}Command = Readonly<{{}}>;
/** {name} owns one explicit application responsibility. */
export class {name}{role}{generic} {{
  constructor(private port: {port}) {{}}
  {method}({args}) {{ {body} }}
}}
"""
        test = f"import {{test}} from 'node:test';\nimport assert from 'node:assert/strict';\ntest('{name}: define its business outcome', () => {{ assert.fail('Replace with an aggregate or port assertion'); }});\n"
        files = {snake+'.ts': source, snake+'.test.ts': test}
        registration = f"// Add to the context's typed dependency record and container.register:\n{name[0].lower()+name[1:]}: asFunction(c => new {name}{role}(c.owningPort)).singleton(),\n"
    elif service == 'operations':
        port = {'command': 'AggregateCommandPort[S]', 'query': 'QueryPort[S]', 'subscription': 'DurableCommandPort[C]'}[kind]
        generic = '[C]' if kind == 'subscription' else '[S]'
        args = {'command': f'metadata: Metadata, command: {name}Command', 'query': 'identity: str',
                'subscription': 'metadata: Metadata, command: C'}[kind]
        result = 'Loaded[S] | None' if kind == 'query' else 'Outcome'
        body = {'command': '''def decide(state: S | None) -> Change[S]:
            raise NotImplementedError("Define the aggregate transition")
        return self.port.execute(metadata, decide)''', 'query': 'return self.port.get(identity)',
                'subscription': 'return self.port.enqueue(metadata, command)'}[kind]
        method = 'handle' if kind == 'subscription' else 'execute'
        imports = {'command': 'AggregateCommandPort, Metadata, Outcome, Change', 'query': 'QueryPort, Loaded', 'subscription': 'DurableCommandPort, Metadata, Outcome'}[kind]
        source = f'''from typing import TypedDict
from operations.foundation.application import {imports}

class {name}Command(TypedDict):
    """Define the owner-local inputs for {name}."""
    pass

class {name}{role}{generic}:
    """{name} owns one explicit application responsibility."""
    def __init__(self, port: {port}) -> None:
        self.port = port

    def {method}(self, {args}) -> {result}:
        {body}
'''
        test = f'def test_{snake}() -> None:\n    raise AssertionError("Replace with the expected business outcome")\n'
        files = {snake+'.py': source, 'test_'+snake+'.py': test}
        registration = f'# Add to the owning DeclarativeContainer:\n{snake} = providers.Singleton({name}{role}, owning_port)\n'
    else:
        port = {'command': 'a.AggregateCommandPort[S]', 'query': 'a.QueryPort[S]', 'subscription': ''}[kind]
        if kind == 'subscription':
            files = {snake+'.go': f'''package application
import (
 "context"
 a "github.com/pharos-ai-factory/ddd-hexagonal-cqrs-es-eda-microservices-multilingual/services/storefront/foundation/application"
)
// {name}Event describes the incoming fact; replace with the published application DTO.
type {name}Event struct {{}}
// {name}CommandQueue accepts the command durably before returning.
type {name}CommandQueue[C any] interface {{ Enqueue(context.Context, a.Metadata, C) (a.Outcome,error) }}
// {name}IntegrationEventHandler translates the fact into a durable owner command.
type {name}IntegrationEventHandler[C any] struct {{ Commands {name}CommandQueue[C] }}
func(h {name}IntegrationEventHandler[C]) Handle(ctx context.Context, m a.Metadata, event {name}Event) (a.Outcome,error) {{
 panic("Map the incoming fact, then return h.Commands.Enqueue(ctx, m, command)")
}}
'''}
        else:
            # Go query port is context-local; the starter makes that capability explicit.
            source = f'''package application

import (
 "context"
 a "github.com/pharos-ai-factory/ddd-hexagonal-cqrs-es-eda-microservices-multilingual/services/storefront/foundation/application"
)

// {name}Command defines owner-local intent.
type {name}Command struct {{}}
// {name}{role} owns one application responsibility.
type {name}{role}[S any] struct {{ Store {port} }}
'''
            if kind == 'command':
                source += f'''func (h {name}{role}[S]) Execute(ctx context.Context, m a.Metadata, command {name}Command) (a.Outcome, error) {{
 return h.Store.Execute(ctx, m, func(state a.Loaded[S]) (a.Mutation[S], error) {{ panic("Define the aggregate transition") }})
}}
'''
            else:
                source += f'''func (h {name}{role}[S]) Execute(ctx context.Context, id string) (a.Loaded[S], error) {{
 return h.Store.Get(ctx, id)
}}
'''
            files = {snake+'.go': source}
        files[snake+'_test.go'] = f'package application\nimport "testing"\nfunc Test{name}(t *testing.T) {{ t.Fatal("Replace with the expected business outcome") }}\n'
        registration = f'// Add a typed provider to the context Fx module and bind its owning port.\n// fx.Provide(new{name}{role})\n'
    if kind == 'query':
        for filename in list(files):
            files[filename] = files[filename].replace(name+'Command', name+'Query')
    if kind == 'command':
        arranged = {}
        for filename, content in files.items():
            if service == 'engagement':
                filename = filename.replace('_', '-')
                content = content.replace("'../../../foundation/", "'../../../../foundation/")
            if service == 'storefront':
                content = content.replace('package application', 'package commands')
            # Python keeps its native tests under the service tests tree.
            destination = filename if service == 'operations' and filename.startswith('test_') else 'commands/'+filename
            arranged[destination] = content
        files = arranged
        files['placement.txt'] = ('Place commands/ under the owning context application/ directory.\n'
            'Keep the command DTO and handler together; import their module directly.\n'
            'Place Python test modules under tests/contexts/<context>/application/commands/.\n')
    files['registration.txt'] = registration
    if kind == 'subscription':
        # Reactions use the command from its command/handler module.
        for filename, content in files.items():
            content = re.sub(r'/\*\* Define the owner-local inputs for [^\n]+\*/\nexport type '+name+r'Command = Readonly<\{\}>;\n', '', content)
            content = content.replace('from typing import TypedDict\n', '')
            content = re.sub(r'class '+name+r'Command\(TypedDict\):\n    """[^\n]+"""\n    pass\n\n', '', content)
            files[filename] = content
        files['subscription.json'] = json.dumps({'consumer': context+'.'+snake.replace('_','-'),
            'event': 'REPLACE_WITH_PUBLISHED_EVENT', 'command': name[0].lower()+name[1:]}, indent=2)+'\n'
        files['wiring.txt'] = ('Add the typed incoming event and owner command DTO. Map the event inside handle; the starter forwards an already mapped command.\n'
            'Add its private Protobuf payload with an unused field number, decoder and golden fixture.\n'
            'Add the manifest entry, typed codec and provider; register both consumers through commandSubscription/command_subscription.\n'
            'Run completeSubscriptions/complete_subscriptions, generation and the composition tests.\n'
            'For Go, introduce the owner durable command port/adaptor before adding an aggregate-changing subscription; existing Go subscriptions are projections.\n')
    return service, files


def main():
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument('kind', choices=('command','query','subscription'))
    parser.add_argument('context', choices=[owner for owners in CONTEXTS.values() for owner in owners])
    parser.add_argument('name')
    parser.add_argument('--output', type=Path)
    args = parser.parse_args()
    service, files = render(args.kind, args.context, args.name)
    target = args.output or ROOT/'.local/scaffolds'/args.context/args.name
    if target.exists():
        parser.error('Output already exists; select a fresh directory')
    target.mkdir(parents=True)
    for name, content in files.items():
        output = target/name
        output.parent.mkdir(parents=True, exist_ok=True)
        output.write_text(content)
    print(f'{service}/{args.context} starter: {target}\nReview the types and domain policy, then copy into the owner. Tests intentionally fail until the behaviour is specified.')


if __name__ == '__main__':
    main()
