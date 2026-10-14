import type { PublishedEventPayloads } from '../adaptors/generated/event-types.js';
import type { PrivateEventPayloads } from '../adaptors/generated/private-event-registry.js';
/** Generated published bindings and owner-declared private payloads compose the graph. */
export type EventPayloads = PublishedEventPayloads & PrivateEventPayloads;
import type { Metadata, Outcome } from '../foundation/application.js';
import type { PostgresContextDatabase } from '../adaptors/persistence/postgres-database.js';
import type { InternalCommandCodec } from '../adaptors/messaging/internal-command-codec.js';
import type { EventSubscription } from '../adaptors/messaging/rabbitmq-consumer.js';

/** An owner-local subscription installs both durable hand-off and command execution. */
export function commandSubscription<K extends keyof EventPayloads, C extends object>(
  event: K,
  codec: InternalCommandCodec<C>,
  definitions: readonly { consumer: string; event: string; command?: string; publication?: string }[],
  target: (event: EventPayloads[K]) => string,
  eventHandler: { handle: (m: Metadata, event: NoInfer<EventPayloads[K]>) => Promise<Outcome> },
  commandHandler: { execute: (m: Metadata, command: NoInfer<C>) => Promise<Outcome> },
): EventSubscription[] {
  const definition = definitions.find((item) => item.consumer === codec.consumer);
  if (
    !definition ||
    definition.publication ||
    definition.command !== codec.command ||
    definition.event !== event
  )
    throw new Error('Subscription manifest disagrees with command binding');
  const shared = {
    owner: codec.owner,
    consumer: codec.consumer,
    event,
    target: (value: object) => target(value as EventPayloads[K]),
  };
  return [
    { ...shared, handle: (m, value) => eventHandler.handle(m, value as EventPayloads[K]) },
    {
      ...shared,
      decodeCommand: (body) => codec.decode(body),
      handle: (m, value) => commandHandler.execute(m, value as C),
    },
  ];
}

/** Resolved context resources; the service root owns worker start/stop ordering. */
export type ContextRuntime = {
  database: PostgresContextDatabase;
  subscriptions: EventSubscription[];
  commandHeader: (body: Buffer) => { id: string; name: string; context: string; correlationId: string };
  dispose(): Promise<void>;
};

/** Reject missing, duplicate, foreign and mismatched bindings before starting workers. */
export function completeSubscriptions(
  owner: string,
  definitions: readonly { consumer: string; event: string; command?: string; publication?: string }[],
  workers: EventSubscription[],
): EventSubscription[] {
  const expected = new Set(definitions.map((d) => d.consumer));
  if (expected.size !== definitions.length) throw Error('Duplicate subscription declaration');
  for (const definition of definitions) {
    if (Boolean(definition.command) === Boolean(definition.publication))
      throw Error('Subscription requires exactly one reaction mode');
    const actual = workers.filter((w) => w.consumer === definition.consumer);
    if (
      actual.length !== (definition.command ? 2 : 1) ||
      actual.filter((w) => Boolean(w.decodeCommand)).length !== (definition.command ? 1 : 0) ||
      actual.some((w) => w.owner !== owner || w.event !== definition.event)
    )
      throw Error('Incomplete subscription: ' + definition.consumer);
  }
  if (workers.some((w) => !expected.has(w.consumer))) throw Error('Undeclared subscription');
  return workers;
}

/** Bind one independently acknowledged private-event publication subscription. */
export function publicationSubscription<K extends keyof EventPayloads>(
  owner: string,
  consumer: string,
  event: K,
  publication: string,
  definitions: readonly { consumer: string; event: string; command?: string; publication?: string }[],
  handler: { handle(m: Metadata, event: NoInfer<EventPayloads[K]>): Promise<Outcome> },
): EventSubscription {
  const definition = definitions.find((d) => d.consumer === consumer);
  if (
    !definition ||
    definition.command ||
    definition.event !== event ||
    definition.publication !== publication ||
    !consumer.startsWith(owner + '.')
  )
    throw new Error('Subscription manifest disagrees with publication binding');
  return {
    owner,
    consumer,
    event,
    sourceTarget: true,
    target: () => '',
    handle: (m, value) => handler.handle(m, value as EventPayloads[K]),
  };
}
