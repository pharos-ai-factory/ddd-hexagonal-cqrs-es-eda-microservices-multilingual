import {test} from 'node:test';
import assert from 'node:assert/strict';
import {readFileSync} from 'node:fs';
import {decode, encode} from './codec.js';
import {derivedId} from '../foundation/identity.js';
import protobuf from 'protobufjs';
import schema from './generated/events.json' with {type: 'json'};
test('the private domain fixture is understood without treating it as an integration event', () => {
  const raw = Buffer.from(readFileSync(new URL('../../../../contracts/events/fixtures/reward-earned.v1.hex', import.meta.url), 'utf8').trim(), 'hex');
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
