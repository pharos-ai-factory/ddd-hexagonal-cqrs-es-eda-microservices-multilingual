# HTTP behaviour contract, version 1

The separate Go API exposes `/api/v1/<context>/...` on port 28080 and through
browser ingress on port 28000. Browser requests use the operator cookie; mutations
require an allowed Origin. Scripted clients use `Authorization: Bearer <API_KEY>`
from `.local/dev.env`. This API selects a known owner and forwards the command
with that service's distinct internal key, preserving its identity and version.

The service-local routes below omit the API's `/api` prefix. Direct diagnostic
access requires `STOREFRONT_API_KEY`, `OPERATIONS_API_KEY` or `ENGAGEMENT_API_KEY`,
respectively. Service credentials never enter the browser. Health endpoints are
unauthenticated. Session routes and realtime authority are documented in
`CLIENT-SUBSCRIPTIONS.md`.

Commands use `POST` and a JSON object. Empty commands require `{}`. Unknown body
fields, trailing JSON and bodies larger than 64 KiB are rejected.

## Command identity and concurrency

Every command requires:

| Header | Meaning |
| --- | --- |
| `Idempotency-Key` | Canonical lowercase UUID for this command attempt |
| `If-Match` | Expected aggregate version; `0` creates a new aggregate |
| `X-Correlation-ID` | Optional UUID for the workflow; defaults to the command key |

The URL identifies the aggregate root. An OrderLine ID in the body never changes
the command's aggregate target.

The idempotency scope is context, aggregate kind, aggregate ID, command name and
command ID. Repeating that identity with the same expected version and material
input returns its original outcome, even after the aggregate advances. Reusing
it with changed input returns `idempotency_conflict`.

Successful responses are HTTP 200:

```json
{
  "aggregateId": "11111111-1111-4111-8111-111111111111",
  "version": 3,
  "status": "placed"
}
```

| Status | Interpretation |
| --- | --- |
| 400 | Invalid body or identifier; no command decision |
| 401 | Missing or incorrect development key |
| 404 | Requested aggregate does not exist |
| 409 | Stale expected version or conflicting command identity |
| 422 | Typed business rejection in `rejection.code` and `rejection.message` |
| 428 | Missing or invalid expected version |
| 503 | Infrastructure unavailable; retry the same command identity |

Recorded business rejections do not change state/version or emit events.
For `menu_pending` and `drink_revision_pending`, allow the projection to arrive,
then issue a **new command attempt with a new key**. Retrying the old rejected
identity does not re-evaluate its decision.

## Storefront, port 28081

All paths below begin with `/v1`.

| POST path | Body |
| --- | --- |
| `/menu/drinks/{id}` | `{"name":"Cappuccino"}` |
| `/menu/drinks/{id}/revise` | `{"name":"Large cappuccino"}` |
| `/menu/drinks/{id}/publish` | `{}` |
| `/menu/editions/{id}` | `{"currency":"EUR"}` |
| `/menu/editions/{id}/offers` | `{"code":"C1","drinkId":"<uuid>","drinkRevision":1,"minor":300}` |
| `/menu/editions/{id}/prices` | `{"code":"C1","minor":350}` |
| `/menu/editions/{id}/publish` | `{}` |
| `/ordering/orders/{id}` | `{"customerId":"<uuid>","editionId":"<uuid>"}` |
| `/ordering/orders/{id}/lines` | `{"lineId":"<uuid>","editionId":"<uuid>","offerCode":"C1","quantity":2}` |
| `/ordering/orders/{id}/quantities` | `{"lineId":"<uuid>","quantity":3}` |
| `/ordering/orders/{id}/place` | `{}` |

Money uses integer minor units. Adding an offer or changing its price requires
an explicit non-null `minor` field; zero is valid. Drink names contain 1–80
Unicode characters. Offer codes have 1–8 uppercase alphanumeric
characters and begin with a letter. Published editions hold 1–20 offers.
Draft Orders may be empty; placed Orders contain 1–5 drinks in total.

## Operations, port 28082

| POST path | Body |
| --- | --- |
| `/v1/preparation/tickets/{id}/start` | `{}` |
| `/v1/preparation/tickets/{id}/complete` | `{}` |
| `/v1/collection/pickups/{id}/collect` | `{"code":"ABC123"}` |

Tickets and pickups are created by event consumers. Their IDs differ from the
Order ID. The read DTOs expose `orderId` for the operator and test observer.

## Engagement, port 28083

| POST path | Body |
| --- | --- |
| `/v1/loyalty/rewards/{id}/redeem` | `{"orderId":"<uuid>"}` |

Accounts, rewards and notification intents are created by event consumers.
The account ID equals the canonical customer ID. Reward validity begins at
issuance. Redemption records use against an order identity; no order discount
workflow or external order validation is implemented.

## Queries

Each resource below supports `GET <resource>` and `GET <resource>/{id}`:

- Storefront: `/v1/menu/drinks`, `/v1/menu/editions`, `/v1/ordering/orders`.
- Operations: `/v1/preparation/tickets`, `/v1/collection/pickups`.
- Engagement: `/v1/loyalty/accounts`, `/v1/loyalty/rewards`,
  `/v1/communication/notifications`.

An item is `{"exists":true,"version":3,"state":{...}}`; a list is an array of
items. `state` is the JSON snapshot declared in each context's domain package.
Queries never advance an aggregate.

Without pagination parameters, lists return their complete array in aggregate-ID
order. Queries opt into pagination separately from the ordinary query port:

```text
GET /v1/menu/drinks?limit=100
GET /v1/menu/drinks?limit=100&cursor=<nextCursor>
```

Paginated responses are `{"items":[...],"nextCursor":"..."}`. The last page has
`nextCursor: null`; `limit` must be an integer from 1 to 100. A cursor is opaque,
versioned and bound to its resource. Invalid cursors, duplicate pagination
parameters and foreign-resource cursors return 400. A cursor must be accompanied by `limit`. UUID keysets
avoid offset shifts; a scan is not a single snapshot across requests. Existing
item queries and explicitly unpaginated list queries retain their shapes.

The browser's `query<T>` helper supports any ordinary response and `queryAll<T>`
traverses paginated results. Pagination does not introduce periodic polling.
The scripted journey uses fresh IDs and observes its records within bounded
deadlines. `GET /healthz` means the HTTP process has started; it is not proof that
every queue has drained or every workflow has completed.

## Language conformance

`scripts/journey.py` is the executable happy-path and recovery-observation
contract. Implementations must preserve endpoint shapes, typed rejections,
versions, recorded retry outcomes and eventual workflow results. They need not
share SQL table layouts or Go struct names.
