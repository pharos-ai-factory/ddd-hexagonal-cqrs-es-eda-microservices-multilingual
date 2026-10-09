import {test} from 'node:test';
import assert from 'node:assert/strict';
import {asFunction, asValue} from 'awilix';
import {createLoyaltyContainer} from './loyalty.js';
import {createCommunicationContainer} from './communication.js';
import {OrderCollectedIntegrationEventHandler} from '../../contexts/loyalty/application/event-handlers/order-collected.js';
import type {AggregateTransaction} from '../../adaptors/command-execution.js';
import type {LoyaltyAccount, AccountState} from '../../contexts/loyalty/domain/loyalty-account.js';

const url = 'postgresql://unused:unused@127.0.0.1:1/unused';
test('every context provider resolves with isolated resources and pool disposal', async () => {
  const containers = [createLoyaltyContainer(url, () => new Date('2026-01-01T00:00:00Z')),
    createCommunicationContainer(url, 'http://127.0.0.1:1', 'unused')];
  for (const [index, container] of containers.entries()) {
    for (const name of Object.keys(container.registrations)) container.resolve(name);
    const database = container.resolve('database');
    assert.equal(database.owner, index === 0 ? 'loyalty' : 'communication');
    assert.equal(container.resolve('database'), database);
    let closed = 0;
    database.pool.end = async () => { closed++; };
    await container.dispose();
    assert.equal(closed, 1);
  }
});
test('Awilix binds feature execution to the shared command boundary and reports missing dependencies', async () => {
  const container = createLoyaltyContainer(url);
  assert.ok(container.resolve('collected') instanceof OrderCollectedIntegrationEventHandler);
  const saved: AccountState[] = [];
  const transaction: AggregateTransaction<LoyaltyAccount> = {execute: async (metadata, work) => {
    const result = await work({get: async () => undefined, save: async account => { saved.push(account.snapshot()); }});
    return {aggregateId: metadata.target, version: 1, status: result.status};
  }};
  container.register({accountTransaction: asValue(transaction)});
  const id = '00000000-0000-4000-8000-000000000001';
  const outcome = await container.resolve('credit').execute({id, target: id, correlation: id, name: 'fixture', input: {}}, {customerId: id, orderId: id});
  assert.equal(outcome.status, 'active');
  assert.equal(saved.length, 1);
  assert.equal(saved[0]?.collections, 1);
  await container.dispose();
  const broken = createLoyaltyContainer(url);
  broken.register({accountTransaction: asFunction(() => { throw new Error('missing aggregate port'); }).singleton()});
  assert.throws(() => broken.resolve('credit'), /missing aggregate port/);
  await broken.dispose();
});
