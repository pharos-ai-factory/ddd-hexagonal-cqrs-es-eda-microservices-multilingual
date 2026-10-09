import type {DeliveryPort} from '../foundation/application.js';
/** Calls the delivery provider using the notification identity as its idempotency key. */
export class HttpNotificationDelivery implements DeliveryPort {
  constructor(private url: string, private key: string) {}
  async deliver(message: {id: string; recipient: string; subject: string; body: string}): Promise<string> {
    const {id, recipient, subject, body} = message;
    const response = await fetch(this.url+'/messages', {method: 'POST', redirect: 'error',
      signal: AbortSignal.timeout(5000), headers: {'Authorization': 'Bearer '+this.key,
        'Content-Type': 'application/json', 'Idempotency-Key': id},
      body: JSON.stringify({id, recipient, subject, body})});
    if (!response.ok) throw new Error('Provider temporarily unavailable');
    const result = await response.json() as {receipt?: string};
    if (!result.receipt) throw new Error('Provider acceptance omitted its receipt');
    return result.receipt;
  }
}
