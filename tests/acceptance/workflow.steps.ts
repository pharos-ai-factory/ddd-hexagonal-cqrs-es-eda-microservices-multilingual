import assert from 'node:assert/strict';
import {Given, When, Then} from '@cucumber/cucumber';
import {derived} from './client.js';
import type {WorkflowWorld as W} from './world.js';

Given('a published menu is available to Ordering', function(this: W) { return this.menu(); });
Given(/^the customer has collected (\d+) orders?$/, function(this: W, count: string) { return this.collected(Number(count)); });
Given('the customer has collected 3 orders and received a reward', async function(this: W) {
  await this.collected(3);
  this.reward = (await this.rewards(1))[0]!;
  this.noticeBefore = await this.notice(derived('reward-notice', this.reward.state.id));
});
Given("the customer's order is ready for collection", async function(this: W) { await this.ready(); });
Given("the customer's ticket is queued", async function(this: W) { await this.queued(); });
Given('the customer has {int} orders ready for collection', async function(this: W, count: number) {
  for (let i = 0; i < count; i++) await this.ready();
});
When('all six pickups are collected concurrently', async function(this: W) {
  assert.equal(this.orders.length, 6);
  await Promise.all(this.orders.map(order => this.collect(order)));
});
Then('the account eventually has {int} collections, {int} stamps and {int} grants',
  async function(this: W, collections: number, stamps: number, grants: number) {
    const account = await this.account(collections);
    assert.equal(account.state.stampBalance, stamps);
    assert.equal(account.state.grantsEarned, grants);
  });
Then('two different earned grants each create one reward', async function(this: W) {
  const rewards = await this.rewards(2);
  assert.equal(new Set(rewards.map(reward => reward.state.grantId)).size, 2);
  const earned = await this.evidence.events('loyalty', 'loyalty.reward-earned', this.customer);
  assert.equal(earned.length, 2);
  for (const event of earned) await this.evidence.receipt('loyalty.issue-reward', event.id);
  for (const reward of rewards) {
    assert.equal(reward.version, 1);
    assert.equal((await this.evidence.events('loyalty', 'loyalty.reward-issued', reward.state.id)).length, 1);
  }
});
When('another customer collects 2 orders', async function(this: W) { await this.collected(2, this.otherCustomer); });
Then('each customer has 2 stamps and no earned grant', async function(this: W) {
  for (const customer of [this.customer, this.otherCustomer]) {
    const account = await this.account(2, customer);
    assert.equal(account.state.stampBalance, 2);
    assert.equal(account.state.grantsEarned, 0);
    assert.deepEqual(await this.evidence.roots('loyalty', 'reward', 'customerId', customer), []);
  }
});
Given('the next provider acceptance will lose its response', async function(this: W) {
  await this.client.provider('/failures', {lostResponses: 1});
});
When("the customer's order becomes ready for collection", async function(this: W) { await this.ready(); });
Then('the pickup notification eventually records delivery', async function(this: W) {
  const notice = await this.notice(derived('pickup-notice', this.last.pickup));
  assert.ok(notice.state.providerReceipt);
});
Then('the provider has accepted that notification exactly once', async function(this: W) {
  assert.equal((await this.acceptances(derived('pickup-notice', this.last.pickup))).length, 1);
});
When('completion is requested before preparation starts', async function(this: W) {
  this.completionReply = await this.client.request(`preparation/tickets/${this.last.ticket}/complete`, {}, 1, this.completionKey);
});
Then('completion is rejected with {string}', function(this: W, code: string) {
  assert.equal(this.completionReply.status, 422);
  assert.equal(this.completionReply.body.rejection?.code, code);
});
When('the barista starts preparation and retries that completion command', async function(this: W) {
  await this.client.command(`preparation/tickets/${this.last.ticket}/start`, {}, 1);
  const retry = await this.client.request(`preparation/tickets/${this.last.ticket}/complete`, {}, 1, this.completionKey);
  this.redemptionReplies = [retry];
});
Then('the original completion rejection is returned', function(this: W) {
  assert.deepEqual(this.redemptionReplies[0], this.completionReply);
});
Then('no DrinksReady fact has been committed', async function(this: W) {
  assert.deepEqual(await this.evidence.events('preparation', 'preparation.drinks-ready', this.last.ticket), []);
  const ticket = await this.client.get('preparation/tickets/'+this.last.ticket);
  assert.equal(ticket.state.status, 'preparing');
  assert.equal(ticket.version, 2);
});
When('the barista makes a new completion attempt', async function(this: W) {
  await this.client.command(`preparation/tickets/${this.last.ticket}/complete`, {}, 2);
});
Then('the order becomes ready with one DrinksReady fact', async function(this: W) {
  await this.client.state('collection/pickups/'+this.last.pickup, state => state.status === 'ready');
  assert.equal((await this.evidence.events('preparation', 'preparation.drinks-ready', this.last.ticket)).length, 1);
});
