import {CorruptState, identifier, Rejection} from '../../../foundation/domain.js';
export type NotificationState = Readonly<{
  id: string; recipient: string; subject: string; body: string;
  status: 'requested' | 'sent'; providerReceipt?: string;
}>;
function validateContent(id: string, recipient: string, subject: string, body: string) {
  identifier(id); identifier(recipient);
  if (typeof subject !== 'string' || !subject || subject.length > 120 ||
      typeof body !== 'string' || !body || body.length > 4096) {
    throw new Rejection('invalid_notification', 'Notification content is missing or too large');
  }
}
/** Owns requested content and the recorded result of idempotent provider delivery. */
export class Notification {
  #events: {type: 'NotificationRequested'; notificationId: string}[] = [];
  #state: NotificationState;
  constructor(state: NotificationState) {
    try {
      validateContent(state.id, state.recipient, state.subject, state.body);
      if (!['requested', 'sent'].includes(state.status) ||
          (state.status === 'sent' && (typeof state.providerReceipt !== 'string' || !state.providerReceipt)) ||
          (state.status === 'requested' && state.providerReceipt !== undefined)) {
        throw new Error('Invalid delivery state');
      }
    } catch (cause) {
      throw new CorruptState('Corrupt notification state', {cause});
    }
    this.#state = {...state};
  }
  static request(id: string, recipient: string, subject: string, body: string) {
    validateContent(id, recipient, subject, body);
    const notification = new Notification({id, recipient, subject, body, status: 'requested'});
    notification.#events.push({type: 'NotificationRequested', notificationId: id});
    return notification;
  }
  recordDelivery(receipt: string): boolean {
    if (this.#state.status === 'sent') return false;
    if (!receipt) throw new Rejection('missing_delivery_receipt', 'Provider acceptance requires a receipt');
    this.#state = {...this.#state, status: 'sent', providerReceipt: receipt};
    return true;
  }
  events(): readonly {type: 'NotificationRequested'; notificationId: string}[] { return structuredClone(this.#events); }
  snapshot(): NotificationState { return {...this.#state}; }
}
