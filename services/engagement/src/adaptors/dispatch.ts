import {failure, errorClass} from './diagnostics.js';
import type {Database} from './postgres.js';
import {newId} from '../foundation/identity.js';
export type Dispatch = {
  id: string; body: Buffer; channel?: string; token: string; generation: string;
};
export async function claim(db: Database, realtime = false): Promise<Dispatch | undefined> {
  const table = realtime ? 'realtime_dispatches' : 'dispatches';
  const source = realtime ? 'realtime_publications' : 'outbox_events';
  const token = newId();
  const {rows: [row]} = await db.pool.query(`WITH candidate AS (
    SELECT event_id FROM cafe.${table} WHERE completed_at IS NULL AND available_at<=clock_timestamp()
    AND (lease_until IS NULL OR lease_until<clock_timestamp()) ORDER BY available_at,event_id FOR UPDATE SKIP LOCKED LIMIT 1
  ), claimed AS (
    UPDATE cafe.${table} d SET lease_token=$1,lease_until=clock_timestamp()+interval '30 seconds',generation=generation+1
    FROM candidate c WHERE d.event_id=c.event_id RETURNING d.event_id,d.generation
  ) SELECT o.*,c.generation FROM claimed c JOIN cafe.${source} o ON o.id=c.event_id`, [token]);
  return row ? {...row, token} as Dispatch : undefined;
}
export async function finish(db: Database, row: Dispatch, realtime = false, error: string | null = null) {
  const table = realtime ? 'realtime_dispatches' : 'dispatches';
  await db.pool.query(`UPDATE cafe.${table} SET lease_token=NULL,lease_until=NULL,last_error=$4,
    available_at=clock_timestamp()+interval '1 second',
    completed_at=CASE WHEN $4::text IS NULL THEN clock_timestamp() ELSE NULL END
    WHERE event_id=$1 AND lease_token=$2 AND generation=$3 AND lease_until>clock_timestamp()`,
    [row.id, row.token, row.generation, error]);
}
export async function pause(signal: AbortSignal, ms = 100) {
  if (signal.aborted) return;
  await new Promise<void>(resolve => {
    const done = () => { clearTimeout(timer); signal.removeEventListener('abort', done); resolve(); };
    const timer = setTimeout(done, ms);
    signal.addEventListener('abort', done, {once: true});
  });
}
export async function realtimeRelay(db: Database, url: string, key: string, signal: AbortSignal) {
  while (!signal.aborted) {
    let row: Dispatch | undefined;
    try {
      row = await claim(db, true);
      if (!row) { await pause(signal); continue; }
      const response = await fetch(url+'/api/publish', {method: 'POST', signal: AbortSignal.timeout(5000),
        headers: {'Content-Type': 'application/json', 'X-API-Key': key, 'X-Centrifugo-Error-Mode': 'transport'},
        body: JSON.stringify({channel: row.channel, b64data: row.body.toString('base64'), idempotency_key: row.id})});
      const body = await response.json() as {error?: unknown};
      if (!response.ok || body.error) throw new Error('Centrifugo publication failed');
      await finish(db, row, true);
    } catch (error) {
      failure(db.owner, 'realtime.publish', error, row?.id);
      if (row) await finish(db, row, true, errorClass(error)).catch(finishError => failure(db.owner, 'realtime.complete', finishError, row?.id));
      await pause(signal, 1000);
    }
  }
}
