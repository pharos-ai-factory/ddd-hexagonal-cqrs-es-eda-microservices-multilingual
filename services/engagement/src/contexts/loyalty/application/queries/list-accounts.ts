import type { ListAccountsQuery } from '../../../../contracts/requests/generated/loyalty.js';
import type { AccountReadRepository } from '../ports/account-read-repository.js';

/** Execute the ListAccounts use case through its context-owned read port. */
export class ListAccountsQueryHandler {
  constructor(private readRepository: AccountReadRepository) {}
  async execute(query: ListAccountsQuery) {
    return query.page ? this.readRepository.page(query.page) : { items: await this.readRepository.list() };
  }
}
