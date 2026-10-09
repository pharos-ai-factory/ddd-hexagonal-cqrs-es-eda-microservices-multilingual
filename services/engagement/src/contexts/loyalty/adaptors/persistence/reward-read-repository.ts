import type {PostgresContextDatabase} from '../../../../adaptors/postgres.js';
import {PostgresSnapshotReadRepository} from '../../../../adaptors/snapshot-read-repository.js';
import {MappedReadRepository} from '../../../../adaptors/mapped-read-repository.js';
import {restoreRewardSnapshot} from './reward-snapshot.js';
import type {RewardState} from '../../domain/reward.js';
import type {RewardView} from '../../application/read-models/reward.js';

/** Select the stable application read fields independently from storage. */
export const rewardView = (s: RewardState): RewardView => ({id: s.id, grantId: s.grantId, customerId: s.customerId, benefit: s.benefit, status: s.status, expiresAt: s.expiresAt,
    ...(s.redeemedFor !== undefined ? {redeemedFor: s.redeemedFor} : {})});

/** PostgreSQL implementation of the loyalty RewardReadRepository capability. */
export class PostgresRewardReadRepository extends MappedReadRepository<RewardState, RewardView> {
  constructor(database: PostgresContextDatabase) {
    super(new PostgresSnapshotReadRepository(database, 'reward', restoreRewardSnapshot), rewardView);
  }
}
