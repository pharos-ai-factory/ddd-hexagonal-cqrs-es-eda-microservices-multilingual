# Preparation

Accept placed orders and prepare their drinks.

A preparation ticket moves from queued to preparing to ready.

## Find the code

- [Aggregate behaviour](domain)
- [Commands and their handlers](application/commands)
- [Queries and their handlers](application/queries)
- [Event reactions](application/event_handlers)
- [Read models](application/read_models)
- [Application ports](application/ports)
- [Ticket write repository](adaptors/persistence/ticket_write_repository.py)
- [Ticket read repository](adaptors/persistence/ticket_read_repository.py)
- [Messaging boundary, codecs and subscriptions](adaptors/messaging)
- [Persistence mapping](adaptors/persistence)
- [Published contracts](../../../../../../contracts/preparation)
- [Executable specifications](../../../../../../specifications/preparation)
- [Dependency composition](../../apps/composition/preparation.py)
- [Context tests](../../../../tests/contexts/preparation)

## Follow a use case

Start with the named file in `application/commands/` or `application/queries/`.
The DTO and handler live together. Follow the injected application port to the
context adaptor, then find its binding in the composition module.
Event reactions enqueue owner commands durably; their handlers execute later.
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
