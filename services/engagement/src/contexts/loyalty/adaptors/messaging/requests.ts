import {identifier} from '../../../../foundation/domain.js';
import type {Loaded} from '../../../../foundation/application.js';
import {record, text} from '../../../../adaptors/request-values.js';
import type {loyalty as wire} from '../../../../adaptors/generated/request-types.js';
import type {AccountView} from '../../application/read-models/account.js';
import type {RewardView} from '../../application/read-models/reward.js';

export function redeem(value: unknown): {orderId: string} {
  const fields = record(value);
  const wire: wire.RedeemReward = {orderId: text(fields.orderId)};
  return {orderId: identifier(wire.orderId!)};
}
export function account(loaded: Loaded<AccountView>): wire.LoadedAccount {
  const s = loaded.state;
  return {exists: loaded.exists, version: loaded.version, state: {
    id: s.id, stampBalance: s.stampBalance, collections: s.collections, grantsEarned: s.grantsEarned,
    ...(s.lastGrant ? {lastGrant: {id: s.lastGrant.id, accountId: s.lastGrant.accountId,
      benefit: s.lastGrant.benefit, validDays: s.lastGrant.validDays}} : {}),
  }};
}
export function reward(loaded: Loaded<RewardView>): wire.LoadedReward {
  const s = loaded.state;
  return {exists: loaded.exists, version: loaded.version, state: {
    id: s.id, grantId: s.grantId, customerId: s.customerId, benefit: s.benefit,
    status: s.status, expiresAt: s.expiresAt, ...(s.redeemedFor ? {redeemedFor: s.redeemedFor} : {}),
  }};
}
