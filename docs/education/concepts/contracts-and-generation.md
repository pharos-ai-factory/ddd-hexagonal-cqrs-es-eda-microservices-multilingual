# Contracts and generation

A published contract defines the boundary consumers may depend on. Owner schemas
and subscription manifests generate language-specific DTOs, identifiers, codecs
and topology. Handwritten domain and application policy remains with its owner.

## Creation recipe

1. Choose the boundary: published command/query/integration Protobuf, separate browser Protobuf, owner-private queued messages or OpenAPI HTTP.
2. Edit that owner source; preserve existing wire field numbers and historical fixtures. Add envelope alternatives, annotations and constraints where the boundary requires them.
3. For delivered reactions, update the consumer-owned `subscriptions.json`. A `realtime` entry is an explicit browser reaction; `publication` maps a private fact to integration; `command` binds a durable owner command.
4. Run `pnpm generate:contracts` for Protobuf/catalogue changes or `pnpm generate:http` for OpenAPI alone. Commit generated files with the source changes.
5. Import generated command/query inputs and responses into handlers. Use generated typed query bindings and mechanical encoders; implement owner semantic validation, business mappings and handlers.
6. Update API operation mappings and the generated operation client for HTTP exposure. Bind every declared consumer and resolve DI before starting workers.
7. Run generated-drift and compatibility checks, focused tests and full verification. Apply new context migrations and generated broker permissions to the development composition before exercising the new route.

Follow the [end-to-end recipe](../end-to-end-recipe.md) to connect this step to the complete workflow.

## Sources and outputs

| Boundary                    | Authoritative source                                       | Generated roles                                                             |
| --------------------------- | ---------------------------------------------------------- | --------------------------------------------------------------------------- |
| Public commands and queries | Owner messaging Protobuf envelopes                         | Plain inputs/responses, wire bindings, parsers, encoders and typed bindings |
| Integration events          | Owner event Protobuf annotations                           | Plain DTOs, validation, codecs and publication factories                    |
| Private domain delivery     | Owner `private_messages.proto` and `private_messages.json` | Private codecs and routing identifiers                                      |
| Private commands            | Owner `internal_commands.proto`                            | Plain inputs/outcomes, payload mappings, wire schemas and identifiers       |
| Subscriptions               | Consumer-owned `subscriptions.json`                        | Subscription identifiers and topology                                       |
| Realtime                    | Owner snapshot schemas and shared envelope                 | Browser projection codecs and types                                         |
| HTTP                        | Owner OpenAPI sources                                      | API/browser operation contracts                                             |

## A manifest in each language service

These JSON declarations are consumed by the build tooling. Their entries describe
projection, command and publication subscriptions. The runtime composition binds
them to actual handlers and validates the complete set.

### Go service: Ordering

Source: [subscriptions.json](../../../services/storefront/contexts/ordering/adaptors/messaging/subscriptions.json).

```json source="../../../services/storefront/contexts/ordering/adaptors/messaging/subscriptions.json"
[
  {
    "consumer": "ordering.menu-directory",
    "event": "menu.edition-published"
  },
  {
    "consumer": "ordering.publish-order-placed",
    "event": "ordering.order-placed.domain",
    "publication": "ordering.order-placed"
  },
  {
    "consumer": "ordering.realtime-order-placed",
    "event": "ordering.order-placed.domain",
    "realtime": "order"
  },
  {
    "consumer": "ordering.record-collection",
    "event": "collection.order-collected",
    "command": "recordCollection"
  }
]
```

### Python service: Preparation

Source: [subscriptions.json](../../../services/operations/src/operations/contexts/preparation/adaptors/messaging/subscriptions.json).

```json source="../../../services/operations/src/operations/contexts/preparation/adaptors/messaging/subscriptions.json"
[
  {
    "consumer": "preparation.accept-order",
    "event": "ordering.order-placed",
    "command": "acceptOrder"
  },
  {
    "consumer": "preparation.publish-drinks-ready",
    "event": "preparation.drinks-ready.domain",
    "publication": "preparation.drinks-ready"
  },
  {
    "consumer": "preparation.realtime-drinks-ready",
    "event": "preparation.drinks-ready.domain",
    "realtime": "ticket"
  }
]
```

### TypeScript service: Loyalty

Source: [subscriptions.json](../../../services/engagement/src/contexts/loyalty/adaptors/messaging/subscriptions.json).

```json source="../../../services/engagement/src/contexts/loyalty/adaptors/messaging/subscriptions.json"
[
  {
    "consumer": "loyalty.credit-collection",
    "event": "collection.order-collected",
    "command": "creditCollection"
  },
  {
    "consumer": "loyalty.issue-reward",
    "event": "loyalty.reward-earned",
    "command": "issueReward"
  },
  {
    "consumer": "loyalty.publish-reward-issued",
    "event": "loyalty.reward-issued.domain",
    "publication": "loyalty.reward-issued"
  },
  {
    "consumer": "loyalty.realtime-reward-issued",
    "event": "loyalty.reward-issued",
    "realtime": "reward"
  }
]
```

## Generate and verify a contract change

1. Edit the owning schema or private message declaration and the affected consumer
   manifests. Preserve historical field numbers and identities.
2. Update owner mapping, handler registration and behavioural tests.
3. Run `pnpm generate:contracts` for Protobuf/private-message changes, or
   `pnpm generate:http` for OpenAPI alone.
4. Commit generated outputs with their source changes. Run `pnpm check --changed`,
   focused behaviour checks, `pnpm verify` and the required infrastructure lane.
5. Check historical fixtures as well as generated drift. A clean regeneration
   does not by itself establish compatibility with old stored bytes.

Wholly generated packages use ordinary filenames beneath a `generated` directory.
Go plain request models live in `contracts/requests/generated/model/<owner>/`:
`payloads.go` defines inputs and `responses.go` defines replies and resource DTOs.
Event DTOs live in `contracts/events/generated/model/`. Generated encoders and
query registrations live under adaptor generated folders; mixed codec packages
retain `_generated.go` filenames. Keep generated headers and change the generator
or schema when changing these files. See
[decision 0030](../../decisions/0030-generated-package-layout.md).
Service selection keeps generated contracts local to the services that use them.
The application/domain never imports another service's source.

## Follow the event example

Define the integration schema, run generation, then write the
[domain-to-integration mapping](domain-event-handlers.md). Bind its generated
publication factory in composition and declare the private publication subscriber.
The generator creates code from the schema; the handler creates each runtime
payload from the delivered domain fact.

The [contract generation guide](../../../contracts/generation.md) and
[decision 0024](../../decisions/0024-generated-message-identifiers-and-direct-queries.md)
describe the complete sources, outputs and identifier rules.

## Generated DTOs in all three languages

These plain DTOs are generated from the owner schemas. Publication handlers
construct values of these types after private domain-event delivery.

### Go

Source: [ordering.go](../../../services/storefront/contracts/events/generated/model/ordering.go).

```go source="../../../services/storefront/contracts/events/generated/model/ordering.go"
// Code generated by scripts/generate_events.py. DO NOT EDIT.
package model

// Line carries published data owned by ordering.
type Line struct {
	ID        string `json:"id"`
	OfferCode string `json:"offerCode"`
	Name      string `json:"name"`
	Quantity  int    `json:"quantity"`
	Minor     int64  `json:"minor"`
}

// OrderPlacedIntegrationEvent carries published data owned by ordering.
type OrderPlacedIntegrationEvent struct {
	OrderID    string `json:"orderId"`
	CustomerID string `json:"customerId"`
	EditionID  string `json:"editionId"`
	Currency   string `json:"currency"`
	Lines      []Line `json:"lines"`
}
```

### Python

Source: [preparation.py](../../../services/operations/src/operations/contracts/events/generated/preparation.py).

```python source="../../../services/operations/src/operations/contracts/events/generated/preparation.py"
# Generated by scripts/generate_events.py. DO NOT EDIT.
from __future__ import annotations
from typing import TypedDict



class DrinksReadyIntegrationEvent(TypedDict):
    """Published DrinksReadyIntegrationEvent data owned by preparation."""
    orderId: str
    customerId: str
```

### TypeScript

Source: [loyalty.ts](../../../services/engagement/src/contracts/events/generated/loyalty.ts).

```ts source="../../../services/engagement/src/contracts/events/generated/loyalty.ts"
// Generated by scripts/generate_events.py. DO NOT EDIT.

/** Published RewardIssuedIntegrationEvent data owned by loyalty. */
export type RewardIssuedIntegrationEvent = {
  rewardId: string;
  customerId: string;
  benefit: string;
  expiresAt: string;
};
```

## Related concepts

[Commands](commands.md) · [Queries](queries.md) · [Domain events](domain-events.md) · [Integration events](integration-events.md) · [Realtime events](realtime-events.md)

[Concept catalogue](index.md) · [Learning path](../index.md)
