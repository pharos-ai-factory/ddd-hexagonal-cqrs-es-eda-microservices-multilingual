import type { GetAccountQuery } from '../../../../contracts/requests/generated/loyalty.js';
import type { GetAccountResponse } from '../../../../contracts/requests/generated/loyalty-responses.js';
import type { AccountReadRepository } from '../ports/account-read-repository.js';
/** Preserve revision, missing-resource and pagination policy in a generated response. */
export class GetAccountQueryHandler {
  constructor(private readRepository: AccountReadRepository) {}
  async execute(query: GetAccountQuery): Promise<GetAccountResponse> {
    const value = await this.readRepository.get(query.id);
    return value ? { account: value } : { error: { code: 'not_found', message: '' } };
  }
}
