import type {PostgresContextDatabase} from '../../../../adaptors/postgres.js';
import {PostgresAggregateQueries} from '../../../../adaptors/postgres.js';
import {MappedQueries} from '../../../../adaptors/mapped-queries.js';
import {storedObject} from '../../../../adaptors/stored-values.js';
import {Reward, type RewardState} from '../../domain/reward.js';
import type {RewardView} from '../../application/read-models/reward.js';

/** Restore the owning aggregate before exposing stored authority. */
export const restoreReward = (value: unknown): RewardState => new Reward(storedObject(value) as RewardState).snapshot();
/** Select the stable application read fields independently from storage. */
export const rewardView = (s: RewardState): RewardView => ({id: s.id, grantId: s.grantId, customerId: s.customerId, benefit: s.benefit, status: s.status, expiresAt: s.expiresAt,
    ...(s.redeemedFor !== undefined ? {redeemedFor: s.redeemedFor} : {})});

/** PostgreSQL implementation of the loyalty RewardReader capability. */
export class PostgresRewardReader extends MappedQueries<RewardState, RewardView> {
  constructor(database: PostgresContextDatabase) {
    super(new PostgresAggregateQueries(database, 'reward', restoreReward), rewardView);
  }
}
