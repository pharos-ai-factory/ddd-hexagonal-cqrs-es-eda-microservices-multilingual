# Communication

Request customer messages and record provider delivery.

A notification owns its delivery status and stable provider identity.

## Find the code

- [Adaptor responsibilities and navigation](adaptors/README.md)

- [Generated plain event contracts](../../contracts/events/generated)
- [Aggregate behaviour](domain)
- [Commands and their handlers](application/commands)
- [Queries and their handlers](application/queries)
- [Generated command and query identifiers](../../adaptors/generated/request-names.ts)
- [Generated event and subscription identifiers](../../adaptors/generated/message-names.ts)
- [Event reactions](application/event-handlers)
- [Read models](application/read-models)
- [Application ports](application/ports)
- [Notification write repository](adaptors/persistence/notification-write-repository.ts)
- [Notification read repository](adaptors/persistence/notification-read-repository.ts)
- [Messaging boundary, codecs and subscriptions](adaptors/messaging)
- [Persistence mapping](adaptors/persistence)
- [Published contracts](../../../../../contracts/communication)
- [Executable specifications](../../../../../specifications/communication)
- [Dependency composition](../../apps/composition/communication.ts)
- [Context tests](.)

## Follow a use case

Start with the named file in `application/commands/` or `application/queries/`.
The DTO and handler live together. Follow the injected application port to the
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
pnpm context communication
pnpm test:focused engagement --context communication
```

The focused lane first checks the owning service, then runs this context's native
tests and behavioural scenarios. Use `pnpm check:watch` for checks after edits.
Run `pnpm verify` and `pnpm test:integration` for the complete hand-off checks.

Browser features: [notifications](../../../../web/src/features/notifications).

Published event schemas own their identities and boundary constraints. Contract
generation supplies plain DTOs and adaptor conversions. This context's messaging
subscription manifest declares its consumers; composition binds their reactions.
Keep semantic evidence checks and domain-to-publication mappings in its adaptors.

`NotificationRequestedDomainEvent` remains private to this context. Its durable
subscription requests the delivery command. Communication currently publishes no
integration event and needs no integration publication subscriber.
