# Preparation

Accept placed orders and prepare their drinks.

A preparation ticket moves from queued to preparing to ready.

## Find the code

- [Adaptor responsibilities and navigation](adaptors/README.md)

- [Generated plain event contracts](../../contracts/events/generated)
- [Aggregate behaviour](domain)
- [Generated command/query payloads](../../contracts/requests/generated/preparation.py)
- [Commands and their handlers](application/commands)
- [Queries and their handlers](application/queries)
- [Generated command and query identifiers](../../adaptors/generated/request_names.py)
- [Generated event and subscription identifiers](../../adaptors/generated/message_names.py)
- [Event reactions](application/event_handlers)
- [Read models](application/read_models)
- [Application ports](application/ports)
- [Shared bus port contracts](../../foundation/application.py)
- [Ticket write repository](adaptors/persistence/ticket_write_repository.py)
- [Ticket read repository](adaptors/persistence/ticket_read_repository.py)
- [Messaging boundary, codecs and subscriptions](adaptors/messaging)
- [Persistence mapping](adaptors/persistence)
- [Published contracts](../../../../../../contracts/preparation)
- [Executable specifications](../../../../../../specifications/preparation)
- [Shared technical adaptor map](../../adaptors/README.md)
- [Dependency composition](../../apps/composition/preparation.py)
- [Context tests](../../../../tests/contexts/preparation)

## Follow a use case

Start with the named file in `application/commands/` or `application/queries/`.
The handler imports its plain input from the schema-generated request contracts.
Change command/query fields in Protobuf and run `pnpm generate:contracts` first. Follow the injected application port to the
context adaptor, then find its binding in the composition module.
The composition module pairs command names/codecs with concrete handler factories.
Its aggregate registration helper supplies the executor internally.
Aggregate-changing event reactions enqueue owner commands durably; their handlers execute later.
Composition registers each generated query identifier with its input mapper,
concrete handler and reply mapper. Event and subscription identifiers come from
the published and owner-private catalogues.
Queries use named read repository ports and return application-owned views.
Persistence adaptors validate stored
authority and select the fields exposed by those views.

## Work locally

```sh
pnpm context preparation
pnpm test:focused operations --context preparation
```

The focused lane first checks the owning service, then runs this context's native
tests and behavioural scenarios. Use `pnpm check:watch` for checks after edits.
Run `pnpm verify` and `pnpm test:integration` for the complete hand-off checks.

Browser features: [preparation](../../../../../web/src/features/preparation).

Published event schemas own their identities and boundary constraints. Contract
generation supplies plain DTOs and adaptor conversions. This context's messaging
subscription manifest declares its consumers; composition binds their reactions.
Keep semantic evidence checks and domain-to-publication mappings in its adaptors.

The aggregate queues explicit `*DomainEvent` facts. Its save stages private event
bytes with the root. [Publication handlers](application/event_handlers) consume those
facts independently and map them to generated `*IntegrationEvent` DTOs. Their
`IntegrationEventBusPort` commits the subscription receipt, integration bytes and dispatch
together before ACK. A committed root awaiting integration publication is a
permitted pending state; the original fact supplies its revision and terms.

## Explicit browser publication

`CompletePreparation` stages its private fact. The declared
[DrinksReady realtime domain handler](application/event_handlers/realtime_drinks_ready.py)
reads the current ticket view and publishes its actual revision through `RealtimeBusPort`.
A ready ticket awaiting browser publication is a permitted pending state. Accepting
and starting preparation explicitly stage browser snapshots in their command transactions.

The [subscription manifest](adaptors/messaging/subscriptions.json) declares event
producers; composition binds their read repositories and durable realtime buses.
