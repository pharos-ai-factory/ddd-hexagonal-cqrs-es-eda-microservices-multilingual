# Ordering

Build and place orders from published menus.

An order owns its lines and enforces the five-drink limit.

## Find the code

- [Adaptor responsibilities and navigation](adaptors/README.md)

- [Aggregate behaviour](domain)
- [Commands and their handlers](application/commands)
- [Queries and their handlers](application/queries)
- [Generated command and query identifiers](../../contracts/requests/generated/cafe/requests/v1/names.go)
- [Generated event and subscription identifiers](../../contracts/events/names/names_generated.go)
- [Domain publication handlers](application/eventhandlers)
- [Projection handlers](application/projections)
- [Read models](application/readmodels)
- [Application ports](application/ports)
- [Order write repository](adaptors/persistence/postgres/order_write_repository.go)
- [Order read repository](adaptors/persistence/postgres/order_read_repository.go)
- [Messaging boundary, codecs and subscriptions](adaptors/messaging)
- [Persistence mapping](adaptors/persistence/postgres)
- [Published contracts](../../../../contracts/ordering)
- [Executable specifications](../../../../specifications/ordering)
- [Dependency composition](../../apps/storefront/di/ordering.go)
- [Command and query registrations](../../apps/storefront/di/ordering_requests.go)
- [Context tests](.)

## Follow a use case

Start with the named file in `application/commands/` or `application/queries/`.
The DTO and handler live together. Follow the injected application port to the
context adaptor, then find its binding in the composition module.
Command handlers load, mutate and save through named write repositories.
Composition pairs each command with its plain handler factory through the shared
registration helper, which supplies each invocation's repository and transaction. Owner adaptors map aggregate facts into outgoing messages.
Projection handlers update consumer-owned read data without changing aggregates.
Composition registers each generated query identifier with its input mapper,
concrete handler and reply mapper. Event and subscription identifiers come from
the published and owner-private catalogues.
Queries use named read repository ports and return application-owned views.
Persistence adaptors validate stored
authority and select the fields exposed by those views.

## Work locally

```sh
pnpm context ordering
pnpm test:focused storefront --context ordering
```

The focused lane first checks the owning service, then runs this context's native
tests and behavioural scenarios. Use `pnpm check:watch` for checks after edits.
Run `pnpm verify` and `pnpm test:integration` for the complete hand-off checks.

Browser features: [ordering](../../../web/src/features/ordering).

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
