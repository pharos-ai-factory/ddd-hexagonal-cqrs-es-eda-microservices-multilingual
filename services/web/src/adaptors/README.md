# Browser boundaries

| Folder                         | Responsibility                                                                  |
| ------------------------------ | ------------------------------------------------------------------------------- |
| `http/operation-client.ts`     | Retain raw durable-command responses while using generated operation contracts. |
| `realtime/message-codec.ts`    | Validate typed Protobuf messages, delivery targets, identities and revisions.   |
| `generated/http/<tag>/`        | Orval fetch clients, TanStack Query hooks, MSW handlers and Faker factories.    |
| `generated/http/operations.ts` | Typed operation manifest used by the durable-command adaptor.                   |
| `generated/realtime.json`      | Browser realtime Protobuf schema.                                               |

Business screens live under [features](../features). Command recovery, React
subscriptions and state reconciliation live under [shared](../shared). The
browser receives server-authorised channels and preserves uncertain command
attempts across reloads and authentication loss.

See the [browser OpenAPI client guide](../../../../docs/openapi-client.md) for
generation, imports, cache ownership and test doubles.

[Repository-wide map](../../../../docs/adaptor-layout.md).
