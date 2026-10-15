# 0030 — Generated package layout

Accepted: 2026-10-15. Refines [0017](0017-generated-event-boundaries.md),
[0028](0028-schema-generated-command-and-query-inputs.md) and
[0029](0029-schema-generated-command-and-query-responses.md).

## Decision

Wholly generated packages live beneath a directory named `generated`. Their
filenames describe their contents without an additional `generated` prefix or
suffix. Generated source retains its generator header. Compiler conventions such
as `.pb.go`, `_pb2.py` and `_pb2.pyi` remain unchanged.

Go Storefront uses these locations relative to its service root:

| Contents                                 | Location                                                                       |
| ---------------------------------------- | ------------------------------------------------------------------------------ |
| Plain command/query inputs and responses | `contracts/requests/generated/model/<owner>/{payloads,responses}.go`           |
| Plain integration event DTOs             | `contracts/events/generated/model/<producer>.go`                               |
| Event and subscription identifiers       | `contracts/events/generated/names/names.go`                                    |
| Shared response encoders                 | `adaptors/messaging/generated/responses/responses.go`                          |
| Typed owner query registrations          | `contexts/<owner>/adaptors/messaging/generated/queryregistration/responses.go` |
| Protobuf wire bindings                   | `contracts/{events,requests,realtime}/generated/cafe/…`                        |

The `model` package boundary distinguishes plain DTOs from Protobuf bindings.
Application handlers may import their owner's request models, shared response
models and service-local event DTOs. Domain packages remain independent of these
contracts. Generated encoders and registrations remain adaptor responsibilities.

Mixed packages retain `_generated.go` filenames beside their handwritten source
or owner declarations. These include `contracts/events/protobuf`,
`apps/support` and context messaging roots. Keeping codec registrations in their
existing packages preserves access to package-private helpers. Python and
TypeScript already follow the generated-directory convention.

## Generation and checks

Generators own separate output sets. Request wire generation prunes only
`generated/cafe/requests`; plain model generation owns `generated/model`.
Event model generation removes obsolete producer files from its own directory.
Run `pnpm generate:contracts` after schema or generator changes and commit the
outputs together. Regeneration has no effect on wire identities or stored bytes.

Architecture checks classify the relocated outputs, reject redundant filename
markers beneath generated folders and retain narrow application import rules.
Generator tests verify model preservation, stale binding removal, generated
locations and preservation of handwritten codecs. Documentation and scaffolds
use the same import paths as executable code.
