import { test } from 'node:test';
import assert from 'node:assert/strict';
import { RealtimeBuffer } from './publication-buffer.js';

void test('explicit publication copies its value and expires with its invocation', async () => {
  const bus = new RealtimeBuffer(),
    state = { status: 'issued' };
  assert.deepEqual(bus.events, []);
  await bus.publish({ snapshot: state });
  state.status = 'redeemed';
  assert.deepEqual(bus.events, [{ snapshot: { status: 'issued' } }]);
  bus.close();
  await assert.rejects(bus.publish({ snapshot: state }), /outside its transaction/);
});
