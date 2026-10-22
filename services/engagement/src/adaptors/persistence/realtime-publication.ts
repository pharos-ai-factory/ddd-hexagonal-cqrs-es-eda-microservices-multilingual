import type { PoolClient } from 'pg';
import { isDeepStrictEqual } from 'node:util';
import type { Metadata, Outcome, RealtimeMessage } from '../../foundation/application.js';
import { identifier } from '../../foundation/domain.js';
import { realtime } from '../realtime/message-codec.js';

function realtimeChannel(owner: string, event: RealtimeMessage): string {
  if (event.target.kind === 'context') {
    if (event.target.id !== owner || !event.snapshot || event.rejection)
      throw new Error('Invalid context realtime target');
    return 'cafe:' + owner;
  }
  if (event.target.kind === 'principal') {
    if (event.snapshot || !event.rejection) throw new Error('Invalid principal realtime target');
    identifier(event.target.id);
    return 'principal:' + event.target.id;
  }
  throw new Error('Invalid realtime delivery target');
}

/** Bind explicit messages to the snapshot or rejection being committed. */
export async function persistRealtime(
  client: PoolClient,
  owner: string,
  kind: string,
  m: Metadata,
  outcome: Outcome,
  pending: unknown,
  events: RealtimeMessage[],
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
      if (!m.actorId || event.target.id !== m.actorId)
        throw new Error('Realtime rejection target differs from authenticated actor');
      payload = { commandId: m.id, correlationId: m.correlation, ...event.rejection };
      field = 'operationRejected';
      messageKind = 'operation_rejected';
    }
    const publication = realtime(owner, kind, m.target, outcome.version, event.target, payload, field);
    await appendRealtime(
      client,
      realtimeChannel(owner, event),
      kind,
      m.target,
      outcome.version,
      publication,
      messageKind,
    );
  }
}

/** Stage exact bytes and dispatch together on the active owner transaction. */
export async function appendRealtime(
  client: PoolClient,
  channel: string,
  kind: string,
  target: string,
  revision: number,
  publication: { id: string; body: Buffer },
  messageKind = 'snapshot',
) {
  await client.query(
    `INSERT INTO cafe.realtime_publications
    (id,channel,aggregate_kind,aggregate_id,revision,body,message_kind) VALUES($1,$2,$3,$4,$5,$6,$7)`,
    [publication.id, channel, kind, target, revision, publication.body, messageKind],
  );
  await client.query('INSERT INTO cafe.realtime_dispatches(event_id) VALUES($1)', [publication.id]);
}
