import type { CommandBusPort, Metadata } from '../../../../foundation/application.js';
import type { RewardEarnedDomainEventPayload } from '../events.js';
import type { IssueRewardCommand } from '../commands/issue-reward.js';

/** Hands an immutable grant to the independent Reward aggregate command queue. */
export class RewardEarnedDomainEventHandler {
  constructor(private commands: CommandBusPort<IssueRewardCommand>) {}
  handle(m: Metadata, event: RewardEarnedDomainEventPayload) {
    return this.commands.enqueue(m, {
      grantId: event.grantId,
      accountId: event.accountId,
      benefit: event.benefit,
      validDays: event.validDays,
    });
  }
}
