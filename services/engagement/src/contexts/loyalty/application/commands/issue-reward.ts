import type {CommandContext} from '../../../../foundation/write-repository.js';
import type {RewardWriteRepository} from '../ports/reward-write-repository.js';
import {Reward} from '../../domain/reward.js';

/** Carries the immutable grant so reward issuance can recover independently. */
export type IssueRewardCommand = {grantId: string; accountId: string; benefit: string; validDays: number};

/** Creates one Reward from its immutable earned grant and injected clock. */
export class IssueRewardCommandHandler {
  constructor(private repository: RewardWriteRepository, private clock: () => Date) {}
  async execute(context: CommandContext, command: IssueRewardCommand) {
    const loaded = await this.repository.get(context.target);
    if (loaded) {
      const state = loaded.state.snapshot();
      return {status: state.status};
    }
    const reward = Reward.issue(context.target, {id: command.grantId, accountId: command.accountId,
      benefit: command.benefit, validDays: command.validDays}, this.clock());
    await this.repository.save(reward);
    return {status: 'issued'};
  }
}
