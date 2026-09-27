import assert from 'node:assert/strict';
import {Given, When, Then} from '@cucumber/cucumber';
import {derived} from './client.js';
import type {WorkflowWorld as W} from './world.js';

When('the last OrderCollected fact is delivered with a new event identity', async function(this: W) {
  this.accountBefore = await this.client.get('loyalty/accounts/'+this.customer);
  this.replay = await this.evidence.redeliver('collection.order-collected', this.last.pickup);
});
Then('the new delivery has a committed consumer receipt', async function(this: W) {
  await this.evidence.receipt(this.replay.consumer, this.replay.id);
});
Then('the account still has {int} collections, {int} stamps and {int} grants',
  async function(this: W, collections: number, stamps: number, grants: number) {
    const account = await this.client.get('loyalty/accounts/'+this.customer);
    assert.equal(account.state.collections, collections);
    assert.equal(account.state.stampBalance, stamps);
    assert.equal(account.state.grantsEarned, grants);
    if (this.accountBefore) assert.deepEqual(account, this.accountBefore);
  });
Then('the customer has no reward', async function(this: W) {
  assert.deepEqual(await this.evidence.roots('loyalty', 'reward', 'customerId', this.customer), []);
  assert.deepEqual(await this.evidence.events('loyalty', 'loyalty.reward-earned', this.customer), []);
});
When('the RewardEarned fact is delivered with a new event identity', async function(this: W) {
  this.replay = await this.evidence.redeliver('loyalty.reward-earned', this.customer);
});
Then('the original reward and its single notification are unchanged', async function(this: W) {
  assert.deepEqual(await this.evidence.roots('loyalty', 'reward', 'customerId', this.customer), [this.reward]);
  assert.deepEqual(await this.client.get('communication/notifications/'+this.noticeBefore.state.id), this.noticeBefore);
  assert.equal((await this.evidence.events('loyalty', 'loyalty.reward-issued', this.reward.state.id)).length, 1);
  assert.equal((await this.evidence.events('communication', 'communication.notification-requested', this.noticeBefore.state.id)).length, 1);
});
When('the OrderPlaced fact is delivered with a new event identity', async function(this: W) {
  this.before = await this.client.get('preparation/tickets/'+this.last.ticket);
  this.replay = await this.evidence.redeliver('ordering.order-placed', this.last.id);
});
Then('the original ready ticket is unchanged', async function(this: W) {
  assert.equal(this.before.state.status, 'ready');
  assert.deepEqual(await this.evidence.roots('preparation', 'ticket', 'orderId', this.last.id), [this.before]);
  assert.equal((await this.evidence.events('preparation', 'preparation.drinks-ready', this.last.ticket)).length, 1);
});
When('the DrinksReady fact is delivered with a new event identity', async function(this: W) {
  this.before = await this.client.get('collection/pickups/'+this.last.pickup);
  this.accountBefore = await this.client.get('loyalty/accounts/'+this.customer);
  this.replay = await this.evidence.redeliver('preparation.drinks-ready', this.last.ticket);
});
Then('the original collected pickup is unchanged', async function(this: W) {
  assert.equal(this.before.state.status, 'collected');
  assert.deepEqual(await this.evidence.roots('collection', 'pickup', 'orderId', this.last.id), [this.before]);
  assert.equal((await this.evidence.events('collection', 'collection.order-collected', this.last.pickup)).length, 1);
});
Given('the pickup notification has been delivered', async function(this: W) {
  this.noticeBefore = await this.notice(derived('pickup-notice', this.last.pickup));
  assert.equal((await this.acceptances(this.noticeBefore.state.id)).length, 1);
});
When('the PickupOpened fact is delivered with a new event identity', async function(this: W) {
  this.replay = await this.evidence.redeliver('collection.pickup-opened', this.last.pickup);
});
Then('the pickup notification and its single provider acceptance are unchanged', async function(this: W) {
  assert.deepEqual(await this.client.get('communication/notifications/'+this.noticeBefore.state.id), this.noticeBefore);
  assert.equal((await this.evidence.events('communication', 'communication.notification-requested', this.noticeBefore.state.id)).length, 1);
  assert.equal((await this.acceptances(this.noticeBefore.state.id)).length, 1);
});
