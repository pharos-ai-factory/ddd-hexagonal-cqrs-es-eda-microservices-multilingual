import type {Page, PageRequest} from '../../../foundation/pagination.js';
import type {RewardView} from '../application/read-models/reward.js';
import type {GetRewardQueryHandler} from '../application/queries/get-reward.js';
import type {ListRewardsQueryHandler} from '../application/queries/list-rewards.js';

/** Translate HTTP and RabbitMQ query arguments into named loyalty use cases. */
export function rewardQueryEndpoints(get: GetRewardQueryHandler, list: ListRewardsQueryHandler) {
  return {
    get: (id: string) => get.execute({id}),
    list: async () => (await list.execute({})).items,
    page: async (page: PageRequest): Promise<Page<RewardView>> => list.execute({page}),
  };
}
