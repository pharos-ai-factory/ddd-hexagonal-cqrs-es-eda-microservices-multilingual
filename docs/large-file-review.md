# Large-file responsibility review

Reviewed on 27 September 2026. Handwritten files have a 450-line limit.
There are no handwritten exceptions.

The following exact paths are classified as generated source or dependency locks:

| Path | Responsibility |
| --- | --- |
| `services/storefront/contracts/events/generated/cafe/v1/events.pb.go` | Go event bindings |
| `services/storefront/contracts/realtime/generated/cafe/realtime/v1/realtime.pb.go` | Go browser bindings |
| `services/operations/src/operations/adaptors/generated/cafe/v1/events_pb2.py` | Python event bindings |
| `services/operations/src/operations/adaptors/generated/cafe/realtime/v1/realtime_pb2.py` | Python browser bindings |
| `services/operations/src/operations/adaptors/generated/cafe/v1/events_pb2.pyi` | Generated Python event type declarations |
| `services/operations/src/operations/adaptors/generated/cafe/realtime/v1/realtime_pb2.pyi` | Generated Python browser type declarations |
| `services/engagement/src/adaptors/generated/events.json` | TypeScript event descriptor |
| `services/engagement/src/adaptors/generated/realtime.json` | TypeScript browser descriptor |
| `services/web/src/adaptors/generated/realtime.json` | Browser decoder descriptor |
| `pnpm-lock.yaml` | Reproducible Node dependency resolution |
| `services/operations/uv.lock` | Reproducible Python dependency resolution |

Generators own these structures. Splitting them by hand would introduce a second
source of truth. `scripts/check_architecture.py` lists exact exceptions rather
than exempting all files in a generated directory. `pnpm generate:contracts`
reproduces the bindings and descriptors; frozen package-manager installs verify
lockfiles. CI checks generated drift. `.local`, tool caches and build outputs
are excluded, and are never contributor source.

`pnpm check:large-file-review` runs this check alongside dependency rules;
`pnpm verify` includes it. New handwritten exceptions require an explicit
responsibility review and a corresponding checker change.
