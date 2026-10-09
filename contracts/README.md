# Published contracts

Start with the context, then the communication boundary:

```text
contracts/
  menu/
    messaging/
      v1/                  owner request/reply envelopes
      commands/v1/
      queries/v1/
      integration_events/v1/
    realtime/v1/
    http_api/
  ordering/
  preparation/
  collection/
  loyalty/
  communication/
  shared/
    messaging/v1/
    realtime/v1/
    http_api/
  services/
    api/http_api/
    storefront/http_api/
    realtime_gateway/http_api/
    notification_provider/http_api/
```

Each of the six contexts follows the same structure. Create a category when the
context publishes that interface; Communication currently publishes queries but
no integration events or externally callable commands.

| Boundary | Context-owned content | Guide |
| --- | --- | --- |
| `messaging/v1` | Owner request/reply envelopes and packages | [Commands and queries](shared/messaging/requests.md) |
| `messaging/commands` | Published command input messages | [Commands and queries](shared/messaging/requests.md) |
| `messaging/queries` | Query inputs, published result DTOs and replies | [Commands and queries](shared/messaging/requests.md) |
| `messaging/integration_events` | Events published to other contexts and ownership metadata | [Integration events](shared/messaging/events.md) |
| `realtime` | Browser projection snapshots | [Browser updates](shared/realtime/README.md) |
| `http_api` | OpenAPI resource paths and JSON schemas | [HTTP APIs](shared/http_api/README.md) |

`shared/` contains technical envelopes, headers and reusable HTTP components.
`services/` contains complete assembled OpenAPI documents and service-owned
session, operational, gateway and provider endpoints. Generated service-local
bindings are committed build inputs; this directory contains their source
specifications.

Domain events belong to their owning context's domain/application code. Private
RabbitMQ message formats and queue metadata belong to that context's service
adaptors, including durable work between aggregates in the same context. They
remain implementation details and are excluded from this published catalogue.

Database bootstrap belongs to [devops/postgres/bootstrap](../devops/postgres/bootstrap/README.md).
Later context migrations live with their owner's persistence adaptor.

Run `pnpm generate:http` for HTTP changes or `pnpm generate:contracts` for all
bindings and metadata. The source inventory in `scripts/contract_sources.py`
maps these owner-facing locations to stable Protobuf import identities. Owner request packages preserve historical field numbers and wire bytes; their
qualified package names now identify the owning context (decision 0009).
`pnpm check:contracts` fails on any generated drift, including newly generated files.
Run `pnpm verify` and `pnpm test:integration` before hand-off.

## Generated files and compatibility fixtures

Generated code is committed so consumers can build without invoking the compiler.
Edit the sources above, regenerate, and review the generated diff. The main output
locations are:

| Consumer | Generated output |
| --- | --- |
| Go API | `services/api/adaptors/messaging/generated/`, `adaptors/openapi/generated/` and `adaptors/http/backend/generated/` |
| Go Storefront | `services/storefront/contracts/{events,requests,realtime}/generated/` and `foundation/transport/openapi/generated/` |
| Python Operations | `services/operations/src/operations/adaptors/generated/`; Protobuf `.py` and `.pyi` files, including private owner command bindings |
| TypeScript Engagement | `services/engagement/src/adaptors/generated/`; owner-private descriptors also live in each context's `adaptors/messaging/generated/` |
| Browser | `services/web/src/adaptors/generated/{http.ts,realtime.json}` |

Go/Python bindings and generated TypeScript source carry generated-file headers.
Descriptor JSON is identified by its `generated/` directory and the explicit
classification in the architecture checker. Generated messages describe transport;
application commands, queries and read models remain independently owned types.

A `fixtures/` directory stores known encoded messages or their expected meaning
for compatibility tests. The public RewardIssued fixture is decoded across
languages. Owner-private fixtures protect previously persisted/queued work.
These are fixed test evidence, separate from database seed data. Preserve historical
bytes when reorganising code; add a fixture for a new format without rewriting an
accepted one. `pnpm check:compatibility` compares against the accepted Git baseline;
`--against <commit>` selects another revision.

`pnpm verify` combines generated drift, historical compatibility and actual caller
compilation/conformance. `pnpm test:integration` proves the same boundaries through
live owner services. Follow the [developer workflow](../docs/developer-workflow.md)
when adding a published operation or private subscription.
