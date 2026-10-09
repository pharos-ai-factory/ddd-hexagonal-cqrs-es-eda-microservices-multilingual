import {ApplicationError, type AggregateCommandPort, type Metadata} from '../../../../foundation/application.js';
import {Reward, type RewardState} from '../../domain/reward.js';

export type RedeemRewardCommand = {orderId: string};

/** Applies the reward redemption rule using the injected clock and one aggregate transaction. */
export class RedeemRewardCommandHandler {
  constructor(private rewards: AggregateCommandPort<RewardState>, private clock: () => Date) {}
  execute(m: Metadata, command: RedeemRewardCommand) {
    return this.rewards.execute(m, loaded => {
      if (!loaded) throw new RewardNotFoundApplicationError();
      const reward = new Reward(loaded.state);
      reward.redeem(command.orderId, this.clock());
      return {state: reward.snapshot(), status: 'redeemed', changed: true};
    });
  }
}

/** Reports the expected rejection when the targeted reward has no stored state. */
export class RewardNotFoundApplicationError extends ApplicationError {
  constructor() { super('not_found', 'The reward does not exist'); }
}
