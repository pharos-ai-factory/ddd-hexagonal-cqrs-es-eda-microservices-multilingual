# Domain model and tactical rules

## Model before persistence

An entity has continuity of identity. A value object compares by value.
An aggregate root controls a complete immediate consistency boundary and owns
all mutation of its children. A database row does not define an aggregate.

| Context | Aggregate | Invariants and independent lifecycle |
| --- | --- | --- |
| Menu | Drink | Published definitions have stable identity and immutable revisions |
| Menu | MenuEdition | Unique offer codes, one currency, non-empty publication and an immutable complete published snapshot |
| Ordering | Order | Stable child OrderLine identities; one edition; quantities total 1–5 at placement; placed contents are immutable |
| Preparation | PreparationTicket | Queued → Preparing → Ready |
| Collection | Pickup | Ready before handover, correct collection code, at most one collection |
| Loyalty | LoyaltyAccount | Credit each collection once; three credits atomically earn one stable grant |
| Loyalty | Reward | One voucher per grant; at most one redemption; fixed validity |
| Customer Communication | Notification | One intent per business occasion; delivery outcome follows an idempotent provider request |

MenuOffer is an immutable value inside an edition. It references a published
Drink revision and freezes that offer's terms. Drink has its own lifecycle.
OrderLine is a child entity: editing its quantity retains the line's
identity but advances the Order version.

RewardGrant is an immutable earned fact within the account's transition.
Reward is independently issued and redeemed. An earned grant awaiting issuance
is permitted; a silently lost grant is not.

## Commands, queries and events

Commands request one aggregate transition. A handler records typed expected
rejections as stable outcomes. Unavailable infrastructure is retryable and
does not become a business rejection.

Domain and application errors may be named classes or structs in their owning
context. Service-local foundations define the shared expected-rejection shape;
command ports serialise it as the existing code and message, not as a runtime
class. Corrupt restored state and infrastructure failures remain retryable. See
decision 0005.

Clients supply an idempotency key and expected version. Repeating the same
command returns its recorded outcome. Reusing the identity with different
material input is a conflict. Duplicate/no-op transitions emit no new events.

Queries have named input/handler pairs, application-owned read models and named
read repository ports. Persistence adaptors validate stored authority and select view
fields. A query preserves root revisions, missing resources and explicit pagination
without acquiring mutation capability. Domain methods have no network,
database, logging, wall-clock or random-number calls; time and identity arrive
as explicit input.

A successful domain transition records plain past-tense facts. The application
selects private domain delivery and public integration delivery explicitly.
All required aggregate-to-aggregate reactions are durable and asynchronous,
including reactions inside the same process.

Examples:

```text
PlaceOrder → OrderPlaced → AcceptOrder → PreparationTicket
CompletePreparation → DrinksReady → OpenPickup → Pickup
CollectOrder → OrderCollected → CreditCollection → LoyaltyAccount
RewardEarned (private domain delivery) → IssueReward → Reward
RewardIssued (published integration delivery) → RequestNotification → Notification
```

Each event-to-command arrow is a durable hand-off. The reaction commits acceptance
and command bytes, then acknowledges the event. The receiving command consumer
owns the subsequent aggregate transaction and retries. A projection handler can
instead update consumer-owned read data without mutating a root.

## Concurrent decisions

Two edits based on one aggregate version cannot both commit. Changing separate
OrderLines still competes for the Order version. Publishing a menu competes with
editing that edition. Redemption competes with another redemption of the same
Reward, but does not compete with earning another stamp.

Each consumer has an independent receipt. Technical deduplication by message ID
is supplemented by the business key: collection identity for credits and grant
identity for rewards.

## Language translations

Use idiomatic constructs in each implementation language. Preserve encapsulation,
invariants, typed outcomes, transaction boundaries and externally observable
behaviour. Generated wire types stay outside the domain and application rings.
The language-neutral Gherkin catalogue, HTTP scenarios and Protobuf fixtures
form the translation contract. Write business outcomes in
`specifications/<context>` and bind them to that context's real handlers.
Use real infrastructure when the example claims durable retries, duplicate
business-key protection or concurrency; a decision probe cannot prove those.

Python commands, outcomes, published DTOs, read views and restored snapshots have
explicit shapes, including literal lifecycle states. Command ports retain the
owning snapshot type; read repository ports expose application views. Domain facts use
immutable context-owned dataclasses.
Static types do not validate external JSON or stored state. Boundary decoders
must still reject malformed input and classify corrupt authority as a retryable
state failure, preserving the distinction from a business rejection.

## Command entry points

Every aggregate business transition, including creation, begins in an owner
`CommandHandler.Execute`/`execute` method with a named command DTO. The aggregate
owns invariants. Event handlers use a durable command port for aggregate reactions;
private events between roots follow the same rule. Projection updates and
rehydration remain separate operations. This is the reference's explicit convention;
DDD also permits transactionally safe event-handling use cases. Decision 0011
records why this repository accepts the extra durable hand-off.

Each command or query DTO lives beside its handler in one business-named file
under `application/commands/` or `application/queries/`. Each event reaction has
its own module. Application handlers never call another command handler, and a
command execution changes one aggregate through its named write repository. Plain constructor
ports keep these use cases independent of DI, transport and persistence. See
[decisions 0012–0014](docs/decisions/README.md).

## Persistence and command infrastructure

Command handlers load and save through `<Aggregate>WriteRepository`. The root
records domain facts while enforcing its rules. A central command executor and
owner persistence/message adaptors commit the aggregate and selected outgoing
facts together with technical receipt/outcome records. This local transaction
belongs to one aggregate and one context database. Feature code carries no
transaction callbacks or delivery bookkeeping. See [decision 0015](docs/decisions/0015-explicit-persistence-role-names.md).
