# Browser OpenAPI client

The web application uses Orval 8 with the TanStack Query v5 client. The API's
assembled OpenAPI document remains authoritative. Run the repository generator;
do not run Orval against a context fragment:

```sh
pnpm generate:http
```

`pnpm generate:contracts` includes the same step. Both commands assemble the API
document, generate the Go HTTP mappings and run Orval with
`services/web/orval.config.ts`. Generated browser files are committed under
`services/web/src/adaptors/generated/http/` and checked for exact drift.

## Tags and output layout

Every API operation has one OpenAPI tag. Business operations use their owning
context name; technical operations use `session`, `realtime` or `operational`.
Orval's `tags-split` mode creates one client module per tag:

```text
adaptors/generated/http/
├── menu/menu.ts                 # fetch functions and TanStack Query hooks/options
├── menu/menu.msw.ts             # MSW handlers
├── menu/menu.faker.ts           # Faker response factories
├── model/menu/                  # tag-owned models
├── model/                       # shared models and indexes
├── operations.ts                # typed operation manifest
├── index.msw.ts                 # all mock-handler groups
├── index.faker.ts               # all Faker factories
└── index.fixtures.ts            # typed fixture factories
```

The same structure is generated for all six business contexts and the three
technical tags. Add a tag to the authoritative operation and declare it in the
root API document when adding a new group. Generation tests require one declared
tag per operation and complete use of the declaration list.

## Queries and cache ownership

Import a client from its generated tag module. Generated query hooks return the
HTTP result wrapper, so the response body is the nested `data` value:

```tsx
import { useListDrinks } from '../adaptors/generated/http/menu/menu';

const query = useListDrinks({ limit: 100 });
const response = query.data?.data;
```

The application-level `QueryProvider` disables automatic refetches and retries.
This preserves the browser delivery model: Centrifugo supplies live changes and
an explicit reconciliation traverses owner queries after attachment or a history
gap. Reconciliation uses generated request functions and query keys through the
TanStack `QueryClient`; realtime handlers merge newer revisions into the same
application cache.

Generated mutation functions are suitable for ordinary HTTP mutations. Café
business commands retain a raw `Response` until their durable outcome is known,
because 401, expected rejections and uncertain 5xx responses have different
recovery rules. `adaptors/http/operation-client.ts` therefore uses the generated
operation manifest and input/output types while owning that response handling.

## MSW, Faker and fixtures

Orval generates MSW handlers and Faker factories for every tag. A Node test can
start a complete context mock with the generated group:

```ts
import { setupServer } from 'msw/node';
import { getMenuMock } from '../adaptors/generated/http/index.msw';

const server = setupServer(...getMenuMock());
```

Pass a value or callback to an individual generated handler to override its
successful response. Import factories such as `getListDrinksResponseMock` from
`index.faker.ts` when a test needs typed data without an HTTP server. The generated
fixture barrel exposes equivalent typed fixture builders. Use a handwritten MSW
handler before the generated defaults for error statuses or scenario-specific
state transitions.

Mocks use schema examples where the contract supplies them and Faker for the
remaining values. They support component and adaptor tests; live conformance and
integration tests still exercise the Go API and owner services.

## Change and verification workflow

1. Edit the owning fragment under `contracts/<context>/http_api/` or the relevant
   technical API fragment.
2. Run `pnpm generate:http` and review the assembled specification and the full
   generated browser tree.
3. Update browser callers and MSW scenario overrides as required.
4. Run `pnpm check:http`, `pnpm test:focused web` and `pnpm check --changed`.
5. Complete `pnpm verify` and `pnpm test:integration` before hand-off.

Schema-mutation tests compile real feature, command and reconciliation callers.
They cover incompatible bodies and responses, required headers and query
parameters, and compatible optional additions. `pnpm check:contracts` regenerates
the client in isolation and detects changed, missing and unexpected files.
