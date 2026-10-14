# 0017 — Contract-driven event boundaries

Accepted: 2026-10-14.

## Decision

Each published event message declares its name, producing context, aggregate kind
and contract version through `cafe.events.event` in its context-owned Protobuf
source. Portable field constraints use `cafe.events.rules`. Compiler descriptors
are the authority for event registries, plain published DTOs, typed conversions
and publication factories. Handwritten catalogues and payload-name dictionaries
are removed.

The shared event envelope carries technical metadata and `bytes payload`.
It imports no business payload schemas. Services generate only their declared
contract families; generated registrations bind a supported name and version to
its concrete payload. Unknown contracts, wrong owners and unsupported versions
fail before application handling. Source revisions and payload integers use the
JavaScript safe-integer range across all languages.

Plain published DTOs are generated in service-local contract modules grouped by
producer. They have no Protobuf dependencies. Domain models, private facts,
commands, queries and read models remain independently owned. Adaptor generation
supplies mechanical conversions and typed publication factories. Owner publication
mappers still choose which domain facts to publish. Consumer reactions still map
published facts into durable owner commands.

Consumer-owned `adaptors/messaging/subscriptions.json` declarations supply event
bindings. Go projection subscriptions omit `command`; Python and TypeScript
aggregate reactions name their queued command. Topology and replay metadata join
these declarations to event definitions. A producer has no list of consumers;
multiple independent subscriptions can consume the same fact. Private subscriptions
remain restricted to the owning context.

Shared handwritten codecs handle the envelope, size limits, schema constraints
and serialisation. Event-specific semantic checks live in context adaptors:
Preparation checks accepted-order line uniqueness and the total quantity;
Ordering checks unique offer codes and consistent currency before projection.
Private codecs belong to their owning contexts and use the common delivery path.
Generated registries compose those codecs without handwritten central event lists.

## Generator scope

The generator reads actual compiler descriptors, including custom options. The
initial portable payload vocabulary is strings, booleans, signed int32/int64,
unsigned uint32/uint64, singular messages and repeated messages. Scalar defaults
are Protobuf defaults; repeated messages decode to complete arrays. Required
identities and other boundary requirements are explicit constraints.

Optional presence, enums, maps, scalar arrays, nested declarations and recursive
payloads require a reviewed generator extension. Generation rejects these forms
rather than silently changing their meaning. New schema files in an existing
context family are discovered automatically. A new service dependency on another
family still requires its explicit generation declaration.

Schema constraints describe accepted wire evidence: UUIDs, source identity,
non-empty strings, timestamps with offsets, portable regular expressions, integer
bounds and collection bounds. The aggregate remains responsible for its invariant.
These constraints do not execute a business transition or call application code.

## Reference redesign

This replaces the original global payload `oneof`; existing public event bytes
use a different envelope. The shared RewardIssued fixture is regenerated for the
new format. There is no legacy public decoder or migration path in this reference.
Development environments should begin with fresh delivery state when adopting
this change. Private formats and request/realtime interfaces are unchanged.

The historical comparison explicitly recognises the transition from the old
public `oneof` to payload bytes and the addition of boundary annotations. It
continues checking existing metadata and payload field identities and unrelated
interfaces. Comparisons against a current-format revision receive the ordinary
compatibility checks. The transition is recorded in `scripts/event_redesign.py`.

## Verification

Isolated regeneration detects missing, stale and changed outputs. Generator tests
add an event in a new owned schema without editing the envelope or registry;
negative cases cover ownership, constraints and subscription authority. Native
codec tests cover malformed input and round trips. The shared fixture is decoded
in Go, Python and TypeScript. Real PostgreSQL/RabbitMQ and browser lanes prove
atomic outgoing intent, durable command hand-off, retries and end-to-end behaviour.
