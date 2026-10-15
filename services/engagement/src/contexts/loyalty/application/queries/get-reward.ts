import type { GetRewardQuery } from '../../../../contracts/requests/generated/loyalty.js';
import type { RewardReadRepository } from '../ports/reward-read-repository.js';

/** Execute the GetReward use case through its context-owned read port. */
export class GetRewardQueryHandler {
  constructor(private readRepository: RewardReadRepository) {}
  async execute(query: GetRewardQuery) {
    return this.readRepository.get(query.id);
  }
}
