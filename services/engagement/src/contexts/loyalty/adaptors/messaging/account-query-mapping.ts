import { parseGetAccountQuery, parseListAccountsQuery } from './generated/request-payloads.js';
import { identifier } from '../../../../foundation/domain.js';
import type { GetAccountQuery } from '../../../../contracts/requests/generated/loyalty.js';
import type { ListAccountsQuery } from '../../../../contracts/requests/generated/loyalty.js';

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
