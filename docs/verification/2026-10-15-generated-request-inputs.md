# Schema-generated command and query inputs — 15 October 2026

Implemented [decision 0028](../decisions/0028-schema-generated-command-and-query-inputs.md)
across all six contexts and three owner languages.

## Changes verified

- All 22 command payloads (15 public and seven private) and 16 query inputs derive
  from their annotated owner Protobuf definitions. Application handlers import
  generated plain types; handwritten input-field copies were removed.
- Generated boundary mappings preserve scalar types, required presence, optional
  fields and the distinction between complete reads and requested pagination.
  The shared pagination input is generated too.
- Owner and PUBLIC/PRIVATE visibility annotations are checked against the delivery
  envelope. Durable subscription references must resolve to owner-private commands.
  Broker permissions retain the existing public request and owner-private delivery paths.
- Owner adaptors retain semantic UUID, receipt and business-evidence checks.
  Handlers, aggregate behaviour and explicit realtime publication remain owner policy.
- Native architecture checks reject repeated input definitions and permit generated
  plain request imports only within their owning application context. Domain code
  retains its inward dependency boundary.
- Scaffolds, context navigation, architecture guidance and command/query/end-to-end
  educational recipes now start with Protobuf and show generated imports and mappings.

## Executed checks

| Check                                                | Result                                                                                                     |
| ---------------------------------------------------- | ---------------------------------------------------------------------------------------------------------- |
| `pnpm generate:contracts`                            | Passed; generated outputs accompany the schemas                                                            |
| `pnpm test:focused storefront --context ordering`    | Passed                                                                                                     |
| `pnpm test:focused operations --context preparation` | Passed, including 20 focused examples                                                                      |
| `pnpm test:focused engagement --context loyalty`     | Passed, including 11 Loyalty scenarios                                                                     |
| `pnpm check --changed`                               | Passed across all five applications and documentation                                                      |
| `pnpm check:compatibility --against HEAD`            | Passed; published fields and historical private fixtures preserved                                         |
| `pnpm verify`                                        | Passed; isolated generation drift, baseline compatibility, native/race tests and 54 fast Gherkin scenarios |
| `pnpm test:integration`                              | Passed; disposable infrastructure removed after completion                                                 |
| `git diff --check`                                   | Passed                                                                                                     |

Generator fixtures cover schema-added command/query fields, optional presence,
missing annotations, foreign ownership, visibility mismatches, unbound payloads,
unsupported repeated fields, user oneofs and private-subscription authority.
Native parser tests cover absent/present pages, zero values, malformed input,
required fields and numeric bounds. Existing typed registrations and owner use-case
suites execute with the generated inputs.

The live lane exercised migrations, real PostgreSQL/RabbitMQ delivery and broker
permissions, all 31 business API operations, ten live workflow scenarios, twelve
infrastructure scenario reports, provider recovery and browser/session behaviour.
Ordering accepted RecordCollection while its private command consumer was paused;
resuming execution recorded collection and explicitly published the browser update.
Loyalty's paused issuance, command reply loss and identical retries also recovered.

The deterministic workflow binding dry-run intentionally reports ten skipped
scenarios. Those scenarios executed and passed in the live lane. No environment-
dependent verification lane was omitted.

The generator deliberately rejects request forms outside its supported scalar and
PageRequest vocabulary. Extending that vocabulary requires schema/parser fixtures;
all current command and query payloads are supported.
