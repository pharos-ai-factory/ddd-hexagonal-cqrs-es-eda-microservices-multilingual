import {test} from 'node:test';
import assert from 'node:assert/strict';
import {randomUUID} from 'node:crypto';
import {CorruptState} from '../foundation/domain.js';
import {restoreAccount, restoreNotification, restoreReward} from './restore.js';

test('restored snapshots are validated before they reach a handler or query', () => {
  const id = randomUUID();
  const account = {id, collections: 3, grantsEarned: 1, stampBalance: 0,
    lastGrant: {id: randomUUID(), accountId: id, benefit: 'one free drink', validDays: 7}};
  assert.deepEqual(restoreAccount(account), account);
  assert.throws(() => restoreAccount({...account, lastGrant: {...account.lastGrant, accountId: randomUUID()}}), CorruptState);
  assert.throws(() => restoreReward({id, grantId: randomUUID(), customerId: 'invalid', benefit: 'coffee',
    status: 'issued', expiresAt: '2026-10-01T00:00:00Z'}), CorruptState);
  assert.throws(() => restoreNotification({id, recipient: randomUUID(), subject: 'Ready', body: 'Collect',
    status: 'requested', providerReceipt: 'unexpected'}), CorruptState);
});
