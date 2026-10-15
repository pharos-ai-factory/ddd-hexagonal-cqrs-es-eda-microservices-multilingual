import type { GetAccountQuery } from '../../../../contracts/requests/generated/loyalty.js';
import type { AccountReadRepository } from '../ports/account-read-repository.js';

/** Execute the GetAccount use case through its context-owned read port. */
export class GetAccountQueryHandler {
  constructor(private readRepository: AccountReadRepository) {}
  async execute(query: GetAccountQuery) {
    return this.readRepository.get(query.id);
  }
}
