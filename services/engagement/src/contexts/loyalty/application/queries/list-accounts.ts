import type {PageRequest} from '../../../../foundation/pagination.js';
import type {AccountReader} from '../ports/account-reader.js';

/** Select accounts or an explicitly requested page. */
export type ListAccountsQuery = Readonly<{page?: PageRequest | undefined}>;

/** Execute the ListAccounts use case through its context-owned read port. */
export class ListAccountsQueryHandler {
  constructor(private reader: AccountReader) {}
  async execute(query: ListAccountsQuery) { return query.page ? this.reader.page(query.page) : {items: await this.reader.list()}; }
}
