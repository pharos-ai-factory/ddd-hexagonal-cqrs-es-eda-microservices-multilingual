# 0005 — Typed expected errors within service foundations

Status: accepted, 29 September 2026.

## Decision

Each language service owns a small expected-error foundation. Domain and
application errors can have named, context-owned classes or structs when a rule
or use-case failure benefits from a concrete type. A command port recognises a
shared expected-error contract and records its stable `code` and `message` in
the existing outcome and receipt. The code is the durable identity; class names
and exception objects do not cross persistence, HTTP or message boundaries.

Domain errors express aggregate rules. Application errors express expected
command failures such as a missing target or stale expected version. Boundary
input validation can continue to use the generic rejection type. Corrupt
authority, unexpected exceptions and infrastructure failures are not expected
errors; they roll back and remain retryable.

Keep each foundation inside its owning language service. Business-specific
errors belong to the owning context. Do not introduce a cross-language runtime
package or adopt generic CRUD or event-buffer abstractions in place of the
one-aggregate command ports and transactional outbox.

## Consequences

Named errors can be matched by local callers and tests without changing the
published code or message. Recorded outcomes still replay as plain data, so
consumers and HTTP adaptors classify them by stable code. A new subtype must
implement the expected-error contract to be recorded; an ordinary error remains
a retryable failure.
