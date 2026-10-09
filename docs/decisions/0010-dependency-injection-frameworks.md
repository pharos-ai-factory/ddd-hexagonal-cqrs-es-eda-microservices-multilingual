# 0010 — Explicit dependency injection at composition roots

Status: accepted, 9 October 2026.

## Decision

Use **Awilix 13.0.5** for TypeScript Engagement, **Dependency Injector 4.49.1**
for Python Operations and **Uber Fx 1.24.0** for Go Storefront and API processes.
The frontend has no service container. Bootstrap/replay utilities remain small
one-shot programs. Framework dependencies and transitive versions are locked.

Domain and application classes retain ordinary constructors and explicit ports.
Framework imports, provider registrations, container resolution and resource
lifecycle belong under `apps/`. Each context receives its own database and
credentials. Handler constructors receive aggregate/query/command-delivery ports;
PostgreSQL and RabbitMQ appear in adaptor implementations and composition.

Awilix uses explicit factory registrations, strict mode and separate context
containers. The typed cradle helps callers; resolving the complete graph in tests
checks missing registrations. Each pool has a disposer. Worker cancellation and
completion precede container disposal. Dependency Injector uses explicit resource
and singleton providers; resolve shared providers before starting threads.
Provider overrides supply deterministic test ports. Fx declares context modules,
plain handler factories and lifecycle hooks. Graph validation runs without opening
infrastructure. Process shutdown stops workers before closing their resources.

Containers are constructed once per application instance. Handlers retain their
injected dependencies and execute directly on each delivery. Dependency resolution
is therefore outside the aggregate mutation path. The additional queue/transaction
cost comes from decision 0011; DI itself adds startup work and modest retained
objects. No throughput claim is made without measurement.

## Comparison

The assessment prioritises understandable wiring, maintenance, adoption and
runtime dependencies. These are dated observations from 9 October 2026, rather
than a permanent ranking. GitHub stars indicate visibility; package downloads
include CI and transitive installations. Direct dependencies exclude development
and optional integrations; transitive dependencies still enter the lockfile.

| TypeScript option | Ease and fit | Maintenance / adoption at assessment | Runtime dependencies |
| --- | --- | --- | --- |
| **Awilix 13.0.5** | Explicit factories, plain constructors, resource disposers; small registration API | Recent 2026 release/activity; about 4.2k stars and 616k weekly npm downloads | One direct dependency, `fast-glob`, plus its transitives |
| TSyringe 4.10.0 | Compact decorator syntax; factories also support plain handlers; factory resource disposal needs explicit care | Repository active in October 2026; about 6k stars and 14m weekly downloads | `tslib`; runtime `reflect-metadata` polyfill required |
| Inversify 8.2.3 | Rich bindings and lifecycle features; more concepts and configuration | Active 2026 monorepo; original repository about 12k stars; 2.45m weekly downloads | Three direct internal packages and their transitives; documented metadata setup |
| Typed Inject 5.0.0 | Explicit typed factories, compile-time graph checking; more type declarations | Smaller, quieter project; about 628 stars and 3.38m weekly downloads | Zero external runtime dependencies |

Awilix makes the binding graph visible while keeping classes independent of
framework syntax. TSyringe remains a reasonable option: decorators are its usual
style, and factory registrations can avoid decorating application classes. Our
choice favours explicit composition and resource ownership already used here.

Sources: [Awilix](https://github.com/jeffijoe/awilix),
[TSyringe](https://github.com/microsoft/tsyringe),
[Inversify](https://github.com/inversify/monorepo),
[Typed Inject](https://github.com/nicojs/typed-inject).
The download observation uses the npm statistics window 1–7 October 2026.

| Python option | Ease and fit | Maintenance / adoption at assessment | Runtime dependencies on Python ≥3.13 |
| --- | --- | --- | --- |
| **Dependency Injector 4.49.1** | Explicit providers, resource cleanup, overrides and typing support | Active October 2026; about 4.9k stars; largest project of this shortlist | Zero mandatory external packages; Cython extension with published wheels |
| Dishka 1.10.1 | Typed providers, explicit scopes and graph validation | Active October 2026; about 1.3k stars; growing community | Zero mandatory external packages |
| Injector 0.24.0 | Familiar Guice-style modules; resource ownership requires application code | Active October 2026; about 1.5k stars; established project | Zero mandatory external packages |
| Lagom 2.7.7 | Convenient autowiring and explicit-container alternative | Smaller, quieter project; about 356 stars | Zero mandatory external packages |

Dependency Injector's explicit resource providers fit context pools and the
threaded Pika workers. We avoid decorator wiring and container access in handlers.
Dishka's scope model is attractive for request-heavy applications. The current
workers mainly need application-lifetime resources and test overrides.

Sources: [Dependency Injector](https://python-dependency-injector.ets-labs.org/),
[Dishka](https://github.com/reagento/dishka),
[Injector](https://github.com/python-injector/injector),
[Lagom](https://github.com/meadsteve/lagom). Versions are published package versions;
conditional dependencies for older Python versions and optional extras differ.

| Go option | Ease and fit | Maintenance / adoption at assessment | Runtime dependencies |
| --- | --- | --- | --- |
| **Fx 1.24.0** | Modules, graph validation and ordered start/stop hooks; runtime reflection | Mature Uber project; about 7.7k stars | Dig, multierr, zap and x/sys |
| samber/do v2.1.0 | Generic typed registration/resolution, scopes and shutdown | Active October 2026; about 2.8k stars | Published v2.1.0 imports `go-type-to-string` |
| Dig 1.19.0 | Small container API; application lifecycle must be supplied separately | Mature Uber project; about 4.5k stars | Zero external runtime dependencies |
| Wire 0.7.0 | Generated construction with compile-time errors; generation step required | Archived August 2025; about 14.4k stars | Zero generated runtime dependencies; generator has dependencies |

Fx's lifecycle support covers the HTTP servers, database pools and delivery
workers. Wire's archived status excludes it from a new reference implementation.
A recent commit alone does not establish maintenance quality, and a mature
project can remain useful without frequent releases.

Sources: [Fx](https://uber-go.github.io/fx/), [Dig](https://github.com/uber-go/dig),
[samber/do](https://github.com/samber/do), [Wire](https://github.com/google/wire).
The published samber/do module dependency is counted even where overview prose
advertises no dependencies.

## Consequences

The composition modules add explicit registrations and framework upgrade work.
Tests can replace ports without importing a broker or opening a database.
Architectural checks reject framework imports in core packages. Typed graph tests,
resource lifecycle tests and the live integration lane remain necessary: a
container does not establish transaction or delivery correctness.
