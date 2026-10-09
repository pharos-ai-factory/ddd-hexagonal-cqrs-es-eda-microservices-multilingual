import type {PageRequest} from '../../../../foundation/pagination.js';
import type {RewardReadRepository} from '../ports/reward-read-repository.js';

/** Select rewards or an explicitly requested page. */
export type ListRewardsQuery = Readonly<{page?: PageRequest | undefined}>;

/** Execute the ListRewards use case through its context-owned read port. */
export class ListRewardsQueryHandler {
  constructor(private readRepository: RewardReadRepository) {}
  async execute(query: ListRewardsQuery) { return query.page ? this.readRepository.page(query.page) : {items: await this.readRepository.list()}; }
}
