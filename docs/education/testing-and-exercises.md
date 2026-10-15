# Testing the model and its boundaries

Different tests answer different questions. A domain test can prove a quantity
rule using fixed inputs. It cannot prove that PostgreSQL commits an outbox with
the aggregate or that a consumer acknowledges after committing.

## Choose evidence for the claim

| Claim                                                        | Appropriate evidence in this repository                       |
| ------------------------------------------------------------ | ------------------------------------------------------------- |
| An order rejects an invalid placement quantity               | Aggregate tests covering accepted and rejected transitions    |
| A handler invokes the owner's behaviour and saves its result | Application tests with plain ports                            |
| A transport decoder rejects invalid wire input               | Boundary/codec tests                                          |
| An invariant holds under concurrent writes                   | Real database concurrency tests                               |
| State and outgoing intent commit together                    | Transaction rollback and recovery tests                       |
| Duplicate delivery produces one business effect              | Broker/database tests with durable receipts and business keys |
| A browser recovers uncertain commands and stale updates      | Real API/realtime/browser tests                               |

**Behaviour-driven development** uses concrete examples to establish a shared
understanding of desired behaviour. Gherkin is one notation for those examples;
writing a feature file alone does not prove the implementation satisfies it.
The [specification catalogue](../../specifications/README.md) binds examples to
native or live test lanes and rejects missing execution evidence.

Use fixed identities and supplied timestamps in domain tests. Supply boundary
doubles where they answer the question, and use real infrastructure for claims
about its transaction and delivery semantics. The complete commands and
environment requirements are in [TESTING.md](../../TESTING.md).

## Exercise 1: locate an invariant

Run `pnpm context ordering`, then read the
[Order](../../services/storefront/contexts/ordering/domain/order.go),
[PlaceOrder handler](../../services/storefront/contexts/ordering/application/commands/place_order.go)
and [Ordering scenarios](../../specifications/ordering/ordering.feature).

Describe placing an empty draft, placing a draft with five drinks and adding a
sixth drink to a draft. Identify which operation rejects each invalid transition
and which root owns the rule.

**Check your reasoning:** the combined quantity and placement lifecycle belong
to Order. OrderLine edits are routed through the root. HTTP validation can reject
malformed input, but aggregate behaviour still protects the business invariant.

## Exercise 2: follow the dependency direction

Start at PlaceOrder's named write repository port. Find its PostgreSQL adaptor
and its composition binding through the
[Ordering navigation](../../services/storefront/contexts/ordering/README.md).

Draw two arrow sets: runtime calls and source imports. Explain how the handler
can call PostgreSQL without importing the PostgreSQL driver.

**Check your reasoning:** the handler depends on its port. The concrete adaptor
implements that port, and composition supplies the instance. A replacement
adaptor must preserve the port's behavioural contract as well as its method names.

## Exercise 3: distinguish two retries

Read the [projection recovery scenarios](../../specifications/ordering/projection-recovery.feature).
An order attempt receives `menu_pending`. The menu projection later arrives.
Compare an identical command retry with a new business attempt.

**Check your reasoning:** the identical retry recovers the original recorded
outcome. The new attempt receives a new command identity and can evaluate the
now-available evidence. Reusing the old key with changed input is a conflict.

## Exercise 4: separate an entitlement from its fulfilment

Read [LoyaltyAccount](../../services/engagement/src/contexts/loyalty/domain/loyalty-account.ts)
and [Reward](../../services/engagement/src/contexts/loyalty/domain/reward.ts).
Explain what must survive if the Reward consumer stops immediately after the
third eligible collection is credited.

**Check your reasoning:** the account's earned grant and durable private event
intent must survive. Reward issuance can remain pending and later recover. The
credit and issuance have different transaction boundaries and duplicate guards.

## Run the focused evidence

After installing the [documented prerequisites](../../README.md#run-it), run:

```sh
pnpm test:focused storefront --context ordering
pnpm test:focused operations --context preparation
pnpm test:focused engagement --context loyalty
```

These lanes exercise native behaviour and the registered focused examples.
`pnpm verify` checks the complete deterministic suite. `pnpm test:integration`
provisions a disposable infrastructure stack and proves the live boundaries.
Neither command deploys a production system.

## Observe the whole workflow

For a local demonstration, follow the [walkthrough](../walkthrough.md) using
`pnpm dev:up` and `pnpm demo`. Keep the printed identifiers and inspect the related
workflow evidence. The demo creates data in your development stack. The integration
runner uses its separate disposable project.

Before claiming completion, describe the business outcome, the immediate
invariant, the permitted pending states and the test evidence for each boundary.
Use [the glossary](glossary.md) to revisit any unfamiliar term.
