# Database bootstrap

The two SQL files establish the historical database foundation for this
local/development reference. They are deployment infrastructure. Their immutable
checksums and existing generated copies are verified by the migration tools.
Runtime credentials cannot apply migrations.

Each context owns its later migrations inside its language service:

- Go: `services/storefront/contexts/<owner>/adaptors/persistence/migrations/`
- Python: `services/operations/src/operations/contexts/<owner>/adaptors/persistence/migrations/`
- TypeScript: `services/engagement/src/contexts/<owner>/adaptors/persistence/migrations/`

The administrator applies bootstrap and context migrations before runtime startup.
Preserve the historical files; evolve a context through its own migration manifest.
See [runtime authority](../../../docs/decisions/0004-runtime-authority-and-queries.md)
and [development secrets](../../../docs/development-secrets.md).
