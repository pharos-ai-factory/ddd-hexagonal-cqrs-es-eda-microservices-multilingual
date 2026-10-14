import { record, integer, text } from '../../../../adaptors/messaging/request-values.js';
import { identifier } from '../../../../foundation/domain.js';
import type { Loaded } from '../../../../foundation/application.js';
import type { Page } from '../../../../foundation/pagination.js';
import type { GetAccountQuery } from '../../application/queries/get-account.js';
import type { ListAccountsQuery } from '../../application/queries/list-accounts.js';
import type { AccountView } from '../../application/read-models/account.js';
import { account } from './account-replies.js';

/** Decode the published item query into its application input. */
export function getAccount(value: unknown): GetAccountQuery {
  return { id: identifier(text(record(value).id)) };
}

/** Preserve complete reads and explicitly requested pagination. */
export function listAccounts(value: unknown): ListAccountsQuery {
  const fields = record(value);
  if (!('page' in fields)) return {};
  const page = record(fields.page);
  return {
    page: { limit: integer(page.limit), ...('after' in page ? { after: identifier(text(page.after)) } : {}) },
  };
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
