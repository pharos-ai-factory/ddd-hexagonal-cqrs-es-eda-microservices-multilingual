"""Named query, read-model and reader-port starters in each language's native layout."""


def render_query(service, context, name, snake):
    if service == 'engagement':
        stem = snake.replace('_', '-')
        files = {
            f'read-models/{stem}.ts': f'''/** Select the application read fields for {name}. */
export type {name}View = Readonly<{{id: string}}>;
''',
            f'ports/{stem}-reader.ts': f'''import type {{Loaded}} from '../../../../foundation/application.js';
import type {{{name}View}} from '../read-models/{stem}.js';
/** Read capability required by {name}; implement it in the owning adaptor. */
export interface {name}Reader {{
  get(id: string): Promise<Loaded<{name}View> | undefined>;
}}
''',
            f'queries/{stem}.ts': f'''import type {{{name}Reader}} from '../ports/{stem}-reader.js';
/** Inputs to the {name} read use case. */
export type {name}Query = Readonly<{{id: string}}>;
/** Execute {name} through its application-owned read capability. */
export class {name}QueryHandler {{
  constructor(private reader: {name}Reader) {{}}
  execute(query: {name}Query) {{ return this.reader.get(query.id); }}
}}
''',
            f'queries/{stem}.test.ts': f'''import {{test}} from 'node:test';
import assert from 'node:assert/strict';
test('{name}: specify the observable read result', () => {{ assert.fail('Define missing, present and failure outcomes'); }});
''',
        }
        registration = f'{name[0].lower()+name[1:]}: asFunction(c => new {name}QueryHandler(c.owningReader)).singleton(),\n'
    elif service == 'operations':
        prefix = f'operations.contexts.{context}.application'
        files = {
            f'read_models/{snake}.py': f'''from typing import TypedDict

class {name}View(TypedDict):
    """Select the application read fields for {name}."""
    id: str
''',
            f'ports/{snake}_reader.py': f'''from typing import Protocol
from operations.foundation.application import Loaded
from {prefix}.read_models.{snake} import {name}View

class {name}Reader(Protocol):
    """Read capability required by {name}; implement it in the owning adaptor."""
    def get(self, identity: str) -> Loaded[{name}View] | None: ...
''',
            f'queries/{snake}.py': f'''from dataclasses import dataclass
from operations.foundation.application import Loaded
from {prefix}.ports.{snake}_reader import {name}Reader
from {prefix}.read_models.{snake} import {name}View

@dataclass(frozen=True)
class {name}Query:
    """Inputs to the {name} read use case."""
    identity: str

class {name}QueryHandler:
    """Execute {name} through its application-owned read capability."""
    def __init__(self, reader: {name}Reader) -> None:
        self.reader = reader

    def execute(self, query: {name}Query) -> Loaded[{name}View] | None:
        return self.reader.get(query.identity)
''',
            f'test_{snake}.py': f'def test_{snake}() -> None:\n    raise AssertionError("Define missing, present and failure outcomes")\n',
        }
        registration = f'{snake} = providers.Singleton({name}QueryHandler, owning_reader)\n'
    else:
        prefix = 'github.com/pharos-ai-factory/ddd-hexagonal-cqrs-es-eda-microservices-multilingual/services/storefront'
        files = {
            f'readmodels/{snake}.go': f'''package readmodels
// {name}View selects the application read fields for {name}.
type {name}View struct {{ ID string `json:"id"` }}
''',
            f'ports/{snake}_reader.go': f'''package ports
import (
 "context"
 a "{prefix}/foundation/application"
 view "{prefix}/contexts/{context}/application/readmodels"
)
// {name}Reader supplies the read capability required by {name}.
type {name}Reader interface {{ Get(context.Context,string) (a.Loaded[view.{name}View],error) }}
''',
            f'queries/{snake}.go': f'''package queries
import (
 "context"
 a "{prefix}/foundation/application"
 "{prefix}/contexts/{context}/application/ports"
 view "{prefix}/contexts/{context}/application/readmodels"
)
// {name}Query identifies the requested resource.
type {name}Query struct {{ ID string }}
// {name}QueryHandler executes {name} through its owning read capability.
type {name}QueryHandler struct {{ Reader ports.{name}Reader }}
func(h {name}QueryHandler) Execute(ctx context.Context,q {name}Query) (a.Loaded[view.{name}View],error) {{
 return h.Reader.Get(ctx,q.ID)
}}
''',
            f'queries/{snake}_test.go': f'package queries\nimport "testing"\nfunc Test{name}(t *testing.T) {{ t.Fatal("Define missing, present and failure outcomes") }}\n',
        }
        registration = f'// Add to the owning Fx module:\nfunc(reader ports.{name}Reader) queries.{name}QueryHandler {{ return queries.{name}QueryHandler{{Reader:reader}} }},\n'
    files['registration.txt'] = registration
    files['placement.txt'] = ('Place queries/, ports/ and read models under the owning application/ directory.\n'
        'Reuse an existing reader/view when it supplies this use case; customise these placeholders otherwise.\n'
        'Place Python tests under tests/contexts/<context>/application/queries/ and add documentation-only package initialisers.\n'
        'Implement the reader and transport-to-query mapping under the owning adaptors/.\n'
        'Bind the reader and handler in composition and add the API/Protobuf operation when published.\n')
    return files
