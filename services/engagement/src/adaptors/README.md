# Engagement boundary mechanisms

Loyalty earns grants, issues rewards and records redemption. Communication
requests notifications and records provider delivery. Their [Loyalty](../contexts/loyalty/README.md)
and [Communication](../contexts/communication/README.md) adaptors own business mappings.

| Folder or file         | Responsibility                                                                                                                     |
| ---------------------- | ---------------------------------------------------------------------------------------------------------------------------------- |
| `persistence/`         | PostgreSQL pools, snapshot repositories, aggregate transactions, receipts, command/integration/reply outboxes and dispatch leases. |
| `messaging/`           | Event/private-command/request codecs, confirmed RabbitMQ publication, consumers and outbox/reply relays.                           |
| `realtime/`            | Encode browser projections and publish committed bytes to Centrifugo.                                                              |
| `http/`                | Health and authenticated diagnostics.                                                                                              |
| `generated/`           | Contract catalogues, owner codec registration and schema metadata.                                                                 |
| `command-execution.ts` | Bind a plain named handler to its invocation transaction.                                                                          |
| `diagnostics.ts`       | Assemble process worker and durable backlog evidence.                                                                              |

For a collected order, Loyalty's subscription enqueues CreditCollection durably.
Its command handler saves the account and private RewardEarned fact. A second
subscription enqueues IssueReward; the reward's private fact is then delivered to
its integration publication handler. Each pending hand-off retains its identity
and exact bytes. Notification provider calls live in Communication's `delivery/`.

[Repository-wide map](../../../../docs/adaptor-layout.md).
