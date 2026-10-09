import type {Page, PageRequest} from '../../../foundation/pagination.js';
import type {AccountView} from '../application/read-models/account.js';
import type {GetAccountQueryHandler} from '../application/queries/get-account.js';
import type {ListAccountsQueryHandler} from '../application/queries/list-accounts.js';

/** Translate RabbitMQ query arguments into named loyalty use cases. */
export function accountQueryEndpoints(get: GetAccountQueryHandler, list: ListAccountsQueryHandler) {
  return {
    get: (id: string) => get.execute({id}),
    list: async () => (await list.execute({})).items,
    page: async (page: PageRequest): Promise<Page<AccountView>> => list.execute({page}),
  };
}
