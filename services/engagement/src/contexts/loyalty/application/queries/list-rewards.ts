import type { ListRewardsQuery } from '../../../../contracts/requests/generated/loyalty.js';
import type { ListRewardsResponse } from '../../../../contracts/requests/generated/loyalty-responses.js';
import type { RewardReadRepository } from '../ports/reward-read-repository.js';
/** Preserve revision, missing-resource and pagination policy in a generated response. */
export class ListRewardsQueryHandler {
  constructor(private readRepository: RewardReadRepository) {}
  async execute(query: ListRewardsQuery): Promise<ListRewardsResponse> {
    const page = query.page
      ? await this.readRepository.page(query.page)
      : { items: await this.readRepository.list() };
    return {
      rewards: {
        items: page.items,
        paged: query.page !== undefined,
        ...(page.nextId !== undefined ? { nextId: page.nextId } : {}),
      },
    };
  }
}
