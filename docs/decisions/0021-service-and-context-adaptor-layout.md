# 0021 — Service and context adaptor responsibilities

Status: Accepted

Extends [0020](0020-operations-adaptor-layout.md) across all five services and six
contexts; refines the source layout in [0014](0014-context-navigation-and-query-layout.md).

## Decision

Every service exposes its external boundaries under `adaptors/` (beneath `src/`
for TypeScript and the Python package). Shared persistence, messaging, realtime
and operational HTTP have separate folders. Go retains small packages beneath
those roles. Technical foundation packages hold inward-facing primitives and
cancellation timing. API session and OpenAPI packages keep their existing roles.

Every context adaptor root exposes `messaging/`, `persistence/` and, where needed,
`delivery/` for outbound provider calls. Resource names identify query endpoints,
request/reply mappings, snapshot restoration, repositories and private fact
mappings. One private command codec lives in each business-named file. Shared
codec helpers and subscription identity declarations have explicit filenames.
Go keeps query endpoint and publication packages beneath messaging to avoid
cycles with persistence; its repositories and migrations share `persistence/`.

Communication owns `delivery/http-notification-delivery.ts`. Engagement splits
broker consumers, confirmed publication, outbox relays, request envelopes and
request registration; database claims and reply completion belong to persistence.
Storefront's shared technical implementations move from foundation and request
integrations into its adaptor tree. Realtime gateway delivery has a separate
package. The API exposes session storage, revocation work and gateway disconnects
in separate files. Web owns its HTTP operation client and Protobuf projection
codec under adaptors; React recovery and reconciliation remain in `shared/`.

Directories express responsibility and filenames identify the operation or
resource. The [adaptor map](../adaptor-layout.md) and each context's README explain
where to start. Business invariants and named application handlers remain the
primary entry points to a context.

## Compatibility and enforcement

Published and private wire formats, subscription identities, database migrations,
transaction boundaries, lease fencing, acknowledgement order and session
revocation fences retain their behaviour. Generators, integration lanes and
architecture checks follow the relocated sources. Generated files retain explicit
classification and are regenerated from their sources.

Architecture checks reject flat boundary modules, misplaced context query
endpoints, context dependencies in shared adaptors, network drivers in shared
persistence and database drivers in shared delivery. Negative fixtures cover Go,
Python and TypeScript. Go's command-boundary check explicitly recognises the
relocated shared executor and PostgreSQL implementation as infrastructure.

Run all six focused context lanes, the API and web lanes, static checks, full
verification and real-infrastructure tests. Record execution in the verification
index. No aggregate model or pending workflow state changes as part of this layout.
