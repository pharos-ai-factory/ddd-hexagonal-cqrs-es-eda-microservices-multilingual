import type {Metadata, Outcome} from '../foundation/application.js';
import type {PostgresContextDatabase} from '../adaptors/postgres.js';
import type {InternalCommandCodec} from '../adaptors/internal-commands.js';
import type {EventSubscription} from '../adaptors/broker.js';

/** An owner-local subscription installs both durable hand-off and command execution. */
export function commandSubscription<E extends object, C extends object>(
  codec: InternalCommandCodec<C>, definitions: readonly {consumer: string; event: string; command: string}[], target: (event: E) => string,
  eventHandler: {handle(m: Metadata, event: E): Promise<Outcome>},
  commandHandler: {execute(m: Metadata, command: C): Promise<Outcome>},
): EventSubscription[] {
  const definition = definitions.find(item => item.consumer === codec.consumer);
  if (!definition || definition.command !== codec.command) throw new Error('Subscription manifest disagrees with command binding');
  const event = definition.event;
  const shared = {owner: codec.owner, consumer: codec.consumer, event, target: (value: object) => target(value as E)};
  return [
    {...shared, handle: (m, value) => eventHandler.handle(m, value as E)},
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
