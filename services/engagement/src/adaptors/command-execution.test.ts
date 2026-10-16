import { LoyaltyDomainEvent, LoyaltyIntegrationEvent } from './generated/message-names.js';
import { test } from 'node:test';
import assert from 'node:assert/strict';
import { bindCommand, EventRecordingWriteRepository } from './command-execution.js';

void test('aggregate saves cannot bypass their domain publication subscription', async () => {
  let saved = false;
  const source = {
    get: () => Promise.resolve(undefined),
    save: () => {
      saved = true;
      return Promise.resolve();
    },
  };
  const publicRepository = new EventRecordingWriteRepository(source, () => [
    { name: LoyaltyIntegrationEvent.REWARD_ISSUED, payload: {} },
  ]);
  await assert.rejects(publicRepository.save(1), /private domain/);
  assert.equal(saved, false);
  const privateRepository = new EventRecordingWriteRepository(source, () => [
    { name: LoyaltyDomainEvent.REWARD_ISSUED, payload: {} },
  ]);
  await privateRepository.save(1);
  assert.equal(saved, true);
  assert.equal(privateRepository.publications.length, 1);
});

void test('handler receives the authenticated actor with its target', async () => {
  for (const actorId of [undefined, '11111111-1111-4111-8111-111111111111']) {
    const executor = bindCommand<number, number>(
      {
        execute: async (metadata, work) => {
          const result = await work({ get: async () => undefined, save: async () => {} });
          return { aggregateId: metadata.target, version: 1, status: result.status };
        },
      },
      () => ({
        execute: async (context) => {
          assert.equal(context.target, 'target');
          assert.equal(context.actorId, actorId);
          return { status: 'accepted' };
        },
      }),
    );
    const outcome = await executor.execute(
      {
        id: 'id',
        target: 'target',
        name: 'command',
        correlation: 'correlation',
        input: {},
        ...(actorId ? { actorId } : {}),
      },
      0,
    );
    assert.equal(outcome.status, 'accepted');
  }
});
