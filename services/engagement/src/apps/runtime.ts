import type {OrderCollected, PickupOpened, RewardIssued} from '../contracts/events.js';
import type {RewardEarned} from '../contexts/loyalty/application/events.js';
import type {NotificationRequested} from '../contexts/communication/application/events.js';

/** Event names determine their application payload at the composition boundary. */
export type EventPayloads = {
  'collection.order-collected': OrderCollected;
  'collection.pickup-opened': PickupOpened;
  'loyalty.reward-issued': RewardIssued;
  'loyalty.reward-earned': RewardEarned;
  'communication.notification-requested': NotificationRequested;
};
import type {Metadata, Outcome} from '../foundation/application.js';
import type {PostgresContextDatabase} from '../adaptors/postgres.js';
import type {InternalCommandCodec} from '../adaptors/internal-commands.js';
import type {EventSubscription} from '../adaptors/broker.js';

/** An owner-local subscription installs both durable hand-off and command execution. */
export function commandSubscription<K extends keyof EventPayloads, C extends object>(
  event: K,
  codec: InternalCommandCodec<C>, definitions: readonly {consumer: string; event: string; command: string}[], target: (event: EventPayloads[K]) => string,
  eventHandler: {handle: (m: Metadata, event: NoInfer<EventPayloads[K]>) => Promise<Outcome>},
  commandHandler: {execute: (m: Metadata, command: NoInfer<C>) => Promise<Outcome>},
): EventSubscription[] {
  const definition = definitions.find(item => item.consumer === codec.consumer);
  if (!definition || definition.command !== codec.command || definition.event !== event) throw new Error('Subscription manifest disagrees with command binding');
  const shared = {owner: codec.owner, consumer: codec.consumer, event, target: (value: object) => target(value as EventPayloads[K])};
  return [
    {...shared, handle: (m, value) => eventHandler.handle(m, value as EventPayloads[K])},
    {...shared, decodeCommand: body => codec.decode(body), handle: (m, value) => commandHandler.execute(m, value as C)},
  ];
}

/** Resolved context resources; the service root owns worker start/stop ordering. */
export type ContextRuntime = {
  database: PostgresContextDatabase;
  subscriptions: EventSubscription[];
  commandHeader: (body: Buffer) => {id: string; name: string; context: string; correlationId: string};
  dispose(): Promise<void>;
};

/** Reject missing, duplicate, foreign and mismatched bindings before starting workers. */
export function completeSubscriptions(owner: string, definitions: readonly {consumer: string; event: string; command: string}[], workers: EventSubscription[]): EventSubscription[] {
  const expected = new Set(definitions.map(d => d.consumer));
  if (expected.size !== definitions.length) throw Error('Duplicate subscription declaration');
  for (const definition of definitions) {
    const actual = workers.filter(w => w.consumer === definition.consumer);
    if (actual.length !== 2 || actual.filter(w => Boolean(w.decodeCommand)).length !== 1 ||
      actual.some(w => w.owner !== owner || w.event !== definition.event)) throw Error('Incomplete subscription: '+definition.consumer);
  }
  if (workers.some(w => !expected.has(w.consumer))) throw Error('Undeclared subscription');
  return workers;
}
