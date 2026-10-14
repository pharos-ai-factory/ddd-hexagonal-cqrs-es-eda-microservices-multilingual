import { LoyaltyDomainEvent } from '../../../../adaptors/generated/message-names.js';
import type {
  RewardEarnedDomainEventPayload,
  RewardIssuedDomainEventPayload,
} from '../../application/events.js';
import protobuf from 'protobufjs';
import schema from './generated/private_messages.json' with { type: 'json' };
import { PrivateEventCodec } from '../../../../adaptors/messaging/private-event-codec.js';
import { identifier } from '../../../../foundation/domain.js';
import { timestamp } from '../../../../adaptors/messaging/event-payload-codec.js';

/** Loyalty validates the private grant evidence used to request reward issuance. */
export const privateEventCodec = new PrivateEventCodec(
  protobuf.Root.fromJSON(schema).lookupType('cafe.loyalty.internal.PrivateEvent'),
  { owner: 'loyalty', name: LoyaltyDomainEvent.REWARD_EARNED, kind: 'account', payload: 'rewardEarned' },
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
export type PrivateEventPayloads = {
  [LoyaltyDomainEvent.REWARD_EARNED]: RewardEarnedDomainEventPayload;
  [LoyaltyDomainEvent.REWARD_ISSUED]: RewardIssuedDomainEventPayload;
};

/** Reward publication validates source identity and the immutable issued terms. */
const rewardIssuedCodec = new PrivateEventCodec(
  protobuf.Root.fromJSON(schema).lookupType('cafe.loyalty.internal.PrivateEvent'),
  { owner: 'loyalty', name: LoyaltyDomainEvent.REWARD_ISSUED, kind: 'reward', payload: 'rewardIssued' },
  (payload, source) => {
    if (typeof payload.rewardId !== 'string' || typeof payload.customerId !== 'string')
      throw new Error('Missing reward identity');
    identifier(payload.rewardId);
    identifier(payload.customerId);
    if (
      payload.rewardId !== source ||
      typeof payload.benefit !== 'string' ||
      !payload.benefit.trim() ||
      !timestamp(payload.expiresAt)
    )
      throw new Error('Invalid issued reward evidence');
  },
);
export const privateEventCodecs = {
  [LoyaltyDomainEvent.REWARD_EARNED]: privateEventCodec,
  [LoyaltyDomainEvent.REWARD_ISSUED]: rewardIssuedCodec,
};
