import { AsyncLocalStorage } from 'node:async_hooks';

import type { PoolClient } from 'pg';

import type { Outcome } from '../../foundation/application.js';

import type { PostgresContextDatabase } from './postgres-database.js';

import { newId } from '../../foundation/identity.js';

export type ReplyIntent = {
  route?: string;
  id: string;
  encode: (outcome: Outcome) => Buffer;
  committed: boolean;
};

export const currentReply = new AsyncLocalStorage<ReplyIntent>();

export async function commitOutcome(client: PoolClient, outcome: Outcome): Promise<Outcome> {
  const intent = currentReply.getStore();
  if (intent) {
    const body = intent.encode(outcome);
    if (!body.length) throw new Error('Empty command reply');
    await client.query(
      'INSERT INTO cafe.command_replies(id,body,reply_route) VALUES($1,$2,$3) ON CONFLICT DO NOTHING',
      [intent.id, body, intent.route ?? ''],
    );
    const {
      rows: [saved],
    } = await client.query<{ body: Buffer; reply_route: string }>(
      'SELECT body,reply_route FROM cafe.command_replies WHERE id=$1',
      [intent.id],
    );
    if (!saved || !saved.body.equals(body) || saved.reply_route !== (intent.route ?? ''))
      throw new Error('Request identity reused with different reply bytes');
    await client.query(
      'INSERT INTO cafe.command_reply_dispatches(event_id) VALUES($1) ON CONFLICT DO NOTHING',
      [intent.id],
    );
  }
  await client.query('COMMIT');
  if (intent) intent.committed = true;
  return outcome;
}

export async function claimReply(db: PostgresContextDatabase) {
  const token = newId();
  const {
    rows: [row],
  } = await db.pool.query<{
    id: string;
    body: Buffer;
    reply_route: string;
    generation: number;
    expired: boolean;
  }>(
    `WITH candidate AS (
    SELECT event_id FROM cafe.command_reply_dispatches WHERE completed_at IS NULL AND available_at<=clock_timestamp()
    AND (lease_until IS NULL OR lease_until<clock_timestamp()) ORDER BY available_at,event_id FOR UPDATE SKIP LOCKED LIMIT 1
  ), claimed AS (
    UPDATE cafe.command_reply_dispatches d SET lease_token=$1,lease_until=clock_timestamp()+interval '8 seconds',
    generation=generation+1,attempts=attempts+1 FROM candidate c WHERE d.event_id=c.event_id RETURNING d.event_id,d.generation
  ) SELECT o.id,o.body,o.reply_route,c.generation,o.expires_at<=clock_timestamp() AS expired
    FROM claimed c JOIN cafe.command_replies o ON o.id=c.event_id`,
    [token],
  );
  return row ? { ...row, token } : undefined;
}
/** Completes only the current reply lease; failed publications remain retryable. */
export async function finishReply(
  db: PostgresContextDatabase,
  row: NonNullable<Awaited<ReturnType<typeof claimReply>>>,
  reason: string | null,
) {
  await db.pool.query(
    `UPDATE cafe.command_reply_dispatches SET lease_token=NULL,lease_until=NULL,
            available_at=clock_timestamp()+interval '1 second',last_error=$4,
            completed_at=CASE WHEN $4::text IS NULL OR $4='expired' THEN clock_timestamp() ELSE NULL END
            WHERE event_id=$1 AND lease_token=$2 AND generation=$3 AND lease_until>clock_timestamp()`,
    [row.id, row.token, row.generation, reason],
  );
}
