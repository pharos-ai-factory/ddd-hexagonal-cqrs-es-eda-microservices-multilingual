# 0002 — Service-first polyglot monorepo and browser delivery

Status: accepted, 27 September 2026. Runtime composition and query reconciliation
are refined by [0004](0004-runtime-authority-and-queries.md). Clarifies the deployment/language scope of
0001; its aggregate and event-delivery decisions remain in force.

## Service boundary

Use three domain microservices with two contexts each: Go Storefront (Menu,
Ordering), Python Operations (Preparation, Collection), TypeScript Engagement
(Loyalty, Customer Communication). The Go client API is a separate fourth backend
process. Next.js is a fifth application, responsible for browser presentation.

Organise source under `services/<service>`, with contexts nested inside their
owner. This keeps build tools, dependencies and deployment boundaries explicit.
The root shares wire schemas, persistence specifications and behavioural tests,
not domain implementation. There is no root contexts directory mixing languages.

Generated transport bindings are local build inputs. Domain and application
packages remain framework/provider independent in each language. Co-location
never grants a context access to another context's database.

Operations expresses commands, outcomes, published DTOs and snapshots with
explicit Python types and keeps domain facts in immutable context-owned
dataclasses. Generic ports carry the aggregate-specific snapshot type. Strict
Mypy checks cover handwritten runtime code and native tests. Runtime adaptors
still validate untrusted JSON, stored state and Protobuf values; annotations
alone are not validation. Typed conversion does not change persisted JSON,
published bytes or the original payload used to fingerprint a command receipt.
Protobuf `.pyi` declarations are generated from the canonical schemas and remain
beside the adaptor bindings.

## API and browser

The Go API owns technical session state in Valkey, explicit allowed API routes,
service credential replacement and server-side subscription authorisation.
It owns no business database or broker connection. Next.js contains no domain
command authority and does not route around the API.

Use Centrifugo binary transport, a connect proxy and server-selected
subscriptions. Each root transition commits immutable, typed browser projection
bytes and a dispatch record atomically. Fenced workers deliver those snapshots
to Centrifugo. This is a distinct delivery destination from RabbitMQ, while both
obey durable intent, retry and duplicate-tolerance requirements.

Per-root versions guard snapshot application. Per-channel epoch/offset drives
transport recovery. Initial/unrecoverable attachment reconciles from owner
queries once; recovered history does not introduce business polling. Every
window owns its independent connection and position.

## Development simplifications

A server-provisioned operator has access to all six context channels. An opaque,
fixed-lifetime cookie session replaces a production identity-provider integration
for this example. Logout and expiry still record durable disconnection work.
Connect grants expire after one second, with the deadline calculated before
reading session authority. A worker observing revocation durably records a wait
past that validity window before completing its disconnect, so a delayed grant
cannot establish subscriptions after completion. Established connections use
the server-side refresh proxy to revalidate their original cookie session.
The validity window and worker deadline use the development host's clock.
Customer IDs on screen select domain subjects; they do not confer authority.

The composition is local/development-only and has no production release workflow.
Single-node infrastructure, a local notification simulator, generic JSONB roots
and absent event sourcing remain explicit constraints.
