import type {RewardReader} from '../ports/reward-reader.js';

/** Select one reward. */
export type GetRewardQuery = Readonly<{id: string}>;

/** Execute the GetReward use case through its context-owned read port. */
export class GetRewardQueryHandler {
  constructor(private reader: RewardReader) {}
  async execute(query: GetRewardQuery) { return this.reader.get(query.id); }
}
