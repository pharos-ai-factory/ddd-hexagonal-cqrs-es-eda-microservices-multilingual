import {AsyncLocalStorage} from 'node:async_hooks';
import amqp from 'amqplib';
import type {PoolClient} from 'pg';
import type {Outcome} from '../foundation/application.js';
import type {Database} from './postgres.js';
import {newId} from '../foundation/identity.js';
import {confirmed} from './broker.js';
import {failure} from './diagnostics.js';
import {pause} from './dispatch.js';

export type ReplyIntent = {id: string; encode: (outcome: Outcome) => Buffer; committed: boolean};
export const currentReply = new AsyncLocalStorage<ReplyIntent>();
export async function commitOutcome(client: PoolClient, outcome: Outcome): Promise<Outcome> {
  const intent = currentReply.getStore();
  if (intent) {
    const body = intent.encode(outcome);
    if (!body.length) throw new Error('Empty command reply');
    await client.query('INSERT INTO cafe.command_replies(id,body) VALUES($1,$2) ON CONFLICT DO NOTHING', [intent.id, body]);
    const {rows: [saved]} = await client.query<{body: Buffer}>('SELECT body FROM cafe.command_replies WHERE id=$1', [intent.id]);
    if (!saved || !saved.body.equals(body)) throw new Error('Request identity reused with different reply bytes');
    await client.query('INSERT INTO cafe.command_reply_dispatches(event_id) VALUES($1) ON CONFLICT DO NOTHING', [intent.id]);
  }
  await client.query('COMMIT');
  if (intent) intent.committed = true;
  return outcome;
}

export async function claimReply(db: Database) {
  const token = newId();
  const {rows: [row]} = await db.pool.query<{id: string; body: Buffer; generation: number; expired: boolean}>(`WITH candidate AS (
    SELECT event_id FROM cafe.command_reply_dispatches WHERE completed_at IS NULL AND available_at<=clock_timestamp()
    AND (lease_until IS NULL OR lease_until<clock_timestamp()) ORDER BY available_at,event_id FOR UPDATE SKIP LOCKED LIMIT 1
  ), claimed AS (
    UPDATE cafe.command_reply_dispatches d SET lease_token=$1,lease_until=clock_timestamp()+interval '8 seconds',
    generation=generation+1,attempts=attempts+1 FROM candidate c WHERE d.event_id=c.event_id RETURNING d.event_id,d.generation
  ) SELECT o.id,o.body,c.generation,o.expires_at<=clock_timestamp() AS expired
    FROM claimed c JOIN cafe.command_replies o ON o.id=c.event_id`, [token]);
  return row ? {...row, token} : undefined;
}
export async function replyRelay(db: Database, url: string, signal: AbortSignal) {
  while (!signal.aborted) {
    let connection: Awaited<ReturnType<typeof amqp.connect>> | undefined;
    try {
      connection = await amqp.connect(url);
      connection.on('error', error => failure(db.owner, 'replies.connection', error));
      const channel = await connection.createConfirmChannel();
      channel.on('error', error => failure(db.owner, 'replies.channel', error));
      while (!signal.aborted) {
        const row = await claimReply(db);
        if (!row) { await pause(signal); continue; }
        let reason: string | null = null;
        try {
          if (row.expired) reason = 'expired';
          else await confirmed(channel, 'cafe.replies', 'reply.'+db.owner, row.body, {
            contentType: 'application/x-protobuf', type: 'reply', appId: db.owner, messageId: row.id, correlationId: row.id,
          });
        } catch (error) { reason = 'publication_failed'; throw error; }
        finally {
          await db.pool.query(`UPDATE cafe.command_reply_dispatches SET lease_token=NULL,lease_until=NULL,
            available_at=clock_timestamp()+interval '1 second',last_error=$4,
            completed_at=CASE WHEN $4::text IS NULL OR $4='expired' THEN clock_timestamp() ELSE NULL END
            WHERE event_id=$1 AND lease_token=$2 AND generation=$3 AND lease_until>clock_timestamp()`,
            [row.id, row.token, row.generation, reason]);
        }
      }
    } catch (error) { failure(db.owner, 'replies.reconnect', error); await pause(signal, 1000); }
    finally { await connection?.close().catch(() => {}); }
  }
}
