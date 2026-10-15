import { parseGetAccountQuery, parseListAccountsQuery } from './generated/request-payloads.js';
import { identifier } from '../../../../foundation/domain.js';
import type { Loaded } from '../../../../foundation/application.js';
import type { Page } from '../../../../foundation/pagination.js';
import type { GetAccountQuery } from '../../../../contracts/requests/generated/loyalty.js';
import type { ListAccountsQuery } from '../../../../contracts/requests/generated/loyalty.js';
import type { AccountView } from '../../application/read-models/account.js';
import { account } from './account-replies.js';

/** Decode the published item query into its application input. */
export function getAccount(value: unknown): GetAccountQuery {
  const query = parseGetAccountQuery(value);
  identifier(query.id);
  return query;
}

/** Preserve complete reads and explicitly requested pagination. */
export function listAccounts(value: unknown): ListAccountsQuery {
  const query = parseListAccountsQuery(value);
  if (query.page?.after !== undefined) identifier(query.page.after);
  return query;
}

/** Map the resource revision or the existing missing-resource response. */
export function getReply(_query: GetAccountQuery, loaded: Loaded<AccountView> | undefined) {
  return loaded ? { account: account(loaded) } : { error: { code: 'not_found' } };
}

/** Retain pagination and its continuation from the handler result. */
export function listReply(query: ListAccountsQuery, page: Page<AccountView>) {
  return {
    accounts: {
      items: page.items.map(account),
      paged: query.page !== undefined,
      ...(page.nextId ? { nextId: page.nextId } : {}),
    },
  };
}
