import type {PostgresContextDatabase} from './postgres.js';

type Failure = {at: string; class: string; event: string | undefined; correlation: string | undefined};
type Worker = {owner: string; name: string; failures: number; deadLetterTransfers: number; lastFailure: Failure};
const workers = new Map<string, Worker>();
// Never retain error messages: database and HTTP errors may contain credentials or bodies.
export function errorClass(error: unknown): string {
  return error instanceof Error ? error.constructor.name : 'UnknownError';
}
export function failure(owner: string, name: string, error: unknown, event?: string, correlation?: string, dead = false) {
  const key = owner+'/'+name, previous = workers.get(key);
  const lastFailure = {at: new Date().toISOString(), class: errorClass(error), event, correlation};
  workers.set(key, {owner, name, failures: (previous?.failures ?? 0)+1,
    deadLetterTransfers: (previous?.deadLetterTransfers ?? 0)+(dead ? 1 : 0), lastFailure});
  console.warn(JSON.stringify({message: 'workflow failure', owner, worker: name, ...lastFailure, deadLetterTransfer: dead}));
}
export function processWorkers() { return Object.fromEntries(workers); }
export async function diagnostics(databases: Record<string, PostgresContextDatabase>) {
  const states: Record<string, unknown> = {};
  for (const [owner, db] of Object.entries(databases)) {
    const state: Record<string, unknown> = {};
    for (const [name, dispatch, source] of [['outbox', 'dispatches', 'outbox_events'],
      ['realtime', 'realtime_dispatches', 'realtime_publications'],
      ['commands', 'internal_command_dispatches', 'internal_commands']]) {
      try {
        const client = await db.pool.connect();
        try {
        await client.query("BEGIN");
        await client.query("SET LOCAL statement_timeout='3s'");
        const {rows: [row]} = await client.query(`SELECT count(*)::int AS pending,
          COALESCE(EXTRACT(EPOCH FROM clock_timestamp()-min(o.created_at)),0)::float8 AS "oldestAgeSeconds",
          count(*) FILTER (WHERE d.last_error IS NOT NULL)::int AS failed
          FROM cafe.${dispatch} d JOIN cafe.${source} o ON o.id=d.event_id WHERE d.completed_at IS NULL`);
        await client.query("COMMIT");
        state[name!] = row;
        } catch (error) { await client.query("ROLLBACK"); throw error; }
        finally { client.release(); }
      } catch (error) { failure(owner, 'diagnostics', error); throw new Error('Diagnostics unavailable'); }
    }
    states[owner] = state;
  }
  return {databases: states, processWorkers: processWorkers(),
    counterScope: 'process lifetime; deadLetterTransfers are confirmed transfers, not queue depths'};
}
