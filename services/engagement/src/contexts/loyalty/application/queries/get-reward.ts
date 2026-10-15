import type { GetRewardQuery } from '../../../../contracts/requests/generated/loyalty.js';
import type { GetRewardResponse } from '../../../../contracts/requests/generated/loyalty-responses.js';
import type { RewardReadRepository } from '../ports/reward-read-repository.js';
/** Preserve revision, missing-resource and pagination policy in a generated response. */
export class GetRewardQueryHandler {
  constructor(private readRepository: RewardReadRepository) {}
  async execute(query: GetRewardQuery): Promise<GetRewardResponse> {
    const value = await this.readRepository.get(query.id);
    return value ? { reward: value } : { error: { code: 'not_found', message: '' } };
  }
}
