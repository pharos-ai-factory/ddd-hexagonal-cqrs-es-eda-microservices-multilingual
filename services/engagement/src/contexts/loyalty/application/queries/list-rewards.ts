import type { ListRewardsQuery } from '../../../../contracts/requests/generated/loyalty.js';
import type { RewardReadRepository } from '../ports/reward-read-repository.js';

/** Execute the ListRewards use case through its context-owned read port. */
export class ListRewardsQueryHandler {
  constructor(private readRepository: RewardReadRepository) {}
  async execute(query: ListRewardsQuery) {
    return query.page ? this.readRepository.page(query.page) : { items: await this.readRepository.list() };
  }
}
