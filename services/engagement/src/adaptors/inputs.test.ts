import {test} from 'node:test';
import assert from 'node:assert/strict';
import {redeemInput} from './inputs.js';

test('redeem input accepts one string field and rejects other shapes', () => {
  assert.deepEqual(redeemInput({orderId: 'order'}), {orderId: 'order'});
  for (const value of [null, [], {}, {orderId: 1}, {orderId: 'order', extra: true}]) {
    assert.throws(() => redeemInput(value), {code: 'invalid_request'});
  }
});
