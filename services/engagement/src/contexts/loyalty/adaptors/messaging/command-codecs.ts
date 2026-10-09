import {InternalCommandCodec} from '../../../../adaptors/internal-commands.js';
import {identifier} from '../../../../foundation/domain.js';
import schema from './generated/internal_commands.json' with {type: 'json'};
import type {CreditCollectionCommand} from '../../application/commands/credit-collection.js';
import type {IssueRewardCommand} from '../../application/commands/issue-reward.js';

/** Stable queue identities survive class renames, releases and replay. */
export enum LoyaltySubscription {
  CreditCollection = 'loyalty.credit-collection',
  IssueReward = 'loyalty.issue-reward',
}
function fields(value: unknown): Record<string, unknown> {
  if (!value || typeof value !== 'object' || Array.isArray(value)) throw new Error('Command object required');
  return value as Record<string, unknown>;
}
export const creditCodec = () => new InternalCommandCodec('loyalty', LoyaltySubscription.CreditCollection, 'creditCollection', schema,
  (value): CreditCollectionCommand => {const p = fields(value); return {orderId: identifier(String(p.orderId)), customerId: identifier(String(p.customerId))};});
export const issueCodec = () => new InternalCommandCodec('loyalty', LoyaltySubscription.IssueReward, 'issueReward', schema,
  (value): IssueRewardCommand => {
    const p = fields(value);
    if (typeof p.benefit !== 'string' || !Number.isInteger(p.validDays) || Number(p.validDays) < 1) throw new Error('Invalid grant');
    return {grantId: identifier(String(p.grantId)), accountId: identifier(String(p.accountId)), benefit: p.benefit, validDays: Number(p.validDays)};
  });
