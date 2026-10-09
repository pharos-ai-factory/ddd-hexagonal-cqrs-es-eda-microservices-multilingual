import type {AccountReadRepository} from '../ports/account-read-repository.js';

/** Select one account. */
export type GetAccountQuery = Readonly<{id: string}>;

/** Execute the GetAccount use case through its context-owned read port. */
export class GetAccountQueryHandler {
  constructor(private readRepository: AccountReadRepository) {}
  async execute(query: GetAccountQuery) { return this.readRepository.get(query.id); }
}
