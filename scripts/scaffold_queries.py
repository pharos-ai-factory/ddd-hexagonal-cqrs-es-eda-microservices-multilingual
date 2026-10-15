"""Query starters consume generated input, response and resource contracts."""

from scaffold_inputs import generated_inputs, MODULE


def render_query(service, context, name, snake):
    action = name[0].lower() + name[1:]
    if service == "engagement":
        stem = snake.replace("_", "-")
        files = {
            f"read-models/{stem}.ts": f"export type {{ {name}View }} from '../../../../contracts/requests/generated/{context}-responses.js';\n",
            f"ports/{stem}-read-repository.ts": f"""import type {{Loaded}} from '../../../../foundation/application.js';
import type {{{name}View}} from '../read-models/{stem}.js';
/** Read capability required by {name}. */
export interface {name}ReadRepository {{ get(id: string): Promise<Loaded<{name}View> | undefined>; }}
""",
            f"queries/{stem}.ts": f"""import type {{{name}Query}} from '../../../../contracts/requests/generated/{context}.js';
import type {{{name}Response}} from '../../../../contracts/requests/generated/{context}-responses.js';
import type {{{name}ReadRepository}} from '../ports/{stem}-read-repository.js';
/** Apply the missing-resource policy and preserve the stored revision. */
export class {name}QueryHandler {{
  constructor(private readRepository: {name}ReadRepository) {{}}
  async execute(query: {name}Query): Promise<{name}Response> {{
    const value = await this.readRepository.get(query.id);
    return value ? {{ {action}: value }} : {{error: {{code: 'not_found', message: ''}}}};
  }}
}}
""",
            f"queries/{stem}.test.ts": f"import {{test}} from 'node:test';\nimport assert from 'node:assert/strict';\ntest('{name}: specify observable behaviour', () => {{ assert.fail('Define missing, present and failure outcomes'); }});\n",
        }
        registration = f"{action}: asFunction(c => new {name}QueryHandler(c.owningReadRepository)).singleton(),\nresponses.register{name}(registry, wire.{action}, c.{action});\n"
    elif service == "operations":
        prefix = f"operations.contexts.{context}.application"
        files = {
            f"read_models/{snake}.py": f"from operations.contracts.requests.generated.{context}_responses import {name}View as {name}View\n",
            f"ports/{snake}_read_repository.py": f'''from typing import Protocol
from operations.foundation.application import Loaded
from {prefix}.read_models.{snake} import {name}View
class {name}ReadRepository(Protocol):
    """Read capability required by {name}."""
    def get(self, identity: str) -> Loaded[{name}View] | None: ...
''',
            f"queries/{snake}.py": f'''from operations.contracts.requests.generated.{context} import {name}Query
from operations.contracts.requests.generated.{context}_responses import {name}Response, Loaded{name}
from {prefix}.ports.{snake}_read_repository import {name}ReadRepository
class {name}QueryHandler:
    """Apply missing-resource policy and preserve the stored revision."""
    def __init__(self, read_repository: {name}ReadRepository) -> None:
        self.read_repository = read_repository
    def execute(self, query: {name}Query) -> {name}Response:
        value = self.read_repository.get(query.id)
        if value is None:
            return {{"error": {{"code": "not_found", "message": ""}}}}
        return {{"{action}": Loaded{name}(exists=value["exists"], version=value["version"], state=value["state"])}}
''',
            f"test_{snake}.py": f'def test_{snake}() -> None:\n    raise AssertionError("Define missing, present and failure outcomes")\n',
        }
        registration = f"{snake} = providers.Singleton({name}QueryHandler, owning_read_repository)\nresponses.register_{snake}(registry, wire.{snake}, container.{snake}().execute)\n"
    else:
        files = {
            f"readmodels/{snake}.go": f'package readmodels\nimport model "{MODULE}/contracts/requests/generated/model/{context}"\n// {name}View is the generated resource representation.\ntype {name}View = model.{name}View\n',
            f"ports/{snake}_read_repository.go": f'''package ports
import ("context"; a "{MODULE}/foundation/application"; view "{MODULE}/contexts/{context}/application/readmodels")
// {name}ReadRepository supplies the read capability required by {name}.
type {name}ReadRepository interface {{ Get(context.Context,string) (a.Loaded[view.{name}View],error) }}
''',
            f"queries/{snake}.go": f'''package queries
import ("context"; "{MODULE}/contexts/{context}/application/ports"; model "{MODULE}/contracts/requests/generated/model/{context}"; shared "{MODULE}/contracts/requests/generated/model/shared")
// {name}QueryHandler applies missing-resource policy and preserves revisions.
type {name}QueryHandler struct {{ ReadRepository ports.{name}ReadRepository }}
func(h {name}QueryHandler) Execute(ctx context.Context,q model.{name}Query) (model.{name}Response,error) {{
 value, err := h.ReadRepository.Get(ctx,q.ID)
 if err != nil {{ return model.{name}Response{{}}, err }}
 if !value.Exists {{ return model.{name}Response{{Error: &shared.RequestError{{Code: "not_found"}}}}, nil }}
 return model.{name}Response{{{name}: &model.Loaded{name}{{Exists: value.Exists, Version: value.Version, State: value.State}}}}, nil
}}
''',
            f"queries/{snake}_test.go": f'package queries\nimport "testing"\nfunc Test{name}(t *testing.T) {{ t.Fatal("Define missing, present and failure outcomes") }}\n',
        }
        registration = f"// Add to the owning Fx module:\nfunc(reader ports.{name}ReadRepository) queries.{name}QueryHandler {{ return queries.{name}QueryHandler{{ReadRepository:reader}} }},\nresponses.Register{name}(registry, wire.{name}, handler.Execute)\n"
    files["registration.txt"] = (
        registration
        + "Import the generated owner response bindings (Go: adaptors/messaging/generated/queryregistration). Keep semantic input validation in the owner parser.\n"
    )
    files["placement.txt"] = (
        "Place queries, read ports and aliases under application/. Implement stored-state validation in adaptors/persistence. Review the generated response and missing-resource policy. Place Python tests under tests/contexts/<context>/application/queries/.\n"
    )
    files[
        "replies.proto.txt"
    ] = f"""// Add to the owner's reply schema. Allocate an unused Reply.payload field for Loaded{name} {snake}.
import "cafe/requests/v1/validation.proto";
message {name}View {{ string id = 1; }}
message Loaded{name} {{
  bool exists = 1;
  uint64 version = 2;
  {name}View state = 3 [(cafe.requests.v1.required_output) = true];
}}
"""
    return generated_inputs(files, service, context, name, "Query")
