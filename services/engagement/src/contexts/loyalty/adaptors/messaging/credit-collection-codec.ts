import { LoyaltyPrivateCommand, LoyaltySubscription } from '../../../../adaptors/generated/message-names.js';
import { parseCreditCollectionCommand } from './generated/request-payloads.js';
import { InternalCommandCodec } from '../../../../adaptors/messaging/internal-command-codec.js';

import { identifier } from '../../../../foundation/domain.js';

import schema from './generated/internal_commands.json' with { type: 'json' };

import type { CreditCollectionCommand } from '../../../../contracts/requests/generated/loyalty.js';

export const creditCodec = () =>
  new InternalCommandCodec(
    'loyalty',
    LoyaltySubscription.CREDIT_COLLECTION,
    LoyaltyPrivateCommand.CREDIT_COLLECTION,
    schema,
    (value): CreditCollectionCommand => {
      const command = parseCreditCollectionCommand(value);
      identifier(command.orderId);
      identifier(command.customerId);
      return command;
    },
  );
