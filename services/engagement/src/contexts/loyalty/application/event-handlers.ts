import type {DurableCommandPort, Metadata} from '../../../foundation/application.js';
import type {OrderCollected} from '../../../contracts/events.js';
import type {RewardEarned} from './events.js';
import type {CreditCollectionCommand} from './commands/credit-collection.js';
import type {IssueRewardCommand} from './commands/issue-reward.js';

/** Maps Collection's fact into a durable intent owned by Loyalty. */
export class OrderCollectedIntegrationEventHandler {
  constructor(private commands: DurableCommandPort<CreditCollectionCommand>) {}
  handle(m: Metadata, event: OrderCollected) {
    return this.commands.enqueue(m, {orderId: event.orderId, customerId: event.customerId});
  }
}
/** Hands an immutable grant to the independent Reward aggregate command queue. */
export class RewardEarnedDomainEventHandler {
  constructor(private commands: DurableCommandPort<IssueRewardCommand>) {}
  handle(m: Metadata, event: RewardEarned) {
    return this.commands.enqueue(m, {grantId: event.grantId, accountId: event.accountId,
      benefit: event.benefit, validDays: event.validDays});
  }
}
