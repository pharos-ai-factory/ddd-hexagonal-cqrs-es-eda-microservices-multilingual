import { test } from 'node:test';
import assert from 'node:assert/strict';
import {
  parseRedeemRewardCommand,
  parseIssueRewardCommand,
  parseListRewardsQuery,
} from './generated/request-payloads.js';

await test('generated query inputs preserve complete reads and explicit pagination', () => {
  assert.deepEqual(parseListRewardsQuery({}), {});
  assert.deepEqual(parseListRewardsQuery({ page: { limit: 0 } }), { page: { limit: 0 } });
  assert.deepEqual(parseListRewardsQuery({ page: { limit: 2, after: '' } }), {
    page: { limit: 2, after: '' },
  });
  assert.throws(() => parseListRewardsQuery({ page: null }));
  assert.throws(() => parseListRewardsQuery({ page: { limit: -1 } }));
});

await test('generated command parsing enforces required presence and scalar ranges', () => {
  assert.throws(() => parseRedeemRewardCommand({}));
  assert.deepEqual(parseRedeemRewardCommand({ orderId: '' }), { orderId: '' });
  assert.equal(parseIssueRewardCommand({ validDays: 0 }).validDays, 0);
  for (const validDays of [2 ** 31, 1.5, NaN, '3']) {
    assert.throws(() => parseIssueRewardCommand({ validDays }));
  }
});
