# Targeted realtime messages — 22 October 2026

[Decision 0031](../decisions/0031-targeted-realtime-messages.md) records the
application targets, gateway authority and typed browser handler API. Go, Python
and TypeScript publishers now select `context` or `principal` targets. The gateway
derives the Centrifugo channel from the authenticated owner, message payload and
target. The browser receives server-selected subscriptions through
`RealtimeProvider` and registers typed `useRealtimeMessage` handlers.

## Executed checks

- `pnpm verify` passed generated drift, historical Protobuf/OpenAPI and stored-byte
  compatibility, documentation, formatting, static analysis, architecture checks,
  173 tooling tests and both Go modules with the race detector.
- Operations passed 77 Python tests, Engagement passed 57 TypeScript tests, the web
  application passed 9 tests and the deterministic BDD lane passed 54 scenarios.
- Documentation checks rendered 146 pages, verified 136 source excerpts and checked
  navigation, links, search and Markdown exports.
- Gateway tests cover owner and payload validation, context/principal channel
  derivation, canonical principal identities and conflicting legacy channel hints.
  Codec tests cover target encoding in all three publisher languages and target/
  channel agreement in the browser.

## Environment-dependent checks

`pnpm test:integration` passed against the disposable Docker Compose stack. This
covered PostgreSQL and RabbitMQ integration, 31 live API operations across all six
contexts, realtime gateway permission probes, live Gherkin workflows, provider
recovery, browser command/realtime recovery and in-flight session revocation. The
harness removed its containers, network and volumes after the run.

UAT, Minikube/k3s and deployment checks were not run.
