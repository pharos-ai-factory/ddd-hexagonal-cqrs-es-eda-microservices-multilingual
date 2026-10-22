import { realtimeTarget, type RealtimeBusPort, type RealtimeMessage } from '../../foundation/application.js';
import { identifier } from '../../foundation/domain.js';
import type { PostgresContextDatabase } from './postgres-database.js';
import { checkRootIdentity } from './receipt-outcome.js';
import { appendRealtime } from './realtime-publication.js';
import { realtime } from '../realtime/message-codec.js';

/** Commit a declared event subscription's receipt and exact browser bytes before ACK. */
export class PostgresRealtimeMessageBus implements RealtimeBusPort<RealtimeMessage> {
  constructor(
    private db: PostgresContextDatabase,
    private consumer: string,
    private kind: string,
  ) {
    if (!consumer.startsWith(db.owner + '.')) throw new Error('Foreign realtime subscription');
  }
  async publish(event: RealtimeMessage): Promise<void> {
    const m = event.source;
    if (!m || !event.snapshot || !Number.isSafeInteger(event.revision) || event.revision! < 1)
      throw new Error('Invalid realtime delivery');
    if (event.target.kind !== 'context' || event.target.id !== realtimeTarget.context(this.db.owner).id)
      throw new Error('Invalid realtime delivery target');
    for (const value of [m.id, m.target, m.sourceId, m.correlation]) identifier(String(value));
    if (m.consumer !== this.consumer || !/^[a-f0-9]{64}$/.test(m.sourceHash ?? ''))
      throw new Error('Invalid realtime receipt');
    checkRootIdentity(event.snapshot, m.target);
    const client = await this.db.pool.connect(),
      target = 'realtime:' + this.kind + ':' + m.target;
    try {
      await client.query('BEGIN');
      await client.query('SELECT pg_advisory_xact_lock(hashtextextended($1,0))', [
        'consumer:' + this.consumer + ':' + m.sourceId,
      ]);
      const {
        rows: [previous],
      } = await client.query<{ fingerprint: string; target: string }>(
        'SELECT fingerprint,target FROM cafe.consumer_receipts WHERE consumer=$1 AND event_id=$2',
        [this.consumer, m.sourceId],
      );
      if (previous) {
        if (previous.fingerprint !== m.sourceHash || previous.target !== target)
          throw new Error('Conflicting realtime delivery identity');
      } else {
        const publication = realtime(
          this.db.owner,
          this.kind,
          m.target,
          event.revision!,
          event.target,
          event.snapshot,
        );
        await appendRealtime(
          client,
          'cafe:' + this.db.owner,
          this.kind,
          m.target,
          event.revision!,
          publication,
        );
        const outcome = { aggregateId: m.target, version: event.revision, status: 'realtime_queued' };
        await client.query(
          'INSERT INTO cafe.consumer_receipts(consumer,event_id,fingerprint,target,outcome) VALUES($1,$2,$3,$4,$5)',
          [this.consumer, m.sourceId, m.sourceHash, target, outcome],
        );
      }
      await client.query('COMMIT');
    } catch (error) {
      await client.query('ROLLBACK').catch(() => {});
      throw error;
    } finally {
      client.release();
    }
  }
}
