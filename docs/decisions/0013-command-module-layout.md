# 0013 — One command and handler per use-case file

Status: accepted, 9 October 2026. Extends decisions 0011 and 0012.

## Decision

All six contexts organise command execution under `application/commands/`.
Each file contains one named command DTO and its matching `CommandHandler`.
The filename describes the business action without repeating the enclosing
`commands` directory name:

```text
ordering/application/commands/
  create_order.go
  add_line.go
  change_quantity.go
  place_order.go
  errors.go
  ordering_bdd_test.go

preparation/application/commands/
  __init__.py
  accept_order.py
  start_preparation.py
  complete_preparation.py

loyalty/application/commands/
  credit-collection.ts
  issue-reward.ts
  redeem-reward.ts
```

Go's directory is a real `commands` package. Callers import it directly, using
context-qualified aliases where useful. TypeScript and Python callers import the
specific module. Python package initialisers contain package documentation;
they do not collect or re-export every use case. There are no compatibility
aliases in the previous application modules.

Shared command errors may have a separate file in the command package. An error
used by only one command can stay beside that handler. Shared application event
DTOs remain in their owning application layer, and projection/event handlers
remain outside `commands`. Commands may depend on these shared values; the parent
Go application package does not import or re-export its command subpackage.

Go and TypeScript tests stay beside their subjects. Go's context Gherkin bindings
stay together in the command package because their scenarios exercise sequences
of commands. Python tests retain the separate service test tree; new focused
command tests mirror `contexts/<context>/application/commands/` there.

## Consequences

Developers can locate, review and change one use case without scanning unrelated
handlers. The command shape and transaction orchestration remain visible together.
This adds small files and more explicit imports; it also keeps the shared context
and domain dependencies visible. It adds no dispatch abstraction or runtime bus.

The scaffold emits this layout. Compiler/AST checks reject multiple command pairs,
separated command/handler declarations and command declarations outside the
command folder. Go mutation checks recognise application subpackages. Existing
scenario and infrastructure tests continue to exercise the real moved handlers.
This is a source organisation change: published schemas, queue identities and
transaction behaviour retain their existing definitions.
