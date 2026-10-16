import { Rejection } from '../foundation/domain.js';
import { RealtimeBuffer } from './realtime/publication-buffer.js';
import catalogue from './generated/catalogue.json' with { type: 'json' };
import type {
  Loaded,
  Metadata,
  Outcome,
  Publication,
  RealtimeBusPort,
  RealtimeEvent,
  DomainEventBusPort,
} from '../foundation/application.js';
import type { CommandContext, CommandResult, WriteRepository } from '../foundation/write-repository.js';

/** Internal commit material; feature handlers return only their business result. */
export type TransactionResult = {
  status: string;
  publications?: Publication[];
  realtime?: RealtimeEvent[];
  rejection?: { code: string; message: string };
};
/** Infrastructure boundary for one aggregate in one context database. */
export interface AggregateTransaction<A> {
  execute(
    metadata: Metadata,
    work: (repository: WriteRepository<A>) => Promise<TransactionResult>,
  ): Promise<Outcome>;
}
/** Transport-facing command execution includes durable outcomes and receipts. */
export interface CommandExecutor<C> {
  execute(metadata: Metadata, command: C): Promise<Outcome>;
}
/** A feature handler receives its own repository for each invocation. */
export interface FeatureCommandHandler<C> {
  execute(context: CommandContext, command: C): Promise<CommandResult>;
}
/** Centralise transaction policy and construct a fresh feature handler for each command. */
export function bindCommand<A, C>(
  transaction: AggregateTransaction<A>,
  handler: (
    repository: WriteRepository<A>,
    realtime: RealtimeBusPort<RealtimeEvent>,
  ) => FeatureCommandHandler<C>,
  metadata: (value: Metadata, command: C) => Metadata = (value) => value,
): CommandExecutor<C> {
  return {
    execute: (value, command) => {
      const m = metadata(value, command);
      return transaction.execute(m, async (repository) => {
        const bus = new RealtimeBuffer();
        try {
          const result = await handler(repository, bus).execute(
            { target: m.target, ...(m.actorId ? { actorId: m.actorId } : {}) },
            command,
          );
          return { ...result, realtime: bus.events };
        } catch (error) {
          if (!(error instanceof Rejection)) throw error;
          return { status: '', rejection: error.outcome(), realtime: bus.events };
        } finally {
          bus.close();
        }
      });
    },
  };
}
/** Stage outgoing messages from the saved aggregate's domain facts. */
export class EventRecordingWriteRepository<A> implements WriteRepository<A>, DomainEventBusPort<Publication> {
  publications: Publication[] = [];
  constructor(
    private source: WriteRepository<A>,
    private map: (aggregate: A) => Publication[],
  ) {}
  get(id: string): Promise<Loaded<A> | undefined> {
    return this.source.get(id);
  }
  async save(aggregate: A) {
    const publications = this.map(aggregate);
    this.publications = [];
    for (const publication of publications) await this.publish(publication);
    await this.source.save(aggregate);
  }

  // eslint-disable-next-line @typescript-eslint/require-await -- Stage through the asynchronous domain bus port.
  async publish(event: Publication) {
    if (!catalogue.some((e) => e.name === event.name && e.visibility === 'domain'))
      throw new Error('Aggregate saves may stage only private domain events');
    this.publications.push(event);
  }
}
