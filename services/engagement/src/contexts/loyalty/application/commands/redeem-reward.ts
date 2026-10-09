import type {CommandContext} from '../../../../foundation/write-repository.js';
import type {RewardWriteRepository} from '../ports/reward-write-repository.js';
import {ApplicationError} from '../../../../foundation/application.js';
import {Reward} from '../../domain/reward.js';

export type RedeemRewardCommand = {orderId: string};

/** Applies the reward redemption rule using the injected clock and one aggregate transaction. */
export class RedeemRewardCommandHandler {
  constructor(private repository: RewardWriteRepository, private clock: () => Date) {}
  async execute(context: CommandContext, command: RedeemRewardCommand) {
    const loaded = await this.repository.get(context.target);
    if (!loaded) throw new RewardNotFoundApplicationError();
    const reward = loaded.state;
    reward.redeem(command.orderId, this.clock());
    await this.repository.save(reward);
    return {status: 'redeemed'};
  }
}

/** Reports the expected rejection when the targeted reward has no stored state. */
export class RewardNotFoundApplicationError extends ApplicationError {
  constructor() { super('not_found', 'The reward does not exist'); }
}
