import type {WriteRepository} from '../../../../foundation/write-repository.js';

import {MappedWriteRepository} from '../../../../adaptors/mapped-write-repository.js';
import {Reward, type RewardState} from '../../domain/reward.js';

/** Restore authoritative aggregates and stage snapshots in the active PostgreSQL transaction. */
export class PostgresRewardWriteRepository extends MappedWriteRepository<RewardState, Reward> {
  constructor(source: WriteRepository<RewardState>) {
    super(source, state => new Reward(state), aggregate => aggregate.snapshot());
  }
}
