import type {Loaded, Metadata, Outcome, Publication} from '../foundation/application.js';
import type {CommandContext, CommandResult, WriteRepository} from '../foundation/write-repository.js';

/** Internal commit material; feature handlers return only their business result. */
export type TransactionResult = {status: string; publications?: Publication[]};
/** Infrastructure boundary for one aggregate in one context database. */
export interface AggregateTransaction<A> {
  execute(metadata: Metadata, work: (repository: WriteRepository<A>) => Promise<TransactionResult>): Promise<Outcome>;
}
/** Transport-facing command execution includes durable outcomes and receipts. */
export interface CommandExecutor<C> { execute(metadata: Metadata, command: C): Promise<Outcome>; }
/** A feature handler receives its own repository for each invocation. */
export interface FeatureCommandHandler<C> {
  execute(context: CommandContext, command: C): Promise<CommandResult>;
}
/** Centralise transaction policy and construct a fresh feature handler for each command. */
export function bindCommand<A, C>(transaction: AggregateTransaction<A>, handler: (repository: WriteRepository<A>) => FeatureCommandHandler<C>,
  metadata: (value: Metadata, command: C) => Metadata = value => value): CommandExecutor<C> {
  return {execute: (value, command) => {
    const m = metadata(value, command);
    return transaction.execute(m, async repository => handler(repository).execute({target: m.target}, command));
  }};
}
/** Stage outgoing messages from the saved aggregate's domain facts. */
export class EventRecordingWriteRepository<A> implements WriteRepository<A> {
  publications: Publication[] = [];
  constructor(private source: WriteRepository<A>, private map: (aggregate: A) => Publication[]) {}
  get(id: string): Promise<Loaded<A> | undefined> { return this.source.get(id); }
  async save(aggregate: A) {
    await this.source.save(aggregate);
    this.publications = this.map(aggregate);
  }
}
