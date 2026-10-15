import { parseGetRewardQuery, parseListRewardsQuery } from './generated/request-payloads.js';
import { identifier } from '../../../../foundation/domain.js';
import type { Loaded } from '../../../../foundation/application.js';
import type { Page } from '../../../../foundation/pagination.js';
import type { GetRewardQuery } from '../../../../contracts/requests/generated/loyalty.js';
import type { ListRewardsQuery } from '../../../../contracts/requests/generated/loyalty.js';
import type { RewardView } from '../../application/read-models/reward.js';
import { reward } from './reward-requests.js';

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

/** Map the resource revision or the existing missing-resource response. */
export function getReply(_query: GetRewardQuery, loaded: Loaded<RewardView> | undefined) {
  return loaded ? { reward: reward(loaded) } : { error: { code: 'not_found' } };
}

/** Retain pagination and its continuation from the handler result. */
export function listReply(query: ListRewardsQuery, page: Page<RewardView>) {
  return {
    rewards: {
      items: page.items.map(reward),
      paged: query.page !== undefined,
      ...(page.nextId ? { nextId: page.nextId } : {}),
    },
  };
}
