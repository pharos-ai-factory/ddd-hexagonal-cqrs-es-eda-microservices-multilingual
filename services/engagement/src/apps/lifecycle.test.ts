import {test} from 'node:test';
import assert from 'node:assert/strict';
import {RuntimeLifecycle} from './lifecycle.js';
test('partial composition failure closes earlier resources once', async () => {
  const scope = new RuntimeLifecycle(); let closed = 0;
  await scope.acquire(() => ({dispose: async () => {closed++;}}));
  await assert.rejects(scope.acquire(() => {throw Error('registration');}));
  await scope.stop(); await scope.stop(); assert.equal(closed, 1);
});
test('shutdown drains an in-flight worker before resource disposal', async () => {
  const scope = new RuntimeLifecycle(), order: string[] = [];
  let release!: () => void;
  scope.track(new Promise<void>(resolve => {release = () => {order.push('worker'); resolve();};}));
  await scope.acquire(() => ({dispose: async () => {order.push('pool');}}));
  const stopped = scope.stop(); assert.deepEqual(order, []);
  release(); await stopped; assert.deepEqual(order, ['worker', 'pool']);
});
test('one failed disposer does not strand another resource', async () => {
 const scope = new RuntimeLifecycle(); let closed = false;
 await scope.acquire(() => ({dispose: async () => {closed=true;}}));
 await scope.acquire(() => ({dispose: async () => {throw Error('close');}}));
 await assert.rejects(scope.stop()); assert.ok(closed);
});
