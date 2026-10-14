import { CommunicationSubscription } from './communication-subscriptions.js';
import { fields } from './command-fields.js';
import { InternalCommandCodec } from '../../../../adaptors/messaging/internal-command-codec.js';

import { identifier } from '../../../../foundation/domain.js';

import schema from './generated/internal_commands.json' with { type: 'json' };

import type { DeliverNotificationCommand } from '../../application/commands/deliver-notification.js';

export function deliveryCodec() {
  return new InternalCommandCodec(
    'communication',
    CommunicationSubscription.DeliverNotice,
    'deliverNotification',
    schema,
    (value): DeliverNotificationCommand => ({
      notificationId: identifier(String(fields(value).notificationId)),
    }),
  );
}
