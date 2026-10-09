import {once} from 'node:events';
import {setTimeout} from 'node:timers/promises';
import {PostgresContextDatabase} from '../src/adaptors/postgres.js';

// Observe the real pool in a separate process, without allowing a fatal error
// to terminate the test runner or print connection credentials.
process.on('uncaughtException', () => process.exit(86));
const deadline = globalThis.setTimeout(() => process.exit(87), 10000);
const db = new PostgresContextDatabase('loyalty', process.env.LOYALTY_DATABASE_URL!);
try {
  await db.verify();
  const {rows: [row]} = await db.pool.query('SELECT pg_backend_pid() AS pid');
  const resume = once(process, 'message');
  process.send!({pid: row.pid});
  await resume;
  while (db.pool.totalCount !== 0) await setTimeout(10);
  const {rows: [recovered]} = await db.pool.query('SELECT 1 AS value');
  if (recovered.value !== 1) throw new Error('Pool did not recover');
} finally {
  clearTimeout(deadline);
  await db.pool.end();
  process.disconnect!();
}
