import protobuf from 'protobufjs';
import type {Metadata, Outcome, DurableCommandPort} from '../foundation/application.js';
import {identifier} from '../foundation/domain.js';
import {derivedId} from '../foundation/identity.js';
import type {PostgresContextDatabase} from './postgres.js';

/** Serialises an owner-private command and preserves its originating receipt material. */
export class InternalCommandCodec<C extends object> {
  private envelope: protobuf.Type;
  constructor(readonly owner: string, readonly consumer: string, readonly command: string,
    schema: protobuf.INamespace, private parse: (value: unknown) => C) {
    if (!consumer.startsWith(owner+'.')) throw new Error('Foreign command subscription');
    this.envelope = protobuf.Root.fromJSON(schema).lookupType(`cafe.internal.${owner}.CommandEnvelope`);
    if (!this.envelope.fields[command]) throw new Error('Unknown owner command');
  }
  encode(m: Metadata, command: C): Buffer {
    this.validate(m);
    const value = {id: m.id, consumer: m.consumer, sourceEventId: m.sourceId, sourceHash: m.sourceHash,
      target: m.target, correlationId: m.correlation, receiptMaterial: Buffer.from(JSON.stringify(m.input)),
      [this.command]: this.parse(command)};
    const error = this.envelope.verify(value);
    if (error) throw new Error(error);
    return Buffer.from(this.envelope.encode(this.envelope.create(value)).finish());
  }
  decode(body: Buffer): {metadata: Metadata; payload: C} {
    const envelope = this.envelope.toObject(this.envelope.decode(body), {oneofs: true}) as Record<string, unknown>;
    if (envelope.command !== this.command || !Buffer.isBuffer(envelope.receiptMaterial)) throw new Error('Wrong command payload');
    const metadata: Metadata = {id: identifier(String(envelope.id)), consumer: String(envelope.consumer),
      sourceId: identifier(String(envelope.sourceEventId)), sourceHash: String(envelope.sourceHash),
      target: identifier(String(envelope.target)), correlation: identifier(String(envelope.correlationId)),
      name: this.consumer, causation: identifier(String(envelope.sourceEventId)),
      input: JSON.parse(envelope.receiptMaterial.toString('utf8')) as unknown};
    this.validate(metadata);
    return {metadata, payload: this.parse(envelope[this.command])};
  }
  private validate(m: Metadata) {
    [m.id, m.target, m.correlation, m.sourceId].forEach(value => identifier(String(value)));
    if (m.consumer !== this.consumer || m.id !== derivedId(this.consumer, m.sourceId!) ||
      !/^[a-f0-9]{64}$/.test(m.sourceHash ?? '')) throw new Error('Invalid command identity');
  }
}

/** Commits the event receipt and immutable command before allowing the event ACK. */
export class PostgresDurableCommandOutbox<C extends object> implements DurableCommandPort<C> {
  constructor(private db: PostgresContextDatabase, private codec: InternalCommandCodec<C>) {
    if (db.owner !== codec.owner) throw new Error('Command outbox crossed its owner');
  }
  async enqueue(m: Metadata, command: C): Promise<Outcome> {
    const body = this.codec.encode(m, command);
    const client = await this.db.pool.connect();
    try {
      await client.query('BEGIN');
      await client.query('SELECT pg_advisory_xact_lock(hashtextextended($1,0))', ['handoff:'+m.consumer+':'+m.sourceId]);
      const {rows: [previous]} = await client.query(
        'SELECT id,fingerprint,target FROM cafe.internal_commands WHERE consumer=$1 AND event_id=$2', [m.consumer, m.sourceId]);
      if (previous) {
        if (previous.id !== m.id || previous.fingerprint !== m.sourceHash || previous.target !== m.target)
          throw new Error('Conflicting event hand-off identity');
      } else {
        await client.query(`INSERT INTO cafe.internal_commands(id,consumer,event_id,fingerprint,target,body)
          VALUES($1,$2,$3,$4,$5,$6)`, [m.id, m.consumer, m.sourceId, m.sourceHash, m.target, body]);
        await client.query('INSERT INTO cafe.internal_command_dispatches(event_id) VALUES($1)', [m.id]);
      }
      await client.query('COMMIT');
      return {aggregateId: m.target, version: 0, status: 'queued'};
    } catch (error) { await client.query('ROLLBACK').catch(() => {}); throw error; }
    finally { client.release(); }
  }
}

/** Header reader shared by the command relay; typed payload checks belong to its receiver. */
export function commandPublication(schema: protobuf.INamespace, owner: string) {
  const type = protobuf.Root.fromJSON(schema).lookupType(`cafe.internal.${owner}.CommandEnvelope`);
  return (body: Buffer) => {
    const value = type.toObject(type.decode(body)) as {id: string; consumer: string; correlationId: string};
    identifier(value.id); identifier(value.correlationId);
    if (!value.consumer.startsWith(owner+'.')) throw new Error('Foreign command destination');
    return {id: value.id, name: value.consumer+'.command', context: owner, correlationId: value.correlationId,
      visibility: 'command', payload: {}};
  };
}
