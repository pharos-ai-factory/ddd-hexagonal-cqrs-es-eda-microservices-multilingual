import { claim, finish, type OutboxDispatch } from '../persistence/dispatch-leases.js';
import { pause } from '../../foundation/abortable-delay.js';
import { failure, errorClass } from '../diagnostics.js';

import type { PostgresContextDatabase } from '../persistence/postgres-database.js';

export async function realtimeRelay(
  db: PostgresContextDatabase,
  url: string,
  key: string,
  signal: AbortSignal,
) {
  while (!signal.aborted) {
    let row: OutboxDispatch | undefined;
    try {
      row = await claim(db, true);
      if (!row) {
        await pause(signal);
        continue;
      }
      const response = await fetch(url + '/api/publish', {
        method: 'POST',
        signal: AbortSignal.timeout(5000),
        headers: {
          'Content-Type': 'application/json',
          'X-API-Key': key,
          'X-Centrifugo-Error-Mode': 'transport',
        },
        body: JSON.stringify({
          channel: row.channel,
          b64data: row.body.toString('base64'),
          idempotency_key: row.id,
        }),
      });
      const body = (await response.json()) as { error?: unknown };
      if (!response.ok || body.error) throw new Error('Centrifugo publication failed');
      await finish(db, row, true);
    } catch (error) {
      failure(db.owner, 'realtime.publish', error, row?.id);
      if (row)
        await finish(db, row, true, errorClass(error)).catch((finishError) =>
          failure(db.owner, 'realtime.complete', finishError, row?.id),
        );
      await pause(signal, 1000);
    }
  }
}
