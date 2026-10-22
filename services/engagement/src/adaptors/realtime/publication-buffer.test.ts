import { test } from 'node:test';
import assert from 'node:assert/strict';
import { RealtimeBuffer } from './publication-buffer.js';
import { realtimeMessage, realtimeTarget } from '../../foundation/application.js';

void test('explicit publication copies its value and expires with its invocation', async () => {
  const bus = new RealtimeBuffer(),
    state = { status: 'issued' };
  assert.deepEqual(bus.events, []);
  await bus.publish(realtimeMessage.snapshot(realtimeTarget.context('loyalty'), state));
  state.status = 'redeemed';
  assert.deepEqual(bus.events, [
    { target: { kind: 'context', id: 'loyalty' }, snapshot: { status: 'issued' } },
  ]);
  bus.close();
  await assert.rejects(
    bus.publish(realtimeMessage.snapshot(realtimeTarget.context('loyalty'), state)),
    /outside its transaction/,
  );
});
