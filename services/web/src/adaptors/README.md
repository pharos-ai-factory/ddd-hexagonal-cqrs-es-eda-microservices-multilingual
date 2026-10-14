# Browser boundaries

| Folder                         | Responsibility                                                                 |
| ------------------------------ | ------------------------------------------------------------------------------ |
| `http/operation-client.ts`     | Call generated HTTP operations with typed inputs/replies and session handling. |
| `realtime/projection-codec.ts` | Validate Protobuf projection bytes, owner channels, identities and revisions.  |
| `generated/`                   | HTTP operation types and the browser projection schema.                        |

Business screens live under [features](../features). Command recovery, React
subscriptions and state reconciliation live under [shared](../shared). The
browser receives server-authorised channels and preserves uncertain command
attempts across reloads and authentication loss.

[Repository-wide map](../../../../docs/adaptor-layout.md).
