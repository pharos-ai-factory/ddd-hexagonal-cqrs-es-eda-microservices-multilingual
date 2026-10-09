import {InternalCommandCodec} from '../../../../adaptors/internal-commands.js';
import {identifier} from '../../../../foundation/domain.js';
import schema from './generated/internal_commands.json' with {type: 'json'};
import type {RequestNotificationCommand} from '../../application/commands/request-notification.js';
import type {DeliverNotificationCommand} from '../../application/commands/deliver-notification.js';

/** Stable owner-local subscriptions; each installs a separate command queue. */
export enum CommunicationSubscription {
  PickupNotice = 'communication.pickup-notice', RewardNotice = 'communication.reward-notice',
  DeliverNotice = 'communication.deliver-notice',
}
function fields(value: unknown): Record<string, unknown> {
  if (!value || typeof value !== 'object' || Array.isArray(value)) throw new Error('Command object required');
  return value as Record<string, unknown>;
}
export function noticeCodec(consumer: CommunicationSubscription) {
  return new InternalCommandCodec('communication', consumer, 'requestNotification', schema, (value): RequestNotificationCommand => {
    const p = fields(value);
    if (typeof p.recipient !== 'string' || typeof p.subject !== 'string' || typeof p.body !== 'string') throw new Error('Invalid notification');
    return {recipient: p.recipient, subject: p.subject, body: p.body};
  });
}
export function deliveryCodec() {
  return new InternalCommandCodec('communication', CommunicationSubscription.DeliverNotice, 'deliverNotification', schema,
    (value): DeliverNotificationCommand => ({notificationId: identifier(String(fields(value).notificationId))}));
}
