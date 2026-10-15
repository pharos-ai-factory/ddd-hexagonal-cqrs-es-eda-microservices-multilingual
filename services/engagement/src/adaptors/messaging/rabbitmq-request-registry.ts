import { toOutcome } from './generated/response-payloads.js';
import { replyType, decodeRequest, type Request } from './request-envelope.js';
import amqp from 'amqplib';

import { confirmed } from './rabbitmq-publisher.js';

import { failure } from '../diagnostics.js';

import { currentReply, type ReplyIntent } from '../persistence/command-reply-outbox.js';

import { pause } from '../../foundation/abortable-delay.js';

import { COMMAND_NAMES, QUERY_NAMES, type CommandName, type QueryName } from '../generated/request-names.js';

import type { Metadata, Outcome, QueryBusPort } from '../../foundation/application.js';

type Handler = (request: Request) => Promise<Record<string, unknown>>;

/** Maps owner RPC requests to typed application handlers and manages durable replies. */
export class RabbitMQRequestRegistry {
  private readonly handlers = new Map<string, Handler>();
  constructor(readonly owner: string) {}
  command<I extends object>(
    name: CommandName,
    parse: (value: unknown) => I,
    execute: (metadata: Metadata, value: I) => Promise<Outcome>,
  ) {
    if (!COMMAND_NAMES[this.owner]?.includes(name))
      throw new Error('Command identifier disagrees with owner');
    if (this.handlers.has(name)) throw new Error('Duplicate request handler: ' + name);
    this.handlers.set(name, async (request) => {
      if (!request.metadata) throw new Error('Command metadata is required');
      const input = parse(request.body);
      const result = await execute({ ...request.metadata, input }, input);
      return {
        outcome: toOutcome(result),
      };
    });
  }
  /** Connect one generated query identifier to its parser, plain handler and reply mapper. */
  query<Q, R>(
    name: QueryName,
    parse: (value: unknown) => Q,
    handler: { execute(query: Q): Promise<R> },
    reply: (query: Q, result: R) => Record<string, unknown>,
  ) {
    if (!QUERY_NAMES[this.owner]?.includes(name)) throw new Error('Query identifier disagrees with owner');
    if (this.handlers.has(name)) throw new Error('Duplicate request handler: ' + name);
    const bus: QueryBusPort<Q, R> = { ask: (query) => handler.execute(query) };
    this.handlers.set(name, async (request) => {
      const query = parse(request.body);
      return reply(query, await bus.ask(query));
    });
  }
  async handle(request: Request): Promise<Buffer> {
    if (request.owner !== this.owner) throw new Error('Foreign context request');
    const handler = this.handlers.get(request.name);
    if (!handler) throw new Error('Unknown owner request');
    let result: Record<string, unknown>;
    try {
      result = await handler(request);
    } catch (error) {
      failure(this.owner, 'requests.handle', error, request.id);
      result = { error: { code: 'temporarily_unavailable' } };
    }
    const reply = replyType(this.owner).fromObject({
      contractVersion: 1,
      requestId: request.id,
      context: this.owner,
      ...result,
    });
    return Buffer.from(replyType(this.owner).encode(reply).finish());
  }
  async run(url: string, signal: AbortSignal) {
    await Promise.all([this.runKind(url, signal, 'command'), this.runKind(url, signal, 'query')]);
  }
  private async runKind(url: string, signal: AbortSignal, kind: 'command' | 'query') {
    const inFlight = new Set<Promise<void>>();
    const queue = 'ref.' + this.owner + (kind === 'command' ? '.commands' : '.queries');
    while (!signal.aborted) {
      let connection: Awaited<ReturnType<typeof amqp.connect>> | undefined;
      try {
        connection = await amqp.connect(url);
        const current = connection;
        const channel = await current.createConfirmChannel();
        current.on('error', (error) => failure(this.owner, 'requests.connection', error));
        channel.on('error', (error) => failure(this.owner, 'requests.channel', error));
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
              let request: Request;
              try {
                const properties = message.properties;
                if (
                  properties.contentType !== 'application/x-protobuf' ||
                  properties.deliveryMode !== 2 ||
                  properties.type !== kind ||
                  properties.appId !== 'api' ||
                  properties.messageId !== properties.correlationId ||
                  message.fields.routingKey !== 'request.' + this.owner + '.' + kind
                )
                  throw new Error('Invalid request properties');
                request = decodeRequest(message.content, this.owner);
                if (
                  request.kind !== kind ||
                  request.id !== properties.messageId ||
                  !this.handlers.has(request.name)
                )
                  throw new Error('Invalid request identity or owner');
              } catch (error) {
                await confirmed(
                  channel,
                  'ref.' + this.owner + '.delivery',
                  queue + '.dead',
                  message.content,
                  message.properties,
                );
                failure(this.owner, 'requests.validate', error, undefined, undefined, true);
                channel.ack(message);
                return;
              }
              const intent: ReplyIntent = {
                id: request.id,
                committed: false,
                encode: (outcome) => {
                  const result = toOutcome(outcome);
                  const type = replyType(this.owner);
                  return Buffer.from(
                    type
                      .encode(
                        type.fromObject({
                          contractVersion: 1,
                          requestId: request.id,
                          context: this.owner,
                          outcome: result,
                        }),
                      )
                      .finish(),
                  );
                },
              };
              const reply =
                kind === 'command'
                  ? await currentReply.run(intent, () => this.handle(request))
                  : await this.handle(request);
              if (intent.committed) {
                channel.ack(message);
                return;
              }
              await confirmed(channel, 'cafe.replies', 'reply.' + this.owner, reply, {
                contentType: 'application/x-protobuf',
                type: 'reply',
                appId: this.owner,
                messageId: request.id,
                correlationId: request.id,
              });
              channel.ack(message);
            })().catch((error) => {
              failure(this.owner, 'requests.reply', error);
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
        failure(this.owner, 'requests.reconnect', error);
        await pause(signal, 1000);
      } finally {
        await Promise.allSettled(inFlight);
        await connection?.close().catch(() => {});
      }
    }
  }
}
