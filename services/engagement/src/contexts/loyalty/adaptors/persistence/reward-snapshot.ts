import {storedObject} from '../../../../adaptors/stored-values.js';
import {Reward, type RewardState} from '../../domain/reward.js';

/** Restore the owning aggregate before exposing stored authority. */
export const restoreRewardSnapshot = (value: unknown): RewardState => new Reward(storedObject(value) as RewardState).snapshot();
