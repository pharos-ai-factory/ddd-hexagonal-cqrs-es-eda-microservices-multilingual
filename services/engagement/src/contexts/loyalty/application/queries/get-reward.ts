import type {RewardReadRepository} from '../ports/reward-read-repository.js';

/** Select one reward. */
export type GetRewardQuery = Readonly<{id: string}>;

/** Execute the GetReward use case through its context-owned read port. */
export class GetRewardQueryHandler {
  constructor(private readRepository: RewardReadRepository) {}
  async execute(query: GetRewardQuery) { return this.readRepository.get(query.id); }
}
