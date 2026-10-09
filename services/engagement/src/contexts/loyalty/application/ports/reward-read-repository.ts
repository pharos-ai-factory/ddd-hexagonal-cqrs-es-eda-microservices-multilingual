import type {PagedQueryPort} from '../../../../foundation/pagination.js';
import type {RewardView} from '../read-models/reward.js';

/** Read rewards and their stable revisions through the loyalty boundary. */
export interface RewardReadRepository extends PagedQueryPort<RewardView> {}
