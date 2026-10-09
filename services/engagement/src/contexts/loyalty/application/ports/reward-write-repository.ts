import type {WriteRepository} from '../../../../foundation/write-repository.js';

import type {Reward} from '../../domain/reward.js';

/** Authoritative reward aggregate persistence inside its command transaction. */
export type RewardWriteRepository = WriteRepository<Reward>;
