# Menu

Publish drinks and immutable priced menu editions.

A menu edition owns its offers; every offer pins a published drink revision.

## Find the code

- [Adaptor responsibilities and navigation](adaptors/README.md)

- [Aggregate behaviour](domain)
- [Generated command/query payloads](../../contracts/requests/model/menu/payloads_generated.go)
- [Commands and their handlers](application/commands)
- [Queries and their handlers](application/queries)
- [Generated command and query identifiers](../../contracts/requests/generated/cafe/requests/v1/names.go)
- [Generated event and subscription identifiers](../../contracts/events/names/names_generated.go)
- [Domain publication handlers](application/eventhandlers)
- [Projection handlers](application/projections)
- [Read models](application/readmodels)
- [Application ports](application/ports)
- [Shared bus port contracts](../../foundation/application/buses.go)
- [Drink write repository](adaptors/persistence/postgres/drink_write_repository.go)
- [Drink read repository](adaptors/persistence/postgres/drink_read_repository.go)
- [Edition read repository](adaptors/persistence/postgres/edition_read_repository.go)
- [Messaging boundary, codecs and subscriptions](adaptors/messaging)
- [Persistence mapping](adaptors/persistence/postgres)
- [Published contracts](../../../../contracts/menu)
- [Executable specifications](../../../../specifications/menu)
- [Dependency composition](../../apps/storefront/di/menu.go)
- [Command and query registrations](../../apps/storefront/di/menu_requests.go)
- [Context tests](.)

## Follow a use case

Start with the named file in `application/commands/` or `application/queries/`.
The handler imports its plain input from the schema-generated request contracts.
Change command/query fields in Protobuf and run `pnpm generate:contracts` first. Follow the injected application port to the
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
`IntegrationEventBusPort` commits the subscription receipt, integration bytes and dispatch
together before ACK. A committed root awaiting integration publication is a
permitted pending state; the original fact supplies its revision and terms.

## Explicit browser publication

Menu command handlers explicitly publish drink and edition snapshots through
`RealtimeBusPort` after saving. The executor commits those selected bytes with the
root. Menu integration publication remains an independent domain subscription.

The [subscription manifest](adaptors/messaging/subscriptions.json) declares event
producers; composition binds their read repositories and durable realtime buses.
