# Documentation and concept reference, 15 October 2026

Rspress 2.0.23 now builds the repository's Markdown into a local searchable site.
The stable Mermaid plugin renders browser SVGs. The education section contains
nine introductory guides, a glossary and 31 linked concept pages with Go, Python
and TypeScript examples. The [coverage report](../education/implementation-coverage.md)
identifies implementation gaps and distinguishes illustrative examples from
existing source.

## Executed checks

- `pnpm check --changed` passed after the documentation and tooling changes.
  The documentation gate checks tooling types/lint, link and excerpt regression
  tests, every Markdown page, navigation, search output and Markdown exports.
  It verifies 112 marked source excerpts and requires all three languages on
  every concept page.
- `pnpm verify` passed for all five applications, including generated drift,
  historical contract compatibility, architecture checks, native tests and
  51 fast Gherkin scenarios. The final run includes the expanded concept pages.
- `pnpm test:integration` passed PostgreSQL/RabbitMQ checks, all 31 live API
  operations, 12 infrastructure Gherkin scenarios, provider recovery, both
  browser journeys and the separate session-revocation fixture. Disposable
  containers, networks and volumes were removed. No required environment-dependent
  checks were skipped. This run preceded the education-only expansion; service
  implementations and published contracts remain unchanged.
- Browser inspection verified the concept catalogue, all three command-definition
  and registration examples, GitHub source links, search results and navigation
  to the command-handler page.
- Mermaid flowcharts for context boundaries, CQRS, modelling, event delivery,
  integration publication and write repositories rendered without error in both
  light and dark themes. The delivery sequence diagram rendered seven messages.
  Long introductory flows use vertical layouts to keep labels readable.
- `pnpm docs:dev` starts on loopback port 3001 and serves the new pages. Its
  repository watcher excludes dependency, generated, local artefact and Python
  cache directories; the original unrestricted watcher exhausted file watches.

The final documentation build also includes this evidence page. Build output,
browser captures and command logs remain under ignored `.local/`. Remote CI,
load tests and deployment were not run.

Logs: `.local/rspress-final-check.log`, `.local/rspress-final-verify.log`,
`.local/rspress-integration.log` and `.local/rspress-dev.log`.

See [decision 0025](../decisions/0025-repository-documentation-site.md) and the
[authoring guide](../documentation.md).

## Command creation recipe follow-up

The command page now provides eight ordered steps, with source-backed schema,
generated identifier, parser/import and registration examples in all three
languages. It distinguishes the application DTO, generated payload type, routing
identifier and handwritten parser, then covers API exposure and dispatch tests.
The developer workflow and related concept links follow the same sequence.

`pnpm check --changed` passed with 131 pages and 129 checked source excerpts.
This follow-up changes documentation only. The full verification and integration
results above remain the runtime evidence; those suites were not repeated for
the recipe edit. Log: `.local/rspress-command-recipe-check.log`.
