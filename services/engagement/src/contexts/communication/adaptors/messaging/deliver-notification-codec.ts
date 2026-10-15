import {
  CommunicationPrivateCommand,
  CommunicationSubscription,
} from '../../../../adaptors/generated/message-names.js';
import { parseDeliverNotificationCommand } from './generated/request-payloads.js';
import { InternalCommandCodec } from '../../../../adaptors/messaging/internal-command-codec.js';

import { identifier } from '../../../../foundation/domain.js';

import schema from './generated/internal_commands.json' with { type: 'json' };

import type { DeliverNotificationCommand } from '../../../../contracts/requests/generated/communication.js';

export function deliveryCodec() {
  return new InternalCommandCodec(
    'communication',
    CommunicationSubscription.DELIVER_NOTICE,
    CommunicationPrivateCommand.DELIVER_NOTIFICATION,
    schema,
    (value): DeliverNotificationCommand => {
      const command = parseDeliverNotificationCommand(value);
      identifier(command.notificationId);
      return command;
    },
  );
}
