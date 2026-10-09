import type {Reward} from '../../domain/reward.js';
import type {Publication} from '../../../../foundation/application.js';

/** Map the saved aggregate's private facts to this owner's delivered messages. */
export function rewardPublications(aggregate: Reward): Publication[] {
  return aggregate.events().map(fact => ({name: 'loyalty.reward-issued', payload: {rewardId: fact.reward.id, customerId: fact.reward.customerId, benefit: fact.reward.benefit, expiresAt: fact.reward.expiresAt}}));
}
