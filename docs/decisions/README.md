# Accepted architecture decisions

The guides describe the current working rules. Decisions preserve their rationale
and history; later decisions explicitly refine earlier choices. Start with
[the contributor guide](../../CONTRIBUTING.md) and
[the developer workflow](../developer-workflow.md) when implementing a change.

## Current rules from recent decisions

| Decision | Rule to apply |
| --- | --- |
| [0010 — Dependency injection](0010-dependency-injection-frameworks.md) | Fx, Dependency Injector and Awilix live in composition; handlers receive plain constructor ports |
| [0011 — Durable commands](0011-durable-commands-before-aggregate-mutations.md) | Every runtime aggregate business mutation starts in a named command handler; event reactions durably enqueue commands |
| [0012 — Enforced boundaries and workflow](0012-enforcement-and-developer-workflow.md) | One direct store invocation, no application handler chaining, exhaustive subscriptions, lifecycle cleanup and historical compatibility |
| [0013 — Command modules](0013-command-module-layout.md) | One command DTO and its matching handler per business-action file under `application/commands/` |
| [0014 — Queries and context navigation](0014-context-navigation-and-query-layout.md) | Named query pairs, application read models/ports, owned boundary mappings, one reaction per file and checked context navigation |
| [0016 — Owner messaging boundary](0016-owner-messaging-boundary.md) | Business HTTP lives in the API; owner processes expose health and diagnostics and receive business requests through RabbitMQ |
| [0015 — Persistence role names](0015-explicit-persistence-role-names.md) | Named read/write repositories, aggregate-owned facts and central command execution with infrastructure-owned transactions |

0011 narrows earlier event-handling examples: event reactions that change a root
now persist an owner command first. Projection handlers can update read data
without a command. 0012 prohibits synchronous command-handler calls from all
application handlers, including handlers in the same context. 0013–0014 change
source organisation while preserving published interfaces and durable identities.
0015 refines persistence names and separates restoration from read repository modules.

## Foundations and public boundaries

| Decision | Scope and later refinement |
| --- | --- |
| [0001 — Consistency and delivery](0001-consistency-and-delivery.md) | Aggregate boundaries, atomic receipts/outbox and current-state persistence; reaction entry points refined by 0011 |
| [0002 — Polyglot services and browser](0002-polyglot-and-browser.md) | Three domain services and browser delivery; API business dispatch refined by 0007 |
| [0003 — Executable behaviour](0003-executable-behaviour.md) | Context-owned Gherkin scenarios, native bindings and separate live workflow evidence |
| [0004 — Runtime authority and queries](0004-runtime-authority-and-queries.md) | Separate context databases, restricted credentials, migrations and complete/explicitly paginated reads |
| [0005 — Typed expected errors](0005-typed-expected-errors.md) | Named business rejections retain stable codes/messages; corruption and infrastructure failure remain retryable |
| [0006 — OpenAPI authority](0006-openapi-http-authority.md) | Source specifications, route coverage and conformance; request transport refined by 0007–0009 |
| [0007 — RabbitMQ API requests](0007-api-broker-requests.md) | HTTP-to-Protobuf anti-corruption mapping and API-to-owner RabbitMQ dispatch; direct owner HTTP exception retired by 0016 |
| [0008 — Contract ownership](0008-published-contract-ownership.md) | Context-first published sources; private facts/queued formats and bootstrap SQL stay with their owners |
| [0009 — Enforcement and command replies](0009-contract-enforcement-and-command-replies.md) | Owner request packages, required inputs, independent consumers, atomic reply bytes and frontend enforcement |

A proposed change to an accepted rule needs a decision explaining the new behaviour,
trade-offs, compatibility and enforcement. Update the current guides and relevant
checks in the same change. Preserve dated assessments and executed verification
as historical evidence; link subsequent refinements rather than silently rewriting
their original rationale.
