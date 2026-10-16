# 0022 — Catalogue-scoped broker authority

Status: Accepted

Refines [0004](0004-runtime-authority-and-queries.md),
[0007](0007-api-broker-requests.md) and
[0017](0017-generated-event-boundaries.md).

## Decision

Provision separate RabbitMQ users for each owning context, the API and topology.
The host development launcher retains administrative provisioning authority.
Topology receives a dedicated credential with an exact generated resource
allow-list. Runtime credentials have no configure permission, no queue write
permission and no exchange read permission.

Generate permissions from the owner subscription catalogues and API request
contract selection. Include event queues, durable command queues, retries, dead
letters, command/query request queues and API reply queues. Escape and anchor
resource names, reject queue/exchange collisions and preserve the existing wire
identities. Grant each owner publication to its own `domain` and `integration`
namespace, its own reply key and its own delivery exchange. Grant API request
publication and consumption of the generated reply queues.

Explicit topic rules cover all three shared topic exchanges. Topology receives
exact catalogue binding keys and no topic publication authority. Private
subscriptions are validated against their event owner before generation. The
trusted catalogue/bootstrap enforces the queue/key pairing because RabbitMQ's
separate resource and topic permissions cannot express that relationship for a
single topology user.

Topology needs write on delivery exchanges and read on queues declared with
dead-letter arguments. Those capabilities also permit publication to those
exchanges and consumption from those queues. Document this required authority
and keep the credential confined to the short-lived bootstrap.

## Consequences and verification

New consumers require regeneration and permission provisioning. Existing
credentials survive development reconfiguration. The launcher stops managed
broker clients before applying permissions and restarts them afterwards to
replace cached authorisation. Retired queues retain their evidence for explicit
administrative cleanup while removed consumers lose runtime access.

Deterministic tests cover resource collisions, literal regular expressions,
complete queue families, negative ownership fixtures, generated drift and user
provisioning. Real RabbitMQ probes cover successful owner and API delivery,
restricted topology, denied binding changes, publication spoofing and arbitrary
queue access. Delivery recovery tests across all three owner runtimes use their
actual credentials and declared queues. See the
[permission model and workflow](../broker-permissions.md).

## Temporary API reply authority — 16 October 2026

API connections may configure their owner-scoped UUIDv4 reply queue family, write
those queues for binding and read `cafe.replies` with exact UUIDv4 topic keys.
Exclusive classic queues prevent another connection consuming those replies.
Owner publication permission covers its legacy reply key and its UUIDv4 reply
keys. Owner topology privileges remain empty. The production permission probes
and overlapping API connection tests enforce this exception.
