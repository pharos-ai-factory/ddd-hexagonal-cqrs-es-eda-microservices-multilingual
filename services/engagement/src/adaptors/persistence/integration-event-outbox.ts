import type { DurablePublicationPort, Metadata, Outcome, Publication } from '../../foundation/application.js';
import { identifier } from '../../foundation/domain.js';
import type { PostgresContextDatabase } from './postgres-database.js';
import { encode } from '../messaging/event-codec.js';
import { outcome as restoredOutcome } from './receipt-outcome.js';

/** Complete a private subscription and stage exact integration bytes in one transaction. */
export class PostgresPublicationOutbox<P> implements DurablePublicationPort<P> {
  constructor(
    private db: PostgresContextDatabase,
    private consumer: string,
    private kind: string,
    private publication: (payload: P) => Publication,
  ) {
    if (!consumer.startsWith(db.owner + '.')) throw new Error('Foreign publication subscription');
  }

  async publish(m: Metadata, payload: P): Promise<Outcome> {
    for (const value of [m.id, m.target, m.sourceId, m.correlation]) identifier(String(value));
    if (
      m.consumer !== this.consumer ||
      m.sourceKind !== this.kind ||
      typeof m.sourceVersion !== 'number' ||
      !Number.isSafeInteger(m.sourceVersion) ||
      m.sourceVersion < 1 ||
      !m.sourceTime ||
      !/^[a-f0-9]{64}$/.test(m.sourceHash ?? '')
    )
      throw new Error('Invalid publication receipt metadata');
    const target = 'publication:' + this.kind + ':' + m.target;
    const client = await this.db.pool.connect();
    try {
      await client.query('BEGIN');
      await client.query('SELECT pg_advisory_xact_lock(hashtextextended($1,0))', [
        'consumer:' + this.consumer + ':' + m.sourceId,
      ]);
      const {
        rows: [previous],
      } = await client.query<{ fingerprint: string; target: string; outcome: unknown }>(
        'SELECT fingerprint,target,outcome FROM cafe.consumer_receipts WHERE consumer=$1 AND event_id=$2',
        [this.consumer, m.sourceId],
      );
      if (previous) {
        if (previous.fingerprint !== m.sourceHash || previous.target !== target)
          throw new Error('Event identity reused with conflicting bytes or target');
        const saved = restoredOutcome(previous.outcome, m.target);
        await client.query('COMMIT');
        return saved;
      }
      const { event, body } = encode(
        this.db.owner,
        this.kind,
        m.target,
        m.sourceVersion,
        { ...m, id: m.sourceId! },
        this.publication(payload),
        m.sourceTime,
      );
      if (event.visibility !== 'integration')
        throw new Error('Publication handler requires an integration event');
      await client.query(
        `INSERT INTO cafe.outbox_events(id,event_name,visibility,aggregate_kind,aggregate_id,
        aggregate_version,correlation_id,causation_id,body) VALUES($1,$2,$3,$4,$5,$6,$7,$8,$9)`,
        [
          event.id,
          event.name,
          event.visibility,
          this.kind,
          m.target,
          m.sourceVersion,
          m.correlation,
          m.sourceId,
          body,
        ],
      );
      await client.query('INSERT INTO cafe.dispatches(event_id) VALUES($1)', [event.id]);
      const outcome: Outcome = {
        aggregateId: m.target,
        version: m.sourceVersion,
        status: 'publication_queued',
      };
      await client.query(
        'INSERT INTO cafe.consumer_receipts(consumer,event_id,fingerprint,target,outcome) VALUES($1,$2,$3,$4,$5)',
        [this.consumer, m.sourceId, m.sourceHash, target, JSON.stringify(outcome)],
      );
      await client.query('COMMIT');
      return outcome;
    } catch (error) {
      await client.query('ROLLBACK').catch(() => {});
      throw error;
    } finally {
      client.release();
    }
  }
}
