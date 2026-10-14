import { CommunicationSubscription } from './communication-subscriptions.js';
import { fields } from './command-fields.js';
import { InternalCommandCodec } from '../../../../adaptors/messaging/internal-command-codec.js';

import schema from './generated/internal_commands.json' with { type: 'json' };

import type { RequestNotificationCommand } from '../../application/commands/request-notification.js';

export function noticeCodec(consumer: CommunicationSubscription) {
  return new InternalCommandCodec(
    'communication',
    consumer,
    'requestNotification',
    schema,
    (value): RequestNotificationCommand => {
      const p = fields(value);
      if (typeof p.recipient !== 'string' || typeof p.subject !== 'string' || typeof p.body !== 'string')
        throw new Error('Invalid notification');
      return { recipient: p.recipient, subject: p.subject, body: p.body };
    },
  );
}
