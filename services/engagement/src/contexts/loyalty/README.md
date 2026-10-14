# Loyalty

Earn grants from collections, issue rewards and record redemption.

A loyalty account earns each third-collection grant; a reward independently controls validity and redemption.

## Find the code

- [Generated plain event contracts](../../contracts/events/generated)
- [Aggregate behaviour](domain)
- [Commands and their handlers](application/commands)
- [Queries and their handlers](application/queries)
- [Event reactions](application/event-handlers)
- [Read models](application/read-models)
- [Application ports](application/ports)
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
The DTO and handler live together. Follow the injected application port to the
context adaptor, then find its binding in the composition module.
Event reactions enqueue owner commands durably; their handlers execute later.
Queries use named read repository ports and return application-owned views.
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
