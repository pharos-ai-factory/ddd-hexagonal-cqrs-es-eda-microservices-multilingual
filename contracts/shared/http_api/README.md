# HTTP behaviour contract, version 1

The OpenAPI 3.0.3 documents are the authoritative HTTP wire contracts:

| Document | Surface |
| --- | --- |
| [api.openapi.json](../../services/api/http_api/api.openapi.json) | Go API: all six contexts' HTTP commands and queries, sessions, health, diagnostics and internal Centrifugo proxies |
| [storefront.openapi.json](../../services/storefront/http_api/storefront.openapi.json) | Direct Go Storefront commands, queries, health and diagnostics |
| [gateway.openapi.json](../../services/realtime_gateway/http_api/gateway.openapi.json) | Internal Go realtime publication and disconnection gateway |
| [provider.openapi.json](../../services/notification_provider/http_api/provider.openapi.json) | Go development notification provider and its failure controls |

Start with a complete document above to inspect an API. To change a contract,
find its owner below. Resource paths and their schemas sit together:

| Change | Edit |
| --- | --- |
| Drinks and menu editions | [menu/http_api](../../menu/http_api/) |
| Orders and order lines | [ordering/http_api](../../ordering/http_api/) |
| Preparation tickets | [preparation/http_api](../../preparation/http_api/) |
| Pickups and collection | [collection/http_api](../../collection/http_api/) |
| Loyalty accounts and rewards | [loyalty/http_api](../../loyalty/http_api/) |
| Notifications | [communication/http_api](../../communication/http_api/) |
| API health and diagnostics | [operational.paths.json](../../services/api/http_api/operational.paths.json) and [operational.schemas.json](../../services/api/http_api/operational.schemas.json) |
| Login, session and logout | [session.paths.json](../../services/api/http_api/session.paths.json) |
| Centrifugo connect and refresh | [realtime.paths.json](../../services/api/http_api/realtime.paths.json) |
| Storefront health and diagnostics | [operational.paths.json](../../services/storefront/http_api/operational.paths.json) and [operational.schemas.json](../../services/storefront/http_api/operational.schemas.json) |
| Common errors and outcomes | [schemas.json](schemas.json) and [responses.json](responses.json) |
| Command headers and pagination parameters | [parameters.json](parameters.json) |

Each `<resource>.paths.json` defines that resource's operations; `schemas.json`
defines its context's published request and response shapes. A complete document
references these fragments and supplies its own URL prefixes and security schemes.
Adding a path requires an entry in each complete document that exposes it.

The API and direct Storefront document reference the same business fragments.
Python and TypeScript business operations exposed through the Go API also use
their owning context's published schemas. Framework-specific administrative
surfaces in those services are outside the Go documents.

Service-specific session, operational and realtime contracts live under
`contracts/services/<service>/http_api/`. Generic transport components live here. Keep each
context's business DTOs with its owning context.

Go compositions embed self-contained bundles generated from these documents.
Route registration fails for an undeclared method/path or a declared operation
without a handler. API translation selects exact operations from the contract;
each configured owner must identify a declared context. Authentication remains
at the HTTP boundary; broker credentials authorise internal requests and replies.

`pnpm verify` checks bundle drift and runs OpenAPI request/response conformance
tests against every Go HTTP operation. Tests validate response bodies, content
types and declared status codes, including typed rejection envelopes, and prove
that malformed wire shapes fail validation. `pnpm test:integration` also checks
live responses from all six contexts through the API after the workflow journey,
including full lists, item queries, cursor traversal and command error outcomes.

Edit the source documents, run `pnpm generate:http`, and run both required gates.
`pnpm generate:contracts` includes the same HTTP bundling step. Generated bundles
are committed inside each owning Go module and must remain reproducible. Normal
builds need no generator or repository-root files.

Generated component names include their source path, allowing different contexts
to define the same schema name independently. The bundler rejects ambiguous names
instead of silently substituting another owner's schema.

Schemas describe transport values. Owners retain business decisions such as
quantity bounds, drink naming, currency rules and lifecycle transitions, returning
recorded 422 outcomes. The existing HTTP body limits, UUID validation, concurrency
headers and authentication are exercised separately. Response buffering and
schema validation run in the conformance checks.

Go API technical failures use the shared `ErrorResponse` struct and status/code
descriptors in its HTTP adaptor. The body remains `{"code":"..."}` with an
optional public `message`; underlying infrastructure errors stay outside it.
Session authentication and Centrifugo replies retain their documented protocol
envelopes. The API returns owner command rejections with their aggregate identity,
version and `rejection` details.

The separate Go API exposes `/api/v1/<context>/...` on port 28080 and through
browser ingress on port 28000. Browser requests use the operator cookie; mutations
require an allowed Origin. Scripted clients use `Authorization: Bearer <API_KEY>`
from `.local/dev.env`. This API selects a known owner and translates the command
into Protobuf over RabbitMQ. The owner adaptor constructs the application command.
Command identity and expected version survive translation. Browser credentials
stay at the HTTP boundary. See [request/reply contracts](../messaging/requests.md).

The service-local routes below omit the API's `/api` prefix. Direct diagnostic
access requires `STOREFRONT_API_KEY`, `OPERATIONS_API_KEY` or `ENGAGEMENT_API_KEY`,
respectively. Service credentials never enter the browser. Health endpoints are
unauthenticated. Session routes and realtime authority are documented in
[CLIENT-SUBSCRIPTIONS.md](../../../CLIENT-SUBSCRIPTIONS.md).

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
items. The OpenAPI schema defines the public `state` shape. Owner query handlers
return application-owned read models; persistence adaptors validate stored authority
and map those fields. Queries never advance an aggregate.

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

The browser uses generated operation names, request/response types and route
metadata through its HTTP adaptor. Typed list operations traverse paginated
results and retain revision guards during reconciliation. Pagination does not introduce periodic polling.
The scripted journey uses fresh IDs and observes its records within bounded
deadlines. `GET /healthz` means the HTTP process has started; it is not proof that
every queue has drained or every workflow has completed.

## Language conformance

`scripts/journey.py` is the executable happy-path and recovery-observation
contract. Implementations must preserve endpoint shapes, typed rejections,
versions, recorded retry outcomes and eventual workflow results. They need not
share SQL table layouts or Go struct names.


## Enforcement commands

`pnpm generate:http` rebuilds embedded OpenAPI bundles, frontend
`services/web/src/adaptors/generated/http.ts` and typed API mappings under
`services/api/adaptors/http/backend/generated/`. Edit
`services/api/adaptors/http/backend/contract_mappings.json` when an HTTP operation
needs a different wire mapping. Owner command/query DTO mappings live in each
context's messaging adaptor.

`pnpm check:http` checks these HTTP outputs. `pnpm check:contracts` checks all
contract outputs, including new and removed files. `pnpm verify` runs generated
drift checks, frontend type checking and real-feature compilation tests with
incompatible schema changes, plus Go startup coverage, request validation and
operation/response conformance examples. `pnpm test:integration` checks live
business responses through RabbitMQ for all 31 OpenAPI operations.
