import { LoyaltyDomainEvent, LoyaltyIntegrationEvent } from '../generated/message-names.js';
import { test } from 'node:test';
import assert from 'node:assert/strict';
import { readFileSync } from 'node:fs';
import { decode, encode } from './event-codec.js';
import { derivedId } from '../../foundation/identity.js';
import protobuf from 'protobufjs';
import schema from '../generated/events.json' with { type: 'json' };
import privateSchema from '../../contexts/loyalty/adaptors/messaging/generated/private_messages.json' with { type: 'json' };
void test('the owner-local private message retains its historical bytes and scope', () => {
  const raw = Buffer.from(
    readFileSync(
      new URL('../../contexts/loyalty/adaptors/messaging/fixtures/reward-earned.hex', import.meta.url),
      'utf8',
    ).trim(),
    'hex',
  );
  const type = protobuf.Root.fromJSON(privateSchema).lookupType('cafe.loyalty.internal.PrivateEvent');
  assert.deepEqual(Buffer.from(type.encode(type.decode(raw)).finish()), raw);
  const event = decode(raw);
  assert.equal(event.context, 'loyalty');
  assert.equal(event.visibility, 'domain');
  assert.equal(event.name, LoyaltyDomainEvent.REWARD_EARNED);
  assert.equal((event.payload as { validDays: number }).validDays, 7);
  assert.throws(
    () =>
      encode(
        'communication',
        'account',
        event.aggregateId,
        1,
        {
          id: event.id,
          target: event.aggregateId,
          correlation: event.correlationId,
          name: 'issue',
          input: {},
        },
        { name: event.name, payload: event.payload },
      ),
    /owner/,
  );
  assert.equal(
    derivedId('earned-grant', '00000000-0000-4000-8000-000000000001'),
    'f4136e57-6728-8a37-b8f6-a29159aecdc3',
  );
});

void test('the shared integration fixture decodes consistently across languages', () => {
  const raw = Buffer.from(
    readFileSync(
      new URL(
        '../../../../../contracts/loyalty/messaging/integration_events/v1/fixtures/reward-issued.v1.hex',
        import.meta.url,
      ),
      'utf8',
    ).trim(),
    'hex',
  );
  const event = decode(raw);
  assert.equal(event.name, LoyaltyIntegrationEvent.REWARD_ISSUED);
  assert.equal(event.visibility, 'integration');
  assert.deepEqual(event.payload, {
    rewardId: '22222222-2222-4222-8222-222222222222',
    customerId: '55555555-5555-4555-8555-555555555555',
    benefit: 'one free drink',
    expiresAt: '2026-10-03T12:00:00Z',
  });
});

function rewardIssuedBytes(payload: object) {
  const type = protobuf.Root.fromJSON(schema).lookupType('cafe.v1.Event');
  const id = '00000000-0000-4000-8000-000000000001';
  const envelope = {
    id,
    name: LoyaltyIntegrationEvent.REWARD_ISSUED,
    context: 'loyalty',
    visibility: 'integration',
    contractVersion: 1,
    aggregateKind: 'reward',
    aggregateId: id,
    aggregateVersion: 1,
    correlationId: id,
    causationId: id,
    occurredAt: '2026-09-27T12:00:00Z',
  };
  const payloadType = protobuf.Root.fromJSON(schema).lookupType('cafe.v1.RewardIssued');
  return type
    .encode(type.create({ ...envelope, payload: payloadType.encode(payloadType.create(payload)).finish() }))
    .finish();
}
const rewardIssued = {
  rewardId: '00000000-0000-4000-8000-000000000001',
  customerId: '00000000-0000-4000-8000-000000000001',
  benefit: 'one free drink',
  expiresAt: '2026-10-04T12:00:00Z',
};
void test('a complete RewardIssued reaches the application unchanged', () => {
  assert.deepEqual(decode(rewardIssuedBytes(rewardIssued)).payload, rewardIssued);
});
for (const field of Object.keys(rewardIssued)) {
  void test(`RewardIssued rejects an omitted ${field} before reaching an application`, () => {
    const incomplete = { ...rewardIssued } as Record<string, unknown>;
    delete incomplete[field];
    assert.throws(() => decode(rewardIssuedBytes(incomplete)), `Missing ${field} must be quarantined`);
  });
}
for (const [name, invalid] of [
  ['source reward', { ...rewardIssued, rewardId: '00000000-0000-4000-8000-000000000002' }],
  ['empty benefit', { ...rewardIssued, benefit: '' }],
  ['expiry', { ...rewardIssued, expiresAt: 'not a date' }],
] as const) {
  void test(`RewardIssued rejects an invalid ${name}`, () =>
    assert.throws(() => decode(rewardIssuedBytes(invalid))));
}

void test('published timestamps reject impossible calendar dates', () => {
  assert.throws(() => decode(rewardIssuedBytes({ ...rewardIssued, expiresAt: '2026-02-30T12:00:00Z' })));
});

void test('oversized publications fail before their bytes can be committed', () => {
  const id = rewardIssued.rewardId;
  assert.throws(
    () =>
      encode(
        'loyalty',
        'reward',
        id,
        1,
        { id, target: id, correlation: id, name: 'issue', input: {} },
        {
          name: LoyaltyIntegrationEvent.REWARD_ISSUED,
          payload: { ...rewardIssued, benefit: 'x'.repeat(256 * 1024) },
        },
      ),
    /limit/,
  );
});

void test('private reward-issued fixture retains its owner, source identity and exact bytes', () => {
  const raw = Buffer.from(
    readFileSync(
      new URL('../../contexts/loyalty/adaptors/messaging/fixtures/reward-issued.domain.hex', import.meta.url),
      'utf8',
    ).trim(),
    'hex',
  );
  const event = decode(raw);
  assert.equal(event.name, LoyaltyDomainEvent.REWARD_ISSUED);
  assert.equal(event.visibility, 'domain');
  const type = protobuf.Root.fromJSON(privateSchema).lookupType('cafe.loyalty.internal.PrivateEvent');
  assert.deepEqual(Buffer.from(type.encode(type.decode(raw)).finish()), raw);
  const value = type.toObject(type.decode(raw), { longs: Number }) as {
    context: string;
    formatRevision: number;
    rewardIssued: Record<string, unknown>;
  };
  for (const invalid of [
    { ...value, context: 'communication' },
    { ...value, formatRevision: 2 },
    { ...value, rewardIssued: { ...value.rewardIssued, rewardId: '55555555-5555-4555-8555-555555555555' } },
    { ...value, rewardIssued: { ...value.rewardIssued, benefit: '' } },
    { ...value, rewardIssued: { ...value.rewardIssued, expiresAt: '2026-10-03' } },
  ])
    assert.throws(() => decode(type.encode(type.create(invalid)).finish()));
});
