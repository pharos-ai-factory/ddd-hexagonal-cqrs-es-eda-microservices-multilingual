import { parseRedeemRewardCommand } from './generated/request-payloads.js';
import type { RedeemRewardCommand } from '../../../../contracts/requests/generated/loyalty.js';
import { identifier } from '../../../../foundation/domain.js';

export function redeem(value: unknown): RedeemRewardCommand {
  const command = parseRedeemRewardCommand(value);
  identifier(command.orderId);
  return command;
}
