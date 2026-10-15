# Loyalty

Earn grants from collections, issue rewards and record redemption.

A loyalty account earns each third-collection grant; a reward independently controls validity and redemption.

## Find the code

- [Adaptor responsibilities and navigation](adaptors/README.md)

- [Generated plain event contracts](../../contracts/events/generated)
- [Aggregate behaviour](domain)
- [Generated command/query payloads](../../contracts/requests/generated/loyalty.ts)
- [Generated responses and resource DTOs](../../contracts/requests/generated/loyalty-responses.ts)
- [Commands and their handlers](application/commands)
- [Queries and their handlers](application/queries)
- [Generated command and query identifiers](../../adaptors/generated/request-names.ts)
- [Generated event and subscription identifiers](../../adaptors/generated/message-names.ts)
- [Event reactions](application/event-handlers)
- [Read models](application/read-models)
- [Application ports](application/ports)
- [Shared bus port contracts](../../foundation/application.ts)
- [Account write repository](adaptors/persistence/account-write-repository.ts)
- [Account read repository](adaptors/persistence/account-read-repository.ts)
- [Reward read repository](adaptors/persistence/reward-read-repository.ts)
- [Messaging boundary, codecs and subscriptions](adaptors/messaging)
- [Persistence mapping](adaptors/persistence)
- [Published contracts](../../../../../contracts/loyalty)
- [Executable specifications](../../../../../specifications/loyalty)
- [Dependency composition](../../apps/composition/loyalty.ts)
- [Context tests](.)

## Follow a use case

Start with the named file in `application/commands/` or `application/queries/`.
The handler imports its plain input and response from the schema-generated request contracts.
Change command/query input and response fields in Protobuf and run `pnpm generate:contracts` first. Follow the injected application port to the
context adaptor, then find its binding in the composition module.
The composition module pairs command names/codecs with concrete handler factories.
Its aggregate registration helper supplies the executor internally.
Aggregate-changing event reactions enqueue owner commands durably; their handlers execute later.
Composition uses a generated query binding with its input mapper and concrete handler.
The binding fixes the identifier, result contract and generated reply encoder. Event and subscription identifiers come from
the published and owner-private catalogues.
Queries use named read repository ports and return generated plain responses.
Persistence adaptors validate stored
authority and select the fields exposed by those views.

## Work locally

```sh
pnpm context loyalty
pnpm test:focused engagement --context loyalty
```

The focused lane first checks the owning service, then runs this context's native
tests and behavioural scenarios. Use `pnpm check:watch` for checks after edits.
Run `pnpm verify` and `pnpm test:integration` for the complete hand-off checks.

Browser features: [rewards](../../../../web/src/features/rewards).

Published event schemas own their identities and boundary constraints. Contract
generation supplies plain DTOs and adaptor conversions. This context's messaging
subscription manifest declares its consumers; composition binds their reactions.
Keep semantic evidence checks and domain-to-publication mappings in its adaptors.

The aggregate queues explicit `*DomainEvent` facts. Its save stages private event
bytes with the root. [Publication handlers](application/event-handlers) consume those
facts independently and map them to generated `*IntegrationEvent` DTOs. Their
`IntegrationEventBusPort` commits the subscription receipt, integration bytes and dispatch
together before ACK. A committed root awaiting integration publication is a
permitted pending state; the original fact supplies its revision and terms.

## Explicit browser publication

`IssueReward` stages a private fact, whose publication handler produces the
integration event. The separately declared
[RewardIssued realtime integration handler](application/event-handlers/realtime-reward-issued.ts)
reads the current reward view and publishes its actual revision through `RealtimeBusPort`.
An issued reward awaiting integration delivery and browser publication is a permitted
pending state. Account credits and redemption explicitly stage command snapshots;
redemption can also publish a recorded business rejection.

The [subscription manifest](adaptors/messaging/subscriptions.json) declares event
producers; composition binds their read repositories and durable realtime buses.
