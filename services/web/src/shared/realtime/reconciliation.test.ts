import { test } from 'node:test';
import assert from 'node:assert/strict';
import { reconciliationQueue } from './reconciliation';
void test('a second history gap during an old query must cause a fresh query', async () => {
  let release!: () => void,
    calls = 0;
  const first = new Promise<void>((resolve) => {
    release = resolve;
  });
  const queue = reconciliationQueue(async () => {
    calls++;
    if (calls === 1) await first;
  });
  const pending = queue.request();
  const second = queue.request(),
    third = queue.request();
  assert.equal(calls, 1);
  release();
  await Promise.all([pending, second, third]);
  assert.equal(calls, 2);
});
void test('a closed window cannot run a queued reconciliation', async () => {
  let release!: () => void,
    calls = 0;
  const queue = reconciliationQueue(async () => {
    calls++;
    await new Promise<void>((resolve) => {
      release = resolve;
    });
  });
  const pending = queue.request();
  const second = queue.request();
  queue.close();
  release();
  await Promise.all([pending, second]);
  assert.equal(calls, 1);
});
