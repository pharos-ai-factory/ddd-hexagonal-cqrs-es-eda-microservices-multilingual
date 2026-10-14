import { LoyaltyDomainEvent, LoyaltyIntegrationEvent } from './generated/message-names.js';
import { test } from 'node:test';
import assert from 'node:assert/strict';
import { EventRecordingWriteRepository } from './command-execution.js';

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
