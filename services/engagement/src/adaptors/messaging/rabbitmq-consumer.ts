import { confirmed } from './rabbitmq-publisher.js';
import { failure, errorClass } from '../diagnostics.js';

import amqp, { type Options } from 'amqplib';

import { createHash } from 'node:crypto';

import { decode, type WireEvent } from './event-codec.js';

import { pause } from '../../foundation/abortable-delay.js';

import { derivedId } from '../../foundation/identity.js';

import type { Metadata, Outcome } from '../../foundation/application.js';

import { Rejection } from '../../foundation/domain.js';

export type EventSubscription = {
  owner: string;
  consumer: string;
  sourceTarget?: boolean;
  event: string;
  target(payload: object): string;
  handle(metadata: Metadata, payload: object): Promise<Outcome>;
  decodeCommand?: (body: Buffer) => { metadata: Metadata; payload: object };
};

function validProperties(properties: Options.Publish, event: WireEvent) {
  return (
    properties.contentType === 'application/x-protobuf' &&
    properties.deliveryMode === 2 &&
    properties.messageId === event.id &&
    properties.type === event.name &&
    properties.appId === event.context &&
    properties.correlationId === event.correlationId &&
    (properties.headers as Record<string, unknown> | undefined)?.['contract-version'] === 1
  );
}

export async function consume(url: string, sub: EventSubscription, signal: AbortSignal) {
  const queue = 'ref.' + sub.consumer + (sub.decodeCommand ? '.command' : '');
  const inFlight = new Set<Promise<void>>();
  while (!signal.aborted) {
    let connection: Awaited<ReturnType<typeof amqp.connect>> | undefined;
    try {
      connection = await amqp.connect(url);
      const current = connection;
      const channel = await current.createConfirmChannel();
      channel.on('error', (error) => failure(sub.owner, sub.consumer, error));
      current.on('error', (error) => failure(sub.owner, sub.consumer, error));
      const closed = new Promise<void>((resolve) => current.once('close', resolve));
      const abort = () => {
        void current.close().catch(() => {});
      };
      signal.addEventListener('abort', abort, { once: true });
      if (signal.aborted) {
        abort();
        return;
      }
      await channel.prefetch(1);
      await channel.consume(
        queue,
        (message) => {
          if (!message) {
            void current.close().catch(() => {});
            return;
          }
          const work = (async () => {
            let validating = true;
            let identity: { id: string; correlationId: string } | undefined;
            const incomingHeaders: Record<string, unknown> = message.properties.headers ?? {};
            const counter = 'ref-attempt' in incomingHeaders ? incomingHeaders['ref-attempt'] : 0;
            const validAttempt =
              typeof counter === 'number' && Number.isInteger(counter) && counter >= 0 && counter <= 3;
            const attempt = validAttempt ? counter : 0;
            try {
              if (!validAttempt) throw new Error('Invalid retry counter');
              let m: Metadata, payload: object;
              if (sub.decodeCommand) {
                ({ metadata: m, payload } = sub.decodeCommand(message.content));
                identity = { id: m.id, correlationId: m.correlation };
                const p = message.properties;
                if (
                  p.contentType !== 'application/x-protobuf' ||
                  p.deliveryMode !== 2 ||
                  p.messageId !== m.id ||
                  p.type !== sub.consumer + '.command' ||
                  p.appId !== sub.owner ||
                  p.correlationId !== m.correlation ||
                  p.headers?.['contract-version'] !== 1
                )
                  throw new Error('Invalid command properties');
              } else {
                const event = decode(message.content);
                identity = event;
                if (
                  event.name !== sub.event ||
                  (event.visibility === 'domain' && event.context !== sub.owner) ||
                  !validProperties(message.properties, event)
                )
                  throw new Error('Invalid delivery metadata');
                payload = event.payload;
                m = {
                  id: derivedId(sub.consumer, event.id),
                  target: sub.sourceTarget ? event.aggregateId : sub.target(payload),
                  name: sub.consumer,
                  correlation: event.correlationId,
                  causation: event.id,
                  input: event.payload,
                  consumer: sub.consumer,
                  sourceId: event.id,
                  sourceHash: createHash('sha256').update(message.content).digest('hex'),
                  sourceKind: event.aggregateKind,
                  sourceVersion: event.aggregateVersion,
                  sourceTime: event.occurredAt,
                };
              }
              validating = false;
              const outcome = await sub.handle(m, payload);
              if (outcome.rejection) throw new Rejection(outcome.rejection.code, outcome.rejection.message);
            } catch (error) {
              const headers = { ...message.properties.headers };
              delete headers['x-death'];
              const suffix = validating || error instanceof Rejection || attempt >= 3 ? '.dead' : '.retry';
              headers['ref-attempt'] = { '!': 'int32', value: attempt + 1 };
              headers['ref-failure'] = errorClass(error);
              await confirmed(channel, 'ref.' + sub.owner + '.delivery', queue + suffix, message.content, {
                ...message.properties,
                headers,
              });
              failure(
                sub.owner,
                sub.consumer,
                error,
                identity?.id,
                identity?.correlationId,
                suffix === '.dead',
              );
            }
            channel.ack(message);
          })().catch((error) => {
            failure(sub.owner, sub.consumer + '.transfer', error);
            void current.close().catch(() => {});
          });
          inFlight.add(work);
          void work.finally(() => inFlight.delete(work));
        },
        { noAck: false },
      );
      await closed;
      signal.removeEventListener('abort', abort);
    } catch (error) {
      failure(sub.owner, sub.consumer + '.reconnect', error);
      await pause(signal, 1000);
    } finally {
      await Promise.allSettled(inFlight);
      await connection?.close().catch(() => {});
    }
  }
}
