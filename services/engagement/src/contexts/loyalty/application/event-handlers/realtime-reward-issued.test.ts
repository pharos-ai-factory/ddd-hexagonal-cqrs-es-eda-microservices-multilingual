import { test } from 'node:test';
import assert from 'node:assert/strict';
import { RewardIssuedRealtimeIntegrationEventHandler } from './realtime-reward-issued.js';
import type { RewardReadRepository } from '../ports/reward-read-repository.js';
import type { RewardView } from '../read-models/reward.js';
import type { Loaded, Metadata, RealtimeEvent } from '../../../../foundation/application.js';

void test('delayed integration delivery publishes the current view revision and requires its existence', async () => {
  const id = '00000000-0000-4000-8000-000000000001';
  const state: RewardView = {
    id,
    grantId: id,
    customerId: id,
    benefit: 'Tea',
    status: 'redeemed',
    redeemedFor: id,
    expiresAt: '2027-01-01T12:00:00.000Z',
  };
  let loaded: Loaded<RewardView> | undefined = { exists: true, state, version: 4 };
  const reader: RewardReadRepository = {
    get: async () => loaded,
    list: async () => assert.fail('Unexpected list'),
    page: async () => assert.fail('Unexpected page'),
  };
  const events: RealtimeEvent[] = [];
  const handler = new RewardIssuedRealtimeIntegrationEventHandler(reader, {
    publish: async (event) => {
      events.push(event);
    },
  });
  const source: Metadata = {
    id,
    target: id,
    name: 'reward-issued',
    correlation: id,
    input: {},
    sourceVersion: 2,
  };
  const event = { rewardId: id, customerId: id, benefit: 'Tea', expiresAt: state.expiresAt };
  assert.equal((await handler.handle(source, event)).version, 4);
  assert.equal(events[0]!.revision, 4);
  assert.equal(events[0]!.source, source);
  assert.deepEqual(events[0]!.snapshot, state);
  loaded = undefined;
  await assert.rejects(handler.handle(source, event), /Committed reward is missing/);
  assert.equal(events.length, 1);
});
