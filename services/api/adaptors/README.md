# API boundaries

The API authenticates browser requests, translates the published HTTP contract
into owner messaging requests, and authorises realtime subscriptions.

| Folder                                                 | Responsibility                                                                                                                                                                   |
| ------------------------------------------------------ | -------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| `http/backend/`                                        | Business HTTP routing and generated owner request translation.                                                                                                                   |
| `http/session/`, `http/realtime/`, `http/operational/` | Session endpoints, subscription authorisation and diagnostics.                                                                                                                   |
| `messaging/`                                           | `request_codec.go` validates envelopes; `rabbitmq_client.go` correlates owner replies.                                                                                           |
| `sessions/`                                            | `valkey_session_store.go` stores authentication state; `revocation_worker.go` drains durable revocations after the grant fence; `centrifugo_disconnect.go` performs gateway I/O. |
| `realtime/`                                            | Projection envelopes and the Centrifugo gateway boundary.                                                                                                                        |
| `openapi/`                                             | Validate the generated HTTP contract.                                                                                                                                            |
| `config/`                                              | Load credentials for composition.                                                                                                                                                |

Owner business policy remains behind published contracts. Session revocation
removes credentials and records disconnect work durably before worker delivery.
The worker preserves pending evidence until the gateway accepts the disconnect.

[Repository-wide map](../../../docs/adaptor-layout.md).
