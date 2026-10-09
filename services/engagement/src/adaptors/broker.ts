import {failure, errorClass} from './diagnostics.js';
import amqp, {type ConfirmChannel, type Options} from 'amqplib';
import {createHash} from 'node:crypto';
import type {PostgresContextDatabase} from './postgres.js';
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
export async function relay(db: PostgresContextDatabase, url: string, signal: AbortSignal, commandHeader?: (body: Buffer) => {id: string; name: string; context: string; correlationId: string}) {
  while (!signal.aborted) {
    let connection: Awaited<ReturnType<typeof amqp.connect>> | undefined;
    try {
      connection = await amqp.connect(url);
      connection.on('error', error => failure(db.owner, 'outbox.connection', error));
      const channel = await connection.createConfirmChannel();
      channel.on('error', error => failure(db.owner, 'outbox.channel', error));
      while (!signal.aborted) {
        const row = await claim(db, commandHeader ? 'commands' : false);
        if (!row) { await pause(signal); continue; }
        try {
          const event = commandHeader ? commandHeader(row.body) : decode(row.body);
          await confirmed(channel, commandHeader ? 'ref.'+db.owner+'.delivery' : 'cafe.events',
            commandHeader ? 'ref.'+event.name : (event as WireEvent).visibility+'.'+event.name, row.body, {
            contentType: 'application/x-protobuf', messageId: event.id, type: event.name,
            appId: event.context, correlationId: event.correlationId,
            headers: {'contract-version': {'!': 'int32', value: 1}}});
          await finish(db, row, commandHeader ? 'commands' : false);
        } catch (error) {
          failure(db.owner, 'outbox.publish', error, row.id);
          await finish(db, row, commandHeader ? 'commands' : false, errorClass(error)); throw error;
        }
      }
    } catch (error) { failure(db.owner, 'outbox.reconnect', error); await pause(signal, 1000); }
    finally { await connection?.close().catch(() => {}); }
  }
}
export type EventSubscription = {
  owner: string; consumer: string; event: string; target(payload: object): string;
  handle(metadata: Metadata, payload: object): Promise<Outcome>;
  decodeCommand?: (body: Buffer) => {metadata: Metadata; payload: object};
};
function validProperties(properties: Options.Publish, event: WireEvent) {
  return properties.contentType === 'application/x-protobuf' && properties.deliveryMode === 2 &&
    properties.messageId === event.id && properties.type === event.name && properties.appId === event.context &&
    properties.correlationId === event.correlationId && properties.headers?.['contract-version'] === 1;
}
export async function consume(url: string, sub: EventSubscription, signal: AbortSignal) {
  const queue = 'ref.'+sub.consumer+(sub.decodeCommand ? '.command' : '');
  const inFlight = new Set<Promise<void>>();
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
        if (signal.aborted) { abort(); return; }
      await channel.prefetch(1);
      await channel.consume(queue, message => {
        if (!message) { void current.close().catch(() => {}); return; }
        const work = (async () => {
          let validating = true;
          let identity: {id: string; correlationId: string} | undefined;
          const incomingHeaders = message.properties.headers ?? {};
          const counter = 'ref-attempt' in incomingHeaders ? incomingHeaders['ref-attempt'] : 0;
          const validAttempt = typeof counter === 'number' && Number.isInteger(counter) && counter >= 0 && counter <= 3;
          const attempt = validAttempt ? counter : 0;
          try {
            if (!validAttempt) throw new Error('Invalid retry counter');
            let m: Metadata, payload: object;
            if (sub.decodeCommand) {
              ({metadata: m, payload} = sub.decodeCommand(message.content));
              identity = {id: m.id, correlationId: m.correlation};
              const p = message.properties;
              if (p.contentType !== 'application/x-protobuf' || p.deliveryMode !== 2 || p.messageId !== m.id ||
                p.type !== sub.consumer+'.command' || p.appId !== sub.owner || p.correlationId !== m.correlation ||
                p.headers?.['contract-version'] !== 1) throw new Error('Invalid command properties');
            } else {
              const event = decode(message.content);
              identity = event;
              if (event.name !== sub.event || (event.visibility === 'domain' && event.context !== sub.owner) ||
                !validProperties(message.properties, event)) throw new Error('Invalid delivery metadata');
              payload = event.payload;
              m = {id: derivedId(sub.consumer, event.id), target: sub.target(payload),
                name: sub.consumer, correlation: event.correlationId, causation: event.id, input: event.payload,
                consumer: sub.consumer, sourceId: event.id, sourceHash: createHash('sha256').update(message.content).digest('hex')};
            }
            validating = false;
            const outcome = await sub.handle(m, payload);
            if (outcome.rejection) throw new Rejection(outcome.rejection.code, outcome.rejection.message);
          } catch (error) {
            const headers = {...message.properties.headers};
            delete headers['x-death'];
            const suffix = validating || error instanceof Rejection || attempt >= 3 ? '.dead' : '.retry';
            headers['ref-attempt'] = {'!': 'int32', value: attempt + 1};
            headers['ref-failure'] = errorClass(error);
            await confirmed(channel, 'ref.'+sub.owner+'.delivery', queue+suffix, message.content,
              {...message.properties, headers});
            failure(sub.owner, sub.consumer, error, identity?.id, identity?.correlationId, suffix === '.dead');
          }
          channel.ack(message);
        })().catch(error => { failure(sub.owner, sub.consumer+'.transfer', error); void current.close().catch(() => {}); });
        inFlight.add(work);
        void work.finally(() => inFlight.delete(work));
      }, {noAck: false});
      await closed;
      signal.removeEventListener('abort', abort);
    } catch (error) { failure(sub.owner, sub.consumer+'.reconnect', error); await pause(signal, 1000); }
    finally { await Promise.allSettled(inFlight); await connection?.close().catch(() => {}); }
  }
}
