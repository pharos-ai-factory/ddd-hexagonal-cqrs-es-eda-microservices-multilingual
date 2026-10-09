# Collection

Open pickups and record a single successful handover.

A pickup verifies its collection code and permits one handover.

## Find the code

- [Aggregate behaviour](domain)
- [Commands and their handlers](application/commands)
- [Queries and their handlers](application/queries)
- [Event reactions](application/event_handlers)
- [Read models](application/read_models)
- [Application ports](application/ports)
- [HTTP request boundary](adaptors/http)
- [Messaging boundary, codecs and subscriptions](adaptors/messaging)
- [Persistence mapping](adaptors/persistence)
- [Published contracts](../../../../../../contracts/collection)
- [Executable specifications](../../../../../../specifications/collection)
- [Dependency composition](../../apps/composition/collection.py)
- [Context tests](../../../../tests/contexts/collection)

## Follow a use case

Start with the named file in `application/commands/` or `application/queries/`.
The DTO and handler live together. Follow the injected application port to the
context adaptor, then find its binding in the composition module.
Event reactions enqueue owner commands durably; their handlers execute later.
Queries return application-owned views. Persistence adaptors validate stored
authority and select the fields exposed by those views.

## Work locally

```sh
pnpm context collection
pnpm test:focused operations --context collection
```

The focused lane includes this context's native tests and behavioural scenarios.
Run `pnpm verify` and `pnpm test:integration` for the complete hand-off checks.

Browser features: [collection](../../../../../web/src/features/collection).
