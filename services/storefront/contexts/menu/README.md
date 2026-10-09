# Menu

Publish drinks and immutable priced menu editions.

A menu edition owns its offers; every offer pins a published drink revision.

## Find the code

- [Aggregate behaviour](domain)
- [Commands and their handlers](application/commands)
- [Queries and their handlers](application/queries)
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
- [Dependency composition](../../apps/storefront/menu.go)
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

The focused lane includes this context's native tests and behavioural scenarios.
Run `pnpm verify` and `pnpm test:integration` for the complete hand-off checks.

Browser features: [menu](../../../web/src/features/menu).
