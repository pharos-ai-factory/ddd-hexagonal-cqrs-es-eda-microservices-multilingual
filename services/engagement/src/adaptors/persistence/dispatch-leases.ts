import type { PostgresContextDatabase } from './postgres-database.js';

import { newId } from '../../foundation/identity.js';

export type OutboxDispatch = {
  id: string;
  body: Buffer;
  channel?: string;
  token: string;
  generation: string;
};

export async function claim(
  db: PostgresContextDatabase,
  realtime: boolean | 'commands' = false,
): Promise<OutboxDispatch | undefined> {
  const table =
    realtime === 'commands' ? 'internal_command_dispatches' : realtime ? 'realtime_dispatches' : 'dispatches';
  const source =
    realtime === 'commands' ? 'internal_commands' : realtime ? 'realtime_publications' : 'outbox_events';
  const token = newId();
  const {
    rows: [row],
  } = await db.pool.query<Omit<OutboxDispatch, 'token'>>(
    `WITH candidate AS (
    SELECT event_id FROM cafe.${table} WHERE completed_at IS NULL AND available_at<=clock_timestamp()
    AND (lease_until IS NULL OR lease_until<clock_timestamp()) ORDER BY available_at,event_id FOR UPDATE SKIP LOCKED LIMIT 1
  ), claimed AS (
    UPDATE cafe.${table} d SET lease_token=$1,lease_until=clock_timestamp()+interval '30 seconds',generation=generation+1
    FROM candidate c WHERE d.event_id=c.event_id RETURNING d.event_id,d.generation
  ) SELECT o.*,c.generation FROM claimed c JOIN cafe.${source} o ON o.id=c.event_id`,
    [token],
  );
  return row ? { ...row, token } : undefined;
}

export async function finish(
  db: PostgresContextDatabase,
  row: OutboxDispatch,
  realtime: boolean | 'commands' = false,
  error: string | null = null,
) {
  const table =
    realtime === 'commands' ? 'internal_command_dispatches' : realtime ? 'realtime_dispatches' : 'dispatches';
  await db.pool.query(
    `UPDATE cafe.${table} SET lease_token=NULL,lease_until=NULL,last_error=$4,
    available_at=clock_timestamp()+interval '1 second',
    completed_at=CASE WHEN $4::text IS NULL THEN clock_timestamp() ELSE NULL END
    WHERE event_id=$1 AND lease_token=$2 AND generation=$3 AND lease_until>clock_timestamp()`,
    [row.id, row.token, row.generation, error],
  );
}
