# A guided tour of the café

Run `pnpm dev:up` and `pnpm demo` from the repository root. Keep the printed
customer, order, reward and correlation IDs. Open http://127.0.0.1:28000 to use the Next.js operator interface.
Commands and recovery queries go through the separate Go API.

## 1. Start with the invariant

Read `services/storefront/contexts/menu/domain/edition.go` and its test before opening the database
adaptor.

A Drink owns its name and publication revisions. A MenuEdition owns the complete
set of offers it will publish. Each offer freezes a published Drink revision and
price. The edition can reject an empty menu, a duplicate code or mixed currencies
without consulting another aggregate.

Changing a Drink later must not silently rewrite an existing menu. Adding a new
menu requires a new edition identity. This removes a mutable cross-aggregate
dependency from ordering.

Now read `services/storefront/contexts/ordering/domain/order.go`. An OrderLine retains its identity
when its quantity changes. The Order controls that change because quantities
across all lines must total no more than five, and placement freezes the entire
order. Two edits to different lines still compete for one Order version.

In `services/engagement/src/contexts/loyalty/domain`, the account atomically spends three stamps and earns
a grant. A Reward can be issued later and redeemed independently. The grant is
the immediate invariant; the issued voucher is a subsequent workflow result.

## 2. Follow one command inward

Use `POST /api/v1/ordering/orders/{id}/place` as the example:

1. `services/storefront/contexts/ordering/adaptors/http/routes.go` binds a named application handler.
2. `services/storefront/foundation/transport/http/http.go` parses a typed body, command identity and
   expected aggregate version.
3. `services/storefront/contexts/ordering/application/commands.go` loads through a port bound to the
   Order kind, restores the root, calls `Place`, and selects an outgoing contract.
4. `services/storefront/contexts/ordering/domain/order.go` enforces the invariant and records its fact.
5. `services/storefront/foundation/persistence/postgres/command.go` commits the
   state, receipt, outcome, outgoing event and exact browser projection bytes together.

The application never receives `pgx.Tx`, an AMQP delivery or a generated Protobuf
message. Its callback cannot obtain another aggregate repository from the port.
The PostgreSQL tests separately prove that a second aggregate write is rejected.

Repeat PlaceOrder with the exact command key and original expected version. The
stored outcome returns. Use a new key with that stale version and the command is
rejected. The demo performs the first of these checks.

## 3. Follow the event across the process boundary

The Order outbox contains one immutable `ordering.order-placed` publication.
The relay claims its mutable dispatch row, publishes the original bytes with
mandatory routing and waits for broker confirmation.

Python Preparation consumes the event under its own credentials. Its application
entry point is `services/operations/src/operations/contexts/preparation/application.py`. Its handler creates
one PreparationTicket with a stable derived identity, stores its receipt, commits,
and only then acknowledges the delivery.

Preparation completion publishes DrinksReady. Collection opens a Pickup. The
operator supplies its code to collect the order. OrderCollected then reaches
Loyalty. Each arrow is a separate transaction and can be temporarily pending.

## 4. Observe a private domain event

Run:

```sh
pnpm test:integration
```

The lane starts an isolated project with `loyalty.issue-reward` paused. It
collects three orders and observes the account's earned grant while no Reward
exists. It then enables that consumer and observes Reward issuance and
notification delivery.

Both aggregates live inside Loyalty and the same Engagement service. The private
RewardEarned message still uses the full PostgreSQL/RabbitMQ path. This is the
concrete demonstration of eventual consistency inside a bounded context.

The same lane stops Operations while Storefront places its first order. The
command succeeds, and preparation converges after Operations restarts.

## 5. Explore the failure evidence

| Test file | Question it answers |
| --- | --- |
| `services/storefront/foundation/persistence/postgres/command_integration_test.go` | Can state escape without its receipt/outbox? What if the database connection is terminated before commit? |
| `services/storefront/foundation/persistence/postgres/ownership_integration_test.go` | Can one transaction write a second root, a runtime role rewrite history, or an expired relay complete another lease? |
| `services/storefront/foundation/persistence/postgres/receipt_integration_test.go` | Can concurrent conflicting copies of one event change two targets? |
| `services/storefront/foundation/transport/amqp/recovery_integration_test.go` | What happens after commit but before ACK, on unroutable publication, or after retry exhaustion? |
| `services/operations/tests/persistence_integration.py` | Can a Python transaction leak state on encoder failure or complete an expired realtime lease? |
| `services/engagement/src/adaptors/postgres.integration.ts` | Does TypeScript preserve receipts, concurrency and atomic browser intent? |
| `scripts/browser/cafe.spec.ts` | Do independent windows recover without polling and revalidate revoked sessions? |
| `scripts/journey.py` | Does the full business workflow converge, including a lost provider response? |

The provider simulator accepts one notification and deliberately loses the
response. Communication retries the same notification identity and records the
returned acceptance receipt. The accepted-message count remains four.

For an exhausted consumer, repair the cause and replay one original message:

```sh
pnpm events:replay loyalty.issue-reward
```

`replayed=false` means the queue contained no message. Replay preserves the event
identity and bytes. Do not delete receipts to make a replay appear new.

## 6. Inspect without changing authority

Use the HTTP query routes in `contracts/shared/http_api/README.md`. The demo's `Client` in
`scripts/journey.py` loads credentials without embedding them in source.

For an administrative database inspection, this command reads the Loyalty
outbox from the development container:

```sh
docker compose --env-file .local/dev.env -f devops/compose.yaml exec postgres \
  psql -U postgres -d cafe_loyalty \
  -c 'SELECT event_name, aggregate_id, aggregate_version, correlation_id FROM cafe.outbox_events ORDER BY created_at;'
```

This is an operator action. Runtime contexts do not use the administrator account.
Command receipts connect incoming causation to outgoing command causation;
dispatch rows show delivery progress without changing the source event.

## Exercises

- Add a second order line, then try to raise the combined quantity above five.
  Check that neither the snapshot nor its version changes on rejection.
- Publish a renamed Drink revision. Verify that the old edition keeps the old
  name and price while a new edition can use the revision.
- Redeem a Reward twice with different command identities. The second transition
  must be rejected while the account remains independently usable.
- Introduce a cancellation requirement. First decide which invariant is local
  and which steps may remain pending; only then design the messages.
- Compare Go Order placement, Python preparation completion and TypeScript
  account credit. Identify the same command, domain fact, output port and atomic
  transaction roles despite their different language constructs.

Cancellation remains a modelling exercise. The multilingual workflow is implemented.
