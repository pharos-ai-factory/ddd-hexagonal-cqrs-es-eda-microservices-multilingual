import type { PoolClient } from 'pg';
import { isDeepStrictEqual } from 'node:util';
import type { Metadata, Outcome, RealtimeEvent } from '../../foundation/application.js';
import { realtime } from '../realtime/projection-codec.js';

/** Bind explicit messages to the snapshot or rejection being committed. */
export async function persistRealtime(
  client: PoolClient,
  owner: string,
  kind: string,
  m: Metadata,
  outcome: Outcome,
  pending: unknown,
  events: RealtimeEvent[],
) {
  for (const event of events) {
    let payload: unknown,
      field = kind,
      messageKind = 'snapshot';
    if (event.snapshot) {
      if (outcome.rejection) continue;
      if (!pending || !isDeepStrictEqual(event.snapshot, pending))
        throw new Error('Realtime snapshot differs from committed state');
      payload = event.snapshot;
    } else {
      if (!event.rejection || !isDeepStrictEqual(event.rejection, outcome.rejection))
        throw new Error('Realtime rejection differs from committed outcome');
      payload = { commandId: m.id, correlationId: m.correlation, ...event.rejection };
      field = 'operationRejected';
      messageKind = 'operation_rejected';
    }
    const publication = realtime(owner, kind, m.target, outcome.version, payload, field);
    await appendRealtime(client, owner, kind, m.target, outcome.version, publication, messageKind);
  }
}

/** Stage exact bytes and dispatch together on the active owner transaction. */
export async function appendRealtime(
  client: PoolClient,
  owner: string,
  kind: string,
  target: string,
  revision: number,
  publication: { id: string; body: Buffer },
  messageKind = 'snapshot',
) {
  await client.query(
    `INSERT INTO cafe.realtime_publications
    (id,channel,aggregate_kind,aggregate_id,revision,body,message_kind) VALUES($1,$2,$3,$4,$5,$6,$7)`,
    [publication.id, 'cafe:' + owner, kind, target, revision, publication.body, messageKind],
  );
  await client.query('INSERT INTO cafe.realtime_dispatches(event_id) VALUES($1)', [publication.id]);
}
