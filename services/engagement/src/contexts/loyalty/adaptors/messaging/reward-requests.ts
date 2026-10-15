import { parseRedeemRewardCommand } from './generated/request-payloads.js';
import type { RedeemRewardCommand } from '../../../../contracts/requests/generated/loyalty.js';
import { identifier } from '../../../../foundation/domain.js';
import type { Loaded } from '../../../../foundation/application.js';
import type { loyalty as wire } from '../../../../adaptors/generated/request-types.js';
import type { RewardView } from '../../application/read-models/reward.js';

export function redeem(value: unknown): RedeemRewardCommand {
  const command = parseRedeemRewardCommand(value);
  identifier(command.orderId);
  return command;
}
export function reward(loaded: Loaded<RewardView>): wire.LoadedReward {
  const s = loaded.state;
  return {
    exists: loaded.exists,
    version: loaded.version,
    state: {
      id: s.id,
      grantId: s.grantId,
      customerId: s.customerId,
      benefit: s.benefit,
      status: s.status,
      expiresAt: s.expiresAt,
      ...(s.redeemedFor ? { redeemedFor: s.redeemedFor } : {}),
    },
  };
}
