import { LoyaltySubscription } from './loyalty-subscriptions.js';
import { fields } from './command-fields.js';
import { InternalCommandCodec } from '../../../../adaptors/messaging/internal-command-codec.js';

import { identifier } from '../../../../foundation/domain.js';

import schema from './generated/internal_commands.json' with { type: 'json' };

import type { CreditCollectionCommand } from '../../application/commands/credit-collection.js';

export const creditCodec = () =>
  new InternalCommandCodec(
    'loyalty',
    LoyaltySubscription.CreditCollection,
    'creditCollection',
    schema,
    (value): CreditCollectionCommand => {
      const p = fields(value);
      return { orderId: identifier(String(p.orderId)), customerId: identifier(String(p.customerId)) };
    },
  );
