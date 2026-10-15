import {
  CommunicationPrivateCommand,
  CommunicationSubscription,
} from '../../../../adaptors/generated/message-names.js';
import { parseRequestNotificationCommand } from './generated/request-payloads.js';
import { InternalCommandCodec } from '../../../../adaptors/messaging/internal-command-codec.js';

import schema from './generated/internal_commands.json' with { type: 'json' };

import type { RequestNotificationCommand } from '../../../../contracts/requests/generated/communication.js';

export function noticeCodec(consumer: CommunicationSubscription) {
  return new InternalCommandCodec(
    'communication',
    consumer,
    CommunicationPrivateCommand.REQUEST_NOTIFICATION,
    schema,
    (value): RequestNotificationCommand => {
      const command = parseRequestNotificationCommand(value);

      return command;
    },
  );
}
