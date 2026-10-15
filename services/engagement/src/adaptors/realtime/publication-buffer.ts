import type { RealtimeBusPort, RealtimeEvent } from '../../foundation/application.js';

/** Stage immutable messages; the owning transaction supplies durable delivery. */
export class RealtimeBuffer implements RealtimeBusPort<RealtimeEvent> {
  readonly events: RealtimeEvent[] = [];
  private active = true;

  // eslint-disable-next-line @typescript-eslint/require-await -- Match the asynchronous publication port.
  async publish(event: RealtimeEvent): Promise<void> {
    if (!this.active) throw new Error('Realtime bus used outside its transaction');
    if (Boolean(event.snapshot) === Boolean(event.rejection)) throw new Error('Select one realtime payload');
    this.events.push(structuredClone(event));
  }
  close() {
    this.active = false;
  }
}
