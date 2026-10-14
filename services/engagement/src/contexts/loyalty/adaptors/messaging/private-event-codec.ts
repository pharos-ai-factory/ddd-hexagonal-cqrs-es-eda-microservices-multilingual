import type { RewardEarned } from '../../application/events.js';
import protobuf from 'protobufjs';
import schema from './generated/private_messages.json' with { type: 'json' };
import { PrivateEventCodec } from '../../../../adaptors/private-event.js';
import { identifier } from '../../../../foundation/domain.js';

/** Loyalty validates the private grant evidence used to request reward issuance. */
export const privateEventCodec = new PrivateEventCodec(
  protobuf.Root.fromJSON(schema).lookupType('cafe.loyalty.internal.PrivateEvent'),
  { owner: 'loyalty', name: 'loyalty.reward-earned', kind: 'account', payload: 'rewardEarned' },
  (payload, source) => {
    for (const key of ['accountId', 'grantId']) {
      if (typeof payload[key] !== 'string') throw new Error('Missing grant identity');
      identifier(payload[key]);
    }
    if (
      payload.accountId !== source ||
      typeof payload.benefit !== 'string' ||
      !payload.benefit.trim() ||
      typeof payload.validDays !== 'number' ||
      !Number.isInteger(payload.validDays) ||
      payload.validDays < 1 ||
      payload.validDays > 30
    ) {
      throw new Error('Invalid grant evidence');
    }
  },
);

/** Private payload association owned by loyalty. */
export type PrivateEventPayloads = { 'loyalty.reward-earned': RewardEarned };
