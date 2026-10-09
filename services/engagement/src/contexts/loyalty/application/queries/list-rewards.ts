import type {PageRequest} from '../../../../foundation/pagination.js';
import type {RewardReader} from '../ports/reward-reader.js';

/** Select rewards or an explicitly requested page. */
export type ListRewardsQuery = Readonly<{page?: PageRequest | undefined}>;

/** Execute the ListRewards use case through its context-owned read port. */
export class ListRewardsQueryHandler {
  constructor(private reader: RewardReader) {}
  async execute(query: ListRewardsQuery) { return query.page ? this.reader.page(query.page) : {items: await this.reader.list()}; }
}
