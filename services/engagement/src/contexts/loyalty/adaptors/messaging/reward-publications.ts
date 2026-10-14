import { published } from '../../../../adaptors/generated/event-types.js';
import type { Reward } from '../../domain/reward.js';
import type { Publication } from '../../../../foundation/application.js';

/** Map the saved aggregate's private facts to this owner's delivered messages. */
export function rewardPublications(aggregate: Reward): Publication[] {
  return aggregate.events().map((fact) =>
    published('loyalty.reward-issued', {
      rewardId: fact.reward.id,
      customerId: fact.reward.customerId,
      benefit: fact.reward.benefit,
      expiresAt: fact.reward.expiresAt,
    }),
  );
}
