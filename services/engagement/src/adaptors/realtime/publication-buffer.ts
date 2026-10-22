import type { RealtimeBusPort, RealtimeMessage } from '../../foundation/application.js';

/** Stage immutable messages; the owning transaction supplies durable delivery. */
export class RealtimeBuffer implements RealtimeBusPort<RealtimeMessage> {
  readonly events: RealtimeMessage[] = [];
  private active = true;

  // eslint-disable-next-line @typescript-eslint/require-await -- Match the asynchronous publication port.
  async publish(event: RealtimeMessage): Promise<void> {
    if (!this.active) throw new Error('Realtime bus used outside its transaction');
    if (Boolean(event.snapshot) === Boolean(event.rejection)) throw new Error('Select one realtime payload');
    if (!event.target.id || !['context', 'principal'].includes(event.target.kind))
      throw new Error('Select one realtime delivery target');
    this.events.push(structuredClone(event));
  }
  close() {
    this.active = false;
  }
}
