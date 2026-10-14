# Menu

Publish drinks and immutable priced menu editions.

A menu edition owns its offers; every offer pins a published drink revision.

## Find the code

- [Aggregate behaviour](domain)
- [Commands and their handlers](application/commands)
- [Queries and their handlers](application/queries)
- [Domain publication handlers](application/eventhandlers)
- [Projection handlers](application/projections)
- [Read models](application/readmodels)
- [Application ports](application/ports)
- [Drink write repository](adaptors/postgres/drink_write_repository.go)
- [Drink read repository](adaptors/postgres/drink_read_repository.go)
- [Edition read repository](adaptors/postgres/edition_read_repository.go)
- [Messaging boundary, codecs and subscriptions](adaptors/messaging)
- [Persistence mapping](adaptors/postgres)
- [Published contracts](../../../../contracts/menu)
- [Executable specifications](../../../../specifications/menu)
- [Dependency composition](../../apps/storefront/di/menu.go)
- [Context tests](.)

## Follow a use case

Start with the named file in `application/commands/` or `application/queries/`.
The DTO and handler live together. Follow the injected application port to the
context adaptor, then find its binding in the composition module.
Command handlers load, mutate and save through named write repositories.
Composition registers a central executor that supplies each invocation's repository
and manages its transaction. Owner adaptors map aggregate facts into outgoing messages.
Projection handlers update consumer-owned read data without changing aggregates.
Queries use named read repository ports and return application-owned views.
Persistence adaptors validate stored
authority and select the fields exposed by those views.

## Work locally

```sh
pnpm context menu
pnpm test:focused storefront --context menu
```

The focused lane first checks the owning service, then runs this context's native
tests and behavioural scenarios. Use `pnpm check:watch` for checks after edits.
Run `pnpm verify` and `pnpm test:integration` for the complete hand-off checks.

Browser features: [menu](../../../web/src/features/menu).

Published event schemas own their identities and boundary constraints. Contract
generation supplies plain DTOs and adaptor conversions. This context's messaging
subscription manifest declares its consumers; composition binds their reactions.
Keep semantic evidence checks and domain-to-publication mappings in its adaptors.

The aggregate queues explicit `*DomainEvent` facts. Its save stages private event
bytes with the root. [Publication handlers](application/eventhandlers) consume those
facts independently and map them to generated `*IntegrationEvent` DTOs. Their
publication port commits the subscription receipt, integration bytes and dispatch
together before ACK. A committed root awaiting integration publication is a
permitted pending state; the original fact supplies its revision and terms.
