import {test} from 'node:test';
import assert from 'node:assert/strict';
import {readFileSync} from 'node:fs';
import {decode, encode} from './codec.js';
import {derivedId} from '../foundation/identity.js';
import protobuf from 'protobufjs';
import schema from './generated/events.json' with {type: 'json'};
import privateSchema from '../contexts/loyalty/adaptors/messaging/generated/private_messages.json' with {type: 'json'};
test('the owner-local private message retains its historical bytes and scope', () => {
  const raw = Buffer.from(readFileSync(new URL('../contexts/loyalty/adaptors/messaging/fixtures/reward-earned.hex', import.meta.url), 'utf8').trim(), 'hex');
  const type = protobuf.Root.fromJSON(privateSchema).lookupType('cafe.loyalty.internal.PrivateEvent');
  assert.deepEqual(Buffer.from(type.encode(type.decode(raw)).finish()), raw);
  const event = decode(raw);
  assert.equal(event.context, 'loyalty'); assert.equal(event.visibility, 'domain');
  assert.equal(event.name, 'loyalty.reward-earned');
  assert.equal((event.payload as {validDays: number}).validDays, 7);
  assert.throws(() => encode('communication', 'account', event.aggregateId, 1,
    {id: event.id, target: event.aggregateId, correlation: event.correlationId, name: 'issue', input: {}},
    {name: event.name, payload: event.payload}), /owner/);
  assert.equal(derivedId('earned-grant', '00000000-0000-4000-8000-000000000001'),
    'f4136e57-6728-8a37-b8f6-a29159aecdc3');
});

test('the shared integration fixture decodes consistently across languages', () => {
  const raw = Buffer.from(readFileSync(new URL('../../../../contracts/loyalty/messaging/integration_events/v1/fixtures/reward-issued.v1.hex', import.meta.url), 'utf8').trim(), 'hex');
  const event = decode(raw);
  assert.equal(event.name, 'loyalty.reward-issued');
  assert.equal(event.visibility, 'integration');
  assert.deepEqual(event.payload, {
    rewardId: '22222222-2222-4222-8222-222222222222',
    customerId: '55555555-5555-4555-8555-555555555555',
    benefit: 'one free drink', expiresAt: '2026-10-03T12:00:00Z',
  });
});

function rewardIssuedBytes(payload: object) {
  const type = protobuf.Root.fromJSON(schema).lookupType('cafe.v1.Event');
  const id = '00000000-0000-4000-8000-000000000001';
  const envelope = {id, name: 'loyalty.reward-issued', context: 'loyalty', visibility: 'integration',
    contractVersion: 1, aggregateKind: 'reward', aggregateId: id, aggregateVersion: 1,
    correlationId: id, causationId: id, occurredAt: '2026-09-27T12:00:00Z'};
  return type.encode(type.create({...envelope, rewardIssued: payload})).finish();
}
const rewardIssued = {rewardId: '00000000-0000-4000-8000-000000000001',
  customerId: '00000000-0000-4000-8000-000000000001', benefit: 'one free drink', expiresAt: '2026-10-04T12:00:00Z'};
test('a complete RewardIssued reaches the application unchanged', () => {
  assert.deepEqual(decode(rewardIssuedBytes(rewardIssued)).payload, rewardIssued);
});
for (const field of Object.keys(rewardIssued)) {
  test(`RewardIssued rejects an omitted ${field} before reaching an application`, () => {
    const incomplete = {...rewardIssued} as Record<string, unknown>;
    delete incomplete[field];
    assert.throws(() => decode(rewardIssuedBytes(incomplete)), `Missing ${field} must be quarantined`);
  });
}
for (const [name, invalid] of [
  ['source reward', {...rewardIssued, rewardId: '00000000-0000-4000-8000-000000000002'}],
  ['empty benefit', {...rewardIssued, benefit: ''}],
  ['expiry', {...rewardIssued, expiresAt: 'not a date'}],
] as const) {
  test(`RewardIssued rejects an invalid ${name}`, () => assert.throws(() => decode(rewardIssuedBytes(invalid))));
}
