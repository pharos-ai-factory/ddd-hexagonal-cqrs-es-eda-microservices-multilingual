# RabbitMQ authority

Each context uses `cafe_<owner>` through `<OWNER>_BROKER_URL`. The API uses
`cafe_api` through `API_BROKER_URL`. The one-shot topology process uses
`cafe_topology` through `TOPOLOGY_BROKER_URL`. These users have no management tags.
The development launcher uses the administrator to provision their credentials
and permissions in the `reference` vhost. The topology container receives only
its dedicated URL secret.

## Catalogue-generated resources

`pnpm generate:contracts` generates
`services/storefront/apps/topology/permissions.generated.json` from owner
`subscriptions.json`, published/private event catalogues and the API request
contract selection in `contracts/generation.json`. `scripts/broker_permissions.py`
expands the queue families below and emits escaped, anchored allow-lists. The
launcher applies that generated policy before running topology. Generated drift
is checked alongside the wire contracts.

| Purpose                  | Queue names                                                 |
| ------------------------ | ----------------------------------------------------------- |
| Event subscription       | `ref.<owner>.<consumer>`, plus `.retry` and `.dead`         |
| Durable command reaction | `ref.<owner>.<consumer>.command`, plus `.retry` and `.dead` |
| API commands             | `ref.<owner>.commands`, plus `.dead`                        |
| API queries              | `ref.<owner>.queries`, plus `.dead`                         |
| API replies              | `ref.api.<owner>.replies`                                   |

Here `<owner>.<consumer>` is the complete subscription identity. Private and
integration subscriptions have independent queues. A subscription with a durable
command creates both event and command queue families. Generation rejects
resource collisions, including a queue named `ref.<owner>.delivery`, and wildcard
consumer identities. Removed consumers lose runtime access when permissions are
reapplied. Existing retired queues require explicit administrative inspection and
cleanup; generation preserves their queued evidence.

## Resource permissions

RabbitMQ checks resource names independently for configure, write and read.
Queue binding requires write on the queue and read on the source exchange;
exchange binding requires write on the destination and read on the source.
Declaring a queue with a dead-letter exchange also requires read on that queue
and write on its dead-letter exchange. See the
[RabbitMQ 4.1 access-control reference](https://www.rabbitmq.com/docs/4.1/access-control#authorisation).

| User     | Configure                      | Write                                                                    | Read                                                                |
| -------- | ------------------------------ | ------------------------------------------------------------------------ | ------------------------------------------------------------------- |
| Context  | None (`^$`)                    | `cafe.events`, `cafe.replies`, its exact `ref.<owner>.delivery` exchange | Its generated event, command, retry, dead-letter and request queues |
| API      | None (`^$`)                    | `cafe.requests`                                                          | Generated API reply queues                                          |
| Topology | Generated queues and exchanges | Generated queues and owner delivery exchanges                            | Generated exchanges and queues declared with dead-letter arguments  |

Context queue permissions enumerate the catalogue; an owner prefix alone grants
no access to additional queues. Runtime users have no queue write or exchange
read permission, so they cannot bind or unbind queues or exchanges. The default
exchange (`amq.default`) is outside every generated write allow-list. Runtime
users cannot declare or delete resources.

The topology credential can consume the retry and request queues for which
RabbitMQ requires read permission during declaration, and can publish to delivery
exchanges because dead-letter declaration requires write permission there. It is
a trusted, short-lived bootstrap credential. It cannot publish to the three
shared topic exchanges or configure resources outside the generated catalogue.

## Topics and private events

The wire prefix for private facts is `domain`; integration publications use
`integration`. Owner event names already start with their owner, producing keys
such as `domain.loyalty.reward-earned` and `integration.ordering.order-placed`.

| User          | Topic write                                                                                 | Topic read (binding keys)                                                             |
| ------------- | ------------------------------------------------------------------------------------------- | ------------------------------------------------------------------------------------- |
| Context `<o>` | `cafe.events`: `^(domain\|integration)\.<o>\..+$`; `cafe.replies`: exact `reply.<o>`        | None                                                                                  |
| API           | `cafe.requests`: exact generated `request.<owner>.command` and `request.<owner>.query` keys | None                                                                                  |
| Topology      | None                                                                                        | Exact catalogue event keys, request keys and reply keys on their respective exchanges |

Every managed user receives explicit topic rules on all three topic exchanges.
Unneeded rules use `^$`; omitting a topic rule would leave the additional topic
restriction unapplied. Topic read controls bindings, while queue read controls
consumption.

Only topology binds runtime queues. Catalogue validation accepts integration
subscriptions across owners and rejects private subscriptions across owners.
Resource and topic permissions cannot express a queue/routing-key pair constraint
for this single topology user. The checked catalogue and trusted bootstrap enforce
that pairing; runtime credentials cannot change it. A shared service process holds
the credentials of its own contexts, so those contexts share that process's trust
boundary.

## Development workflow and evidence

After changing a subscription or request contract selection, regenerate, review
both generated topology and permission files, then run `pnpm dev:up`. The launcher
preserves existing credentials, generates a topology password if absent, stops the
managed API/owner processes, reapplies permissions, recreates topology and starts
the runtimes. This replaces their cached broker authorisation. Restart any
separately launched host workers as well.

The integration lane stops API/owner workers while exercising declared queues.
It proves owner publication and consumption, private command/retry/dead-letter
access, all-owner API commands/queries/replies, topology declarations and denied
foreign access, forged publications, binding changes, default-exchange bypasses
and unregistered resources. Go, Python and TypeScript recovery fixtures retain
their owner credentials and use the declared queues. The full live journey then
runs through the same generated topology.

## Command and query visibility

Payload schemas declare their owner and PUBLIC/PRIVATE visibility. The generator
checks those declarations against public request and private durable-command
envelopes before producing permissions. The generated `request_contracts` inventory
records each payload's role and visibility; private subscription command references
must resolve to an owner-private schema. API contract selection authorises the
published owner request families. Public request keys remain scoped to owner and
command/query kind, while private commands use owner-only delivery exchanges and
queues. Payload-level validation remains in the owner boundary. Changing visibility
requires the corresponding envelope, caller selection and compatibility review.
