import type { ListAccountsQuery } from '../../../../contracts/requests/generated/loyalty.js';
import type { ListAccountsResponse } from '../../../../contracts/requests/generated/loyalty-responses.js';
import type { AccountReadRepository } from '../ports/account-read-repository.js';
/** Preserve revision, missing-resource and pagination policy in a generated response. */
export class ListAccountsQueryHandler {
  constructor(private readRepository: AccountReadRepository) {}
  async execute(query: ListAccountsQuery): Promise<ListAccountsResponse> {
    const page = query.page
      ? await this.readRepository.page(query.page)
      : { items: await this.readRepository.list() };
    return {
      accounts: {
        items: page.items,
        paged: query.page !== undefined,
        ...(page.nextId !== undefined ? { nextId: page.nextId } : {}),
      },
    };
  }
}
