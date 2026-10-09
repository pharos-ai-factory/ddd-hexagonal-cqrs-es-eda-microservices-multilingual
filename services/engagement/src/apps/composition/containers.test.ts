import {test} from 'node:test';
import assert from 'node:assert/strict';
import {asFunction} from 'awilix';
import {createLoyaltyContainer} from './loyalty.js';
import {createCommunicationContainer} from './communication.js';
import {OrderCollectedIntegrationEventHandler} from '../../contexts/loyalty/application/event-handlers.js';
import {CreditCollectionCommandHandler} from '../../contexts/loyalty/application/commands.js';

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
test('Awilix resolves plain application handlers and reports a broken binding', async () => {
  const container = createLoyaltyContainer(url);
  assert.ok(container.resolve('collected') instanceof OrderCollectedIntegrationEventHandler);
  assert.ok(container.resolve('credit') instanceof CreditCollectionCommandHandler);
  await container.dispose();
  const broken = createLoyaltyContainer(url);
  broken.register({accounts: asFunction(() => { throw new Error('missing aggregate port'); }).singleton()});
  assert.throws(() => broken.resolve('credit'), /missing aggregate port/);
  await broken.dispose();
});
