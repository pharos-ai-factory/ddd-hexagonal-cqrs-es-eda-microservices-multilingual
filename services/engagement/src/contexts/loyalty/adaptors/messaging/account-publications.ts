import type {LoyaltyAccount} from '../../domain/loyalty-account.js';
import type {Publication} from '../../../../foundation/application.js';

/** Map the saved aggregate's private facts to this owner's delivered messages. */
export function accountPublications(aggregate: LoyaltyAccount): Publication[] {
  return aggregate.events().flatMap(fact => fact.type === 'RewardEarned' ? [{name: 'loyalty.reward-earned', payload: {accountId: fact.grant.accountId, grantId: fact.grant.id, benefit: fact.grant.benefit, validDays: fact.grant.validDays}}] : []);
}
