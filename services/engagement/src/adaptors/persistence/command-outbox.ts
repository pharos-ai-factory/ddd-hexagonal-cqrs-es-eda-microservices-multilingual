import { type InternalCommandCodec } from '../messaging/internal-command-codec.js';

import type { Metadata, Outcome, CommandBusPort } from '../../foundation/application.js';

import type { PostgresContextDatabase } from './postgres-database.js';

/** Commits the event receipt and immutable command before allowing the event ACK. */
export class PostgresDurableCommandOutbox<C extends object> implements CommandBusPort<C> {
  constructor(
    private db: PostgresContextDatabase,
    private codec: InternalCommandCodec<C>,
  ) {
    if (db.owner !== codec.owner) throw new Error('Command outbox crossed its owner');
  }
  async enqueue(m: Metadata, command: C): Promise<Outcome> {
    const body = this.codec.encode(m, command);
    const client = await this.db.pool.connect();
    try {
      await client.query('BEGIN');
      await client.query('SELECT pg_advisory_xact_lock(hashtextextended($1,0))', [
        'handoff:' + m.consumer + ':' + m.sourceId,
      ]);
      const {
        rows: [previous],
      } = await client.query<{ id: string; fingerprint: string; target: string }>(
        'SELECT id,fingerprint,target FROM cafe.internal_commands WHERE consumer=$1 AND event_id=$2',
        [m.consumer, m.sourceId],
      );
      if (previous) {
        if (previous.id !== m.id || previous.fingerprint !== m.sourceHash || previous.target !== m.target)
          throw new Error('Conflicting event hand-off identity');
      } else {
        await client.query(
          `INSERT INTO cafe.internal_commands(id,consumer,event_id,fingerprint,target,body)
          VALUES($1,$2,$3,$4,$5,$6)`,
          [m.id, m.consumer, m.sourceId, m.sourceHash, m.target, body],
        );
        await client.query('INSERT INTO cafe.internal_command_dispatches(event_id) VALUES($1)', [m.id]);
      }
      await client.query('COMMIT');
      return { aggregateId: m.target, version: 0, status: 'queued' };
    } catch (error) {
      await client.query('ROLLBACK').catch(() => {});
      throw error;
    } finally {
      client.release();
    }
  }
}
