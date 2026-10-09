import {type AggregateCommandPort, type IdentityFactory, type Metadata} from '../../../../foundation/application.js';
import {Reward, type RewardState} from '../../domain/reward.js';

/** Carries the immutable grant so reward issuance can recover independently. */
export type IssueRewardCommand = {grantId: string; accountId: string; benefit: string; validDays: number};

/** Creates one Reward from its immutable earned grant and injected clock. */
export class IssueRewardCommandHandler {
  constructor(private rewards: AggregateCommandPort<RewardState>, private ids: IdentityFactory, private clock: () => Date) {}
  execute(metadata: Metadata, command: IssueRewardCommand) {
    return this.rewards.execute({...metadata, id: this.ids('issue-reward', command.grantId)}, loaded => {
      if (loaded) {
        const state = new Reward(loaded.state).snapshot();
        return {state, status: state.status, changed: false};
      }
      const reward = Reward.issue(metadata.target, {id: command.grantId, accountId: command.accountId,
        benefit: command.benefit, validDays: command.validDays}, this.clock());
      const state = reward.snapshot();
      return {state, status: 'issued', changed: true, publications: [{name: 'loyalty.reward-issued', payload: {
        rewardId: state.id, customerId: state.customerId, benefit: state.benefit, expiresAt: state.expiresAt}}]};
    });
  }
}
