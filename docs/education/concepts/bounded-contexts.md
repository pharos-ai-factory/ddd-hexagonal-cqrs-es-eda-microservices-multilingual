# Bounded contexts

A bounded context owns a model and the language in which that model is valid.
It gives a business decision an identifiable owner. Its boundary also tells
another context which published contract it may rely on.

## The models in this repository

| Service    | Language   | Contexts                        | Separate authority                                |
| ---------- | ---------- | ------------------------------- | ------------------------------------------------- |
| Storefront | Go         | Menu, Ordering                  | Published menu terms; accepted customer orders    |
| Operations | Python     | Preparation, Collection         | Preparation work; physical handover               |
| Engagement | TypeScript | Loyalty, Customer Communication | Earned benefits; notification intent and delivery |

Each context has its own PostgreSQL database and credentials. Two contexts can
share a service process while retaining separate application/domain packages and
data ownership. The Go API and Next.js frontend expose workflows across those
owners; business policy remains in the domain services.

## Example: three meanings connected by a purchase

Ordering's Order owns chosen lines and placement. Preparation's PreparationTicket
owns work instructions and readiness. Loyalty's account owns collection credits.
An order identity connects those records, but each has its own state and rules.
An integration contract communicates the accepted fact between models.

The [context map](../strategic-ddd.md#context-maps-and-translation) shows the main
dependencies. The [integration event handler](integration-event-handlers.md)
translates the upstream language into the consumer's command or projection.

## Find the owner before changing code

Run `pnpm context ordering`, `pnpm context preparation` or `pnpm context loyalty`.
Read the returned context README and name the invariant, aggregate, command or
query, and permitted pending state. Inter-context reads use a consumer-owned
projection or a published owner port. Importing a neighbouring implementation
would bypass its ownership boundary.

A subdomain describes part of the business problem. A bounded context defines
a model boundary. A microservice defines a deployment unit. The
[strategic DDD guide](../strategic-ddd.md) explains how to choose those boundaries.

## Ownership in composition

These excerpts bind the language service to a named context database. Each
context receives its own resource; handlers reach it through owner ports.

### Go: Ordering

Source: [ordering.go](../../../services/storefront/apps/storefront/di/ordering.go).

```go source="../../../services/storefront/apps/storefront/di/ordering.go"
func(s *support.StorefrontRuntime) execution.Transaction[*d.Order] {
	return pg.NewOrderTransaction(s.Databases["ordering"])
},
```

### Python: Preparation

Source: [preparation.py](../../../services/operations/src/operations/apps/composition/preparation.py).

```python source="../../../services/operations/src/operations/apps/composition/preparation.py"
database = providers.Resource(context_database, "preparation", config.database_url)
```

### TypeScript: Loyalty

Source: [loyalty.ts](../../../services/engagement/src/apps/composition/loyalty.ts).

```ts source="../../../services/engagement/src/apps/composition/loyalty.ts"
database: asFunction(() => new PostgresContextDatabase('loyalty', databaseURL))
  .singleton()
  .disposer((db) => db.pool.end()),
```

## Related concepts

[Domain layer](domain-layer.md) · [Application layer](application-layer.md) · [Integration events](integration-events.md) · [Projections](projections.md)

[Concept catalogue](index.md) · [Learning path](../index.md)
