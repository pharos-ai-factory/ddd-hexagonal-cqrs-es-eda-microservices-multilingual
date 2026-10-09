# Communication

Request customer messages and record provider delivery.

A notification owns its delivery status and stable provider identity.

## Find the code

- [Aggregate behaviour](domain)
- [Commands and their handlers](application/commands)
- [Queries and their handlers](application/queries)
- [Event reactions](application/event-handlers)
- [Read models](application/read-models)
- [Application ports](application/ports)
- [HTTP request boundary](adaptors/http)
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
Event reactions enqueue owner commands durably; their handlers execute later.
Queries return application-owned views. Persistence adaptors validate stored
authority and select the fields exposed by those views.

## Work locally

```sh
pnpm context communication
pnpm test:focused engagement --context communication
```

The focused lane includes this context's native tests and behavioural scenarios.
Run `pnpm verify` and `pnpm test:integration` for the complete hand-off checks.

Browser features: [notifications](../../../../web/src/features/notifications).
