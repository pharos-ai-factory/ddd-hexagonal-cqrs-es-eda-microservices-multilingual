# Contract enforcement improvements

The six agreed improvements are tracked here. Published sources retain the
context-first layout; domain facts and private delivery formats remain owner-local.

| Issue | Intended result | Status |
| --- | --- | --- |
| 1. Explicit required input fields | Protobuf contracts distinguish mandatory inputs from future optional additions; validators preserve explicit zero values | Complete |
| 2. Typed boundary translation | API and owner adaptors use explicit mappings into plain application types; business payload mapping has no reflection or JSON bridge | Complete |
| 3. Independent command/query delivery | Each owner has separate command and query queues and consumers | Complete |
| 4. Durable command replies | Exact command reply bytes commit with the aggregate, receipt, outcome and outgoing intent; recovery publishes committed replies | Complete |
| 5. Context-owned request packages | Each context owns its Protobuf request/reply envelopes and package; shared messages contain technical metadata only | Complete |
| 6. Frontend OpenAPI enforcement | Generated operation/request/response types are used by frontend calls; contract drift and incompatible usage fail verification | Complete |

Verification must cover omitted optional inputs, zero values, complete operation
mapping, queue isolation, commit/reply-loss recovery and a frontend compilation
failure after an incompatible specification change. `pnpm verify` and
`pnpm test:integration` remain the hand-off gates. Generated-file drift checks
must detect changed, missing and newly generated files.


Issue 6 records the frontend enforcement gap added to the original five items.
Review follow-ups strengthen these guarantees: native owner envelopes allow
context-local payload tags and names; API mappings consume compiler descriptors;
frontend inputs include required headers/query objects and shared callers; and
all owner languages reclaim abandoned reply leases before response expiry.
All six implementations and their targeted checks are complete. Accepted behaviour
is described in [decision 0009](decisions/0009-contract-enforcement-and-command-replies.md).

`pnpm verify` and `pnpm test:integration` passed after the review fixes.
Coverage includes compiler-driven schema evolution, frontend required-parameter
mutations, real abandoned-lease recovery in all owner languages, acknowledged
commands with pending reply publication, exact stored-byte recovery and browser
journeys. The isolated session-revocation fixture now passes on Docker Desktop
for macOS with published Centrifugo ports and container-to-host addressing.
See [executed verification](verification.md) for the full result and local logs.

Use `pnpm generate:contracts` after Protobuf changes. Use `pnpm generate:http`
after OpenAPI changes. `pnpm check:contracts` and `pnpm verify` reject stale
outputs and incompatible frontend/API usage. Generated bindings, descriptors,
HTTP types and API mappings live in explicitly classified `generated/` folders.
