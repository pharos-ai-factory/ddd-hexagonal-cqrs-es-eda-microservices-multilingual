import {failure, errorClass} from './diagnostics.js';
import amqp, {type ConfirmChannel, type Options} from 'amqplib';
import {createHash} from 'node:crypto';
import type {Database} from './postgres.js';
import {decode, type WireEvent} from './codec.js';
import {claim, finish, pause} from './dispatch.js';
import {derivedId} from '../foundation/identity.js';
import type {Metadata, Outcome} from '../foundation/application.js';
import {Rejection} from '../foundation/domain.js';

// One outstanding publication on this channel. Returns precede confirmations.
export async function confirmed(channel: ConfirmChannel, exchange: string, key: string, body: Buffer, options: Options.Publish) {
  let returned = false;
  const onReturn = () => { returned = true; };
  channel.on('return', onReturn);
  let timeout: ReturnType<typeof setTimeout> | undefined;
  try {
    await Promise.race([
      new Promise<void>((resolve, reject) => channel.publish(exchange, key, body,
        {...options, persistent: true, mandatory: true}, error => error ? reject(error) : resolve())),
      new Promise<void>((_, reject) => { timeout = setTimeout(() => reject(new Error('Publisher confirmation timed out')), 5000); }),
    ]);
    if (returned) throw new Error('Mandatory publication was returned');
  } finally { if (timeout) clearTimeout(timeout); channel.removeListener('return', onReturn); }
}
export async function relay(db: Database, url: string, signal: AbortSignal) {
  while (!signal.aborted) {
    let connection: Awaited<ReturnType<typeof amqp.connect>> | undefined;
    try {
      connection = await amqp.connect(url);
      connection.on('error', error => failure(db.owner, 'outbox.connection', error));
      const channel = await connection.createConfirmChannel();
      channel.on('error', error => failure(db.owner, 'outbox.channel', error));
      while (!signal.aborted) {
        const row = await claim(db);
        if (!row) { await pause(signal); continue; }
        try {
          const event = decode(row.body);
          await confirmed(channel, 'cafe.events', event.visibility+'.'+event.name, row.body, {
            contentType: 'application/x-protobuf', messageId: event.id, type: event.name,
            appId: event.context, correlationId: event.correlationId,
            headers: {'contract-version': {'!': 'int32', value: 1}}});
          await finish(db, row);
        } catch (error) {
          failure(db.owner, 'outbox.publish', error, row.id);
          await finish(db, row, false, errorClass(error)); throw error;
        }
      }
    } catch (error) { failure(db.owner, 'outbox.reconnect', error); await pause(signal, 1000); }
    finally { await connection?.close().catch(() => {}); }
  }
}
export type Subscription = {
  owner: string; consumer: string; event: string; target(payload: object): string;
  handle(metadata: Metadata, payload: object): Promise<Outcome>;
};
function validProperties(properties: Options.Publish, event: WireEvent) {
  return properties.contentType === 'application/x-protobuf' && properties.deliveryMode === 2 &&
    properties.messageId === event.id && properties.type === event.name && properties.appId === event.context &&
    properties.correlationId === event.correlationId && properties.headers?.['contract-version'] === 1;
}
export async function consume(url: string, sub: Subscription, signal: AbortSignal) {
  while (!signal.aborted) {
    let connection: Awaited<ReturnType<typeof amqp.connect>> | undefined;
    try {
      connection = await amqp.connect(url);
      const current = connection;
      const channel = await current.createConfirmChannel();
      channel.on('error', error => failure(sub.owner, sub.consumer, error));
      current.on('error', error => failure(sub.owner, sub.consumer, error));
      const closed = new Promise<void>(resolve => current.once('close', resolve));
      const abort = () => { void current.close().catch(() => {}); };
      signal.addEventListener('abort', abort, {once: true});
      await channel.prefetch(1);
      await channel.consume('ref.'+sub.consumer, message => {
        if (!message) { void current.close().catch(() => {}); return; }
        void (async () => {
          let validating = true;
          let decoded: WireEvent | undefined;
          const incomingHeaders = message.properties.headers ?? {};
          const counter = 'ref-attempt' in incomingHeaders ? incomingHeaders['ref-attempt'] : 0;
          const validAttempt = typeof counter === 'number' && Number.isInteger(counter) && counter >= 0 && counter <= 3;
          const attempt = validAttempt ? counter : 0;
          try {
            if (!validAttempt) throw new Error('Invalid retry counter');
            const event = decode(message.content);
            decoded = event;
            if (event.name !== sub.event || (event.visibility === 'domain' && event.context !== sub.owner) ||
              !validProperties(message.properties, event)) throw new Error('Invalid delivery metadata');
            validating = false;
            const m: Metadata = {id: derivedId(sub.consumer, event.id), target: sub.target(event.payload),
              name: sub.consumer, correlation: event.correlationId, causation: event.id, input: event.payload,
              consumer: sub.consumer, sourceId: event.id, sourceHash: createHash('sha256').update(message.content).digest('hex')};
            const outcome = await sub.handle(m, event.payload);
            if (outcome.rejection) throw new Rejection(outcome.rejection.code, outcome.rejection.message);
          } catch (error) {
            const headers = {...message.properties.headers};
            delete headers['x-death'];
            const suffix = validating || error instanceof Rejection || attempt >= 3 ? '.dead' : '.retry';
            headers['ref-attempt'] = {'!': 'int32', value: attempt + 1};
            headers['ref-failure'] = errorClass(error);
            await confirmed(channel, 'ref.'+sub.owner+'.delivery', 'ref.'+sub.consumer+suffix, message.content,
              {...message.properties, headers});
            failure(sub.owner, sub.consumer, error, decoded?.id, decoded?.correlationId, suffix === '.dead');
          }
          channel.ack(message);
        })().catch(error => { failure(sub.owner, sub.consumer+'.transfer', error); void current.close().catch(() => {}); });
      }, {noAck: false});
      await closed;
      signal.removeEventListener('abort', abort);
    } catch (error) { failure(sub.owner, sub.consumer+'.reconnect', error); await pause(signal, 1000); }
    finally { await connection?.close().catch(() => {}); }
  }
}
