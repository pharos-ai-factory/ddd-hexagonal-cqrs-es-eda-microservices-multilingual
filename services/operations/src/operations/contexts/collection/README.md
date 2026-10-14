# Collection

Open pickups and record a single successful handover.

A pickup verifies its collection code and permits one handover.

## Find the code

- [Adaptor responsibilities and navigation](adaptors/README.md)

- [Generated plain event contracts](../../contracts/events/generated)
- [Aggregate behaviour](domain)
- [Commands and their handlers](application/commands)
- [Queries and their handlers](application/queries)
- [Event reactions](application/event_handlers)
- [Read models](application/read_models)
- [Application ports](application/ports)
- [Pickup write repository](adaptors/persistence/pickup_write_repository.py)
- [Pickup read repository](adaptors/persistence/pickup_read_repository.py)
- [Messaging boundary, codecs and subscriptions](adaptors/messaging)
- [Persistence mapping](adaptors/persistence)
- [Published contracts](../../../../../../contracts/collection)
- [Executable specifications](../../../../../../specifications/collection)
- [Shared technical adaptor map](../../adaptors/README.md)
- [Dependency composition](../../apps/composition/collection.py)
- [Context tests](../../../../tests/contexts/collection)

## Follow a use case

Start with the named file in `application/commands/` or `application/queries/`.
The DTO and handler live together. Follow the injected application port to the
context adaptor, then find its binding in the composition module.
Aggregate-changing event reactions enqueue owner commands durably; their handlers execute later.
Queries use named read repository ports and return application-owned views.
Persistence adaptors validate stored
authority and select the fields exposed by those views.

## Work locally

```sh
pnpm context collection
pnpm test:focused operations --context collection
```

The focused lane first checks the owning service, then runs this context's native
tests and behavioural scenarios. Use `pnpm check:watch` for checks after edits.
Run `pnpm verify` and `pnpm test:integration` for the complete hand-off checks.

Browser features: [collection](../../../../../web/src/features/collection).

Published event schemas own their identities and boundary constraints. Contract
generation supplies plain DTOs and adaptor conversions. This context's messaging
subscription manifest declares its consumers; composition binds their reactions.
Keep semantic evidence checks and domain-to-publication mappings in its adaptors.

The aggregate queues explicit `*DomainEvent` facts. Its save stages private event
bytes with the root. [Publication handlers](application/event_handlers) consume those
facts independently and map them to generated `*IntegrationEvent` DTOs. Their
publication port commits the subscription receipt, integration bytes and dispatch
together before ACK. A committed root awaiting integration publication is a
permitted pending state; the original fact supplies its revision and terms.
