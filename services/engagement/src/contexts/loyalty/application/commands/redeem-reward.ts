import type { RedeemRewardCommand } from '../../../../contracts/requests/generated/loyalty.js';
import { Rejection } from '../../../../foundation/domain.js';
import type { RealtimeBusPort, RealtimeEvent } from '../../../../foundation/application.js';
import type { CommandContext } from '../../../../foundation/write-repository.js';
import type { RewardWriteRepository } from '../ports/reward-write-repository.js';
import { ApplicationError } from '../../../../foundation/application.js';

/** Applies the reward redemption rule using the injected clock and one aggregate transaction. */
export class RedeemRewardCommandHandler {
  constructor(
    private repository: RewardWriteRepository,
    private realtime: RealtimeBusPort<RealtimeEvent>,
    private clock: () => Date,
  ) {}
  async execute(context: CommandContext, command: RedeemRewardCommand) {
    try {
      const loaded = await this.repository.get(context.target);
      if (!loaded) throw new RewardNotFoundApplicationError();
      const reward = loaded.state;
      reward.redeem(command.orderId, this.clock());
      await this.repository.save(reward);
      await this.realtime.publish({ snapshot: reward.snapshot() });
      return { status: 'redeemed' };
    } catch (error) {
      if (error instanceof Rejection) await this.realtime.publish({ rejection: error.outcome() });
      throw error;
    }
  }
}

/** Reports the expected rejection when the targeted reward has no stored state. */
export class RewardNotFoundApplicationError extends ApplicationError {
  constructor() {
    super('not_found', 'The reward does not exist');
  }
}
