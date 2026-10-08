import amqp from 'amqplib';
import protobuf from 'protobufjs';
import schema from './generated/requests.json' with {type: 'json'};
import {confirmed} from './broker.js';
import {failure} from './diagnostics.js';
import {currentReply, type ReplyIntent} from './replies.js';
import {pause} from './dispatch.js';
import {identifier as domainID} from '../foundation/domain.js';
import {record, integer, text} from './request-values.js';
const identifier = (value: unknown) => domainID(text(value));
import type {Metadata, Outcome, Loaded} from '../foundation/application.js';
import type {PagedQueryPort} from '../foundation/pagination.js';

const root = protobuf.Root.fromJSON(schema);
const requestType = (owner: string) => root.lookupType('cafe.'+owner+'.requests.v1.Request');
const replyType = (owner: string) => root.lookupType('cafe.'+owner+'.requests.v1.Reply');
type Request = {kind: string; id: string; owner: string; name: string; body: Record<string, unknown>; metadata?: Metadata};
type Handler = (request: Request) => Promise<Record<string, unknown>>;

function selected(value: Record<string, unknown>): [string, Record<string, unknown>] {
  const name = text(value.payload);
  return [name, record(value[name])];
}
export function decodeRequest(body: Buffer, owner: string): Request {
  if (body.length > 65536) throw new Error('Request is too large');
  const wire = record(requestType(owner).toObject(requestType(owner).decode(body), {longs: Number, oneofs: true}));
  if (wire.contractVersion !== 1 || wire.context !== owner) throw new Error('Invalid request envelope');
  const id = identifier(wire.requestId);
  const [kind, outer] = selected(wire), [name, payload] = selected(outer);
  const envelope = root.lookupType('cafe.'+owner+'.requests.v1.'+(kind === 'command' ? 'Command' : 'Query'));
  const field = envelope.fields[name];
  const owned: Record<string, string[]> = {
    loyalty: ['redeemReward', 'listAccounts', 'getAccount', 'listRewards', 'getReward'],
    communication: ['listNotifications', 'getNotification'],
  };
  if (!field || !owned[owner]?.includes(name)) throw new Error('Foreign context request');
  const request: Request = {kind, id, owner, name, body: payload};
  if (kind === 'command') {
    const metadata = record(outer.metadata);
    const expected = integer(metadata.expectedVersion);
    if (expected < 0 || !Number.isSafeInteger(expected)) throw new Error('Invalid expected version');
    request.metadata = {id: identifier(metadata.commandId), target: identifier(metadata.aggregateId),
      name: owner+'.'+name[0]!.toUpperCase()+name.slice(1), correlation: identifier(metadata.correlationId), expected, input: payload};
    if (!(field?.resolvedType instanceof protobuf.Type)) throw new Error('Unknown command type');
    requiredInputs(field.resolvedType, payload);
  } else if (kind === 'query') {
    if ('id' in payload) identifier(payload.id);
    if ('page' in payload) {
      const page = record(payload.page), limit = integer(page.limit);
      if (limit < 1 || limit > 100) throw new Error('Invalid page limit');
      if ('after' in page) identifier(page.after);
    }
  } else throw new Error('Unknown request kind');
  return request;
}
export class Registry {
  private readonly handlers = new Map<string, Handler>();
  constructor(readonly owner: string) {}
  command<I extends object>(name: string, parse: (value: unknown) => I,
    execute: (metadata: Metadata, value: I) => Promise<Outcome>) {
    this.handlers.set(name, async request => {
      if (!request.metadata) throw new Error('Command metadata is required');
      const input = parse(request.body);
      const result = await execute({...request.metadata, input}, input);
      return {outcome: {aggregateId: result.aggregateId, version: result.version, status: result.status,
        ...(result.rejection ? {rejection: {code: result.rejection.code, message: result.rejection.message}} : {})}};
    });
  }
  queries<S>(singular: string, plural: string, queries: PagedQueryPort<S>, map: (value: Loaded<S>) => object) {
    const title = (value: string) => value[0]!.toUpperCase()+value.slice(1);
    this.handlers.set('get'+title(singular), async request => {
      const loaded = await queries.get(identifier(request.body.id));
      return loaded ? {[singular]: map(loaded)} : {error: {code: 'not_found'}};
    });
    this.handlers.set('list'+title(plural), async request => {
      if ('page' in request.body) {
        const page = record(request.body.page);
        const result = await queries.page({limit: integer(page.limit), after: 'after' in page ? identifier(page.after) : undefined});
        return {[plural]: {items: result.items.map(map), paged: true, ...(result.nextId ? {nextId: result.nextId} : {})}};
      }
      return {[plural]: {items: (await queries.list()).map(map), paged: false}};
    });
  }
  async handle(request: Request): Promise<Buffer> {
    if (request.owner !== this.owner) throw new Error('Foreign context request');
    const handler = this.handlers.get(request.name);
    if (!handler) throw new Error('Unknown owner request');
    let result: Record<string, unknown>;
    try { result = await handler(request); }
    catch (error) { failure(this.owner, 'requests.handle', error, request.id); result = {error: {code: 'temporarily_unavailable'}}; }
    const reply = replyType(this.owner).fromObject({contractVersion: 1, requestId: request.id, context: this.owner, ...result});
    return Buffer.from(replyType(this.owner).encode(reply).finish());
  }
  async run(url: string, signal: AbortSignal) {
    await Promise.all([this.runKind(url,signal,'command'),this.runKind(url,signal,'query')]);
  }
  private async runKind(url: string, signal: AbortSignal, kind: 'command'|'query') {
    const queue = 'ref.'+this.owner+(kind === 'command' ? '.commands' : '.queries');
    while (!signal.aborted) {
      let connection: Awaited<ReturnType<typeof amqp.connect>> | undefined;
      try {
        connection = await amqp.connect(url);
        const current = connection;
        const channel = await current.createConfirmChannel();
        current.on('error', error => failure(this.owner, 'requests.connection', error));
        channel.on('error', error => failure(this.owner, 'requests.channel', error));
        const closed = new Promise<void>(resolve => current.once('close', resolve));
        const abort = () => { void current.close().catch(() => {}); };
        signal.addEventListener('abort', abort, {once: true});
        await channel.prefetch(1);
        await channel.consume(queue, message => {
          if (!message) { void current.close().catch(() => {}); return; }
          void (async () => {
            let request: Request;
            try {
              const properties = message.properties;
              if (properties.contentType !== 'application/x-protobuf' || properties.deliveryMode !== 2 ||
                properties.type !== kind || properties.appId !== 'api' || properties.messageId !== properties.correlationId ||
                message.fields.routingKey !== 'request.'+this.owner+'.'+kind) throw new Error('Invalid request properties');
              request = decodeRequest(message.content, this.owner);
              if (request.kind !== kind || request.id !== properties.messageId || !this.handlers.has(request.name)) throw new Error('Invalid request identity or owner');
            } catch (error) {
              await confirmed(channel, 'ref.'+this.owner+'.delivery', queue+'.dead', message.content, message.properties);
              failure(this.owner, 'requests.validate', error, undefined, undefined, true);
              channel.ack(message); return;
            }
            const intent: ReplyIntent = {id: request.id, committed: false, encode: outcome => {
              const result = {aggregateId: outcome.aggregateId, version: outcome.version, status: outcome.status,
                ...(outcome.rejection ? {rejection: {code: outcome.rejection.code, message: outcome.rejection.message}} : {})};
              const type = replyType(this.owner);
              return Buffer.from(type.encode(type.fromObject({contractVersion: 1, requestId: request.id, context: this.owner, outcome: result})).finish());
            }};
            const reply = kind === 'command' ? await currentReply.run(intent, () => this.handle(request)) : await this.handle(request);
            if (intent.committed) { channel.ack(message); return; }
            await confirmed(channel, 'cafe.replies', 'reply.'+this.owner, reply, {contentType: 'application/x-protobuf',
              type: 'reply', appId: this.owner, messageId: request.id, correlationId: request.id});
            channel.ack(message);
          })().catch(error => { failure(this.owner, 'requests.reply', error); void current.close().catch(() => {}); });
        }, {noAck: false});
        await closed;
        signal.removeEventListener('abort', abort);
      } catch (error) { failure(this.owner, 'requests.reconnect', error); await pause(signal, 1000); }
      finally { await connection?.close().catch(() => {}); }
    }
  }
}

export function requiredInputs(type: protobuf.Type, payload: Record<string, unknown>) {
  for (const [property, rule] of Object.entries(type.fields)) {
    if (rule.options?.['(cafe.requests.v1.required_input)'] && !(property in payload)) throw new Error('Missing command field');
  }
}
