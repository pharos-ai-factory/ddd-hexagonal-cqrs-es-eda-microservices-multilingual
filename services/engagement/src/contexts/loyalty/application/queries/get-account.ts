import type {AccountReader} from '../ports/account-reader.js';

/** Select one account. */
export type GetAccountQuery = Readonly<{id: string}>;

/** Execute the GetAccount use case through its context-owned read port. */
export class GetAccountQueryHandler {
  constructor(private reader: AccountReader) {}
  async execute(query: GetAccountQuery) { return this.reader.get(query.id); }
}
