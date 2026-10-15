import type { EventName } from '../adaptors/generated/message-names.js';
import type { CommandName } from '../adaptors/generated/request-names.js';
import {
  bindCommand,
  type AggregateTransaction,
  type FeatureCommandHandler,
} from '../adaptors/command-execution.js';
import type { InternalCommandCodec } from '../adaptors/messaging/internal-command-codec.js';
import type { RabbitMQRequestRegistry } from '../adaptors/messaging/rabbitmq-request-registry.js';
import type { Metadata, Outcome, RealtimeBusPort, RealtimeEvent } from '../foundation/application.js';
import type { WriteRepository } from '../foundation/write-repository.js';
import { commandSubscription, type EventPayloads } from './runtime.js';

type Definitions = readonly { consumer: string; event: string; command?: string; publication?: string }[];
type Factory<A, C> = (
  repository: WriteRepository<A>,
  realtime: RealtimeBusPort<RealtimeEvent>,
) => FeatureCommandHandler<C>;
type Reaction<K extends EventName & keyof EventPayloads> = {
  handle(metadata: Metadata, event: NoInfer<EventPayloads[K]>): Promise<Outcome>;
};

/** Supply one aggregate's execution policy to direct command-to-handler registrations. */
export class CommandRegistration<A> {
  constructor(private readonly transaction: AggregateTransaction<A>) {}

  request<C extends object>(
    registry: RabbitMQRequestRegistry,
    name: CommandName,
    parse: (value: unknown) => C,
    handler: Factory<A, NoInfer<C>>,
  ) {
    const executor = bindCommand(this.transaction, handler);
    registry.command(name, parse, (metadata, command) => executor.execute(metadata, command));
  }

  subscription<K extends EventName & keyof EventPayloads, C extends object>(
    event: K,
    codec: InternalCommandCodec<C>,
    definitions: Definitions,
    target: (event: EventPayloads[K]) => string,
    reaction: Reaction<K>,
    handler: Factory<A, NoInfer<C>>,
    metadata?: (value: Metadata, command: NoInfer<C>) => Metadata,
  ) {
    return commandSubscription(
      event,
      codec,
      definitions,
      target,
      reaction,
      bindCommand(this.transaction, handler, metadata),
    );
  }

  /** Obtain external evidence before acquiring the aggregate transaction and repository. */
  preparedSubscription<K extends EventName & keyof EventPayloads, C extends object, P>(
    event: K,
    codec: InternalCommandCodec<C>,
    definitions: Definitions,
    target: (event: EventPayloads[K]) => string,
    reaction: Reaction<K>,
    prepare: (command: NoInfer<C>) => Promise<P>,
    handler: (
      repository: WriteRepository<A>,
      prepared: P,
      realtime: RealtimeBusPort<RealtimeEvent>,
    ) => FeatureCommandHandler<NoInfer<C>>,
  ) {
    return commandSubscription(event, codec, definitions, target, reaction, {
      execute: async (metadata, command) => {
        const prepared = await prepare(command);
        return bindCommand(this.transaction, (repository, realtime) =>
          handler(repository, prepared, realtime),
        ).execute(metadata, command);
      },
    });
  }
}
