import { record, integer, text } from '../../../../adaptors/messaging/request-values.js';
import { identifier } from '../../../../foundation/domain.js';
import type { Loaded } from '../../../../foundation/application.js';
import type { Page } from '../../../../foundation/pagination.js';
import type { GetRewardQuery } from '../../application/queries/get-reward.js';
import type { ListRewardsQuery } from '../../application/queries/list-rewards.js';
import type { RewardView } from '../../application/read-models/reward.js';
import { reward } from './reward-requests.js';

/** Decode the published item query into its application input. */
export function getReward(value: unknown): GetRewardQuery {
  return { id: identifier(text(record(value).id)) };
}

/** Preserve complete reads and explicitly requested pagination. */
export function listRewards(value: unknown): ListRewardsQuery {
  const fields = record(value);
  if (!('page' in fields)) return {};
  const page = record(fields.page);
  return {
    page: { limit: integer(page.limit), ...('after' in page ? { after: identifier(text(page.after)) } : {}) },
  };
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
