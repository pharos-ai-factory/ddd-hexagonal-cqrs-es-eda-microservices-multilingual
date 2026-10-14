# 0018 — Static checks and formatting after edits

Status: accepted, 14 October 2026. Extends decision 0012.

## Shared gate

`pnpm check` checks all handwritten formatting, Python and TypeScript types,
language lint rules and architectural boundaries. It runs without live services.
`pnpm check <service>` checks the complete owning project and global boundaries;
`pnpm check --changed` selects owners from tracked edits, deletions and untracked
files. Shared contracts, configuration and tooling select all consumers.
Independent checks continue after a failure and the command exits unsuccessfully.

`pnpm verify` starts with this same gate, followed by generated drift,
compatibility, negative fixtures and native/race/Gherkin tests. CI invokes
`pnpm verify`. Real transaction, delivery and browser evidence remains in
`pnpm test:integration`.

## Language rules

- Operations retains strict Mypy over source and tests, unreachable-code checks
  and error-code-specific suppressions. Ruff checks Operations and repository
  Python tooling. Standalone infrastructure entry points may initialise their
  import path before importing repository helpers.
- Every TypeScript project extends `tsconfig.strict.json`: strict types, checked
  indexing, exact optional properties, complete return paths, explicit overrides
  and switch fallthrough checks. Engagement's checking project includes adjacent
  tests, Gherkin bindings and the pool probe; the build retains its test exclusions.
- ESLint uses the recommended rules with type information plus exhaustive
  switches. Unhandled/misused promises and unsafe `any` operations are errors.
  Test doubles may implement asynchronous ports without awaiting artificial I/O.
  Named repository interfaces may extend one generic capability without adding
  members. Other exceptions remain local and explain their purpose.
- Both Go modules run vet, pinned Staticcheck and pinned errcheck. Go's existing
  compiler and race-enabled test gates remain. Best-effort cleanup explicitly
  discards errors where the operation has already completed or failed.

Generated declarations still participate in compilation and imported type checking.
Their generators own their formatting and lint conventions; regeneration and
compatibility gates protect those outputs. Handwritten source retains the
450-line responsibility limit.

## Formatting and edit feedback

Go uses gofmt, Python uses Ruff and TypeScript/documentation/configuration use
Prettier. `pnpm format:check --all` checks the entire handwritten baseline, including
untracked files and clean checkouts. `pnpm format` formats changed and new files;
explicit file paths and service selection are also supported. Existing handwritten
source is normalised when introducing this gate.

`pnpm check:watch [service ...]` runs initially and after debounced edits. It observes
file additions and deletions, rechecks complete affected projects, notices edits
made while a check runs and remains active after failures. `pnpm dev:watch` starts
these checks beside Compose rebuilds and stops them on exit. Focused tests first
check their owning service. Agents run `pnpm check --changed` after their edit batch.
VS Code recommends the formatters, ESLint and Mypy extensions, enables formatting
on save and exposes explicit check/watch tasks. CLI checks remain the shared gate.

Regression fixtures cover changed-file selection, watcher recovery, TypeScript
test inclusion, strict compiler failures and typed lint failures.
