import { parseGetRewardQuery, parseListRewardsQuery } from './generated/request-payloads.js';
import { identifier } from '../../../../foundation/domain.js';
import type { GetRewardQuery } from '../../../../contracts/requests/generated/loyalty.js';
import type { ListRewardsQuery } from '../../../../contracts/requests/generated/loyalty.js';

/** Decode the published item query into its application input. */
export function getReward(value: unknown): GetRewardQuery {
  const query = parseGetRewardQuery(value);
  identifier(query.id);
  return query;
}

/** Preserve complete reads and explicitly requested pagination. */
export function listRewards(value: unknown): ListRewardsQuery {
  const query = parseListRewardsQuery(value);
  if (query.page?.after !== undefined) identifier(query.page.after);
  return query;
}
