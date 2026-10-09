import type {Loaded} from './application.js';

/** Business status returned by a feature command handler. */
export type CommandResult = {status: string};

/** Load and save the targeted aggregate within the active command transaction. */
export interface WriteRepository<A> {
  get(id: string): Promise<Loaded<A> | undefined>;
  save(aggregate: A): Promise<void>;
}

/** The aggregate targeted by a business operation. */
export type CommandContext = {target: string};
