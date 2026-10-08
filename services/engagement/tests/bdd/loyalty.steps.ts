import assert from 'node:assert/strict';
import {Given, When, Then} from '@cucumber/cucumber';
import {CreditCollection, IssueReward, RedeemReward} from '../../src/contexts/loyalty/application/commands.js';
import {derivedId} from '../../src/foundation/identity.js';

import type {RewardEarned} from '../../src/contexts/loyalty/application/events.js';
import {customer, metadata, selectedOrder} from './probes.js';
import type {EngagementWorld as W} from './world.js';

async function credit(w: W, count: number) {
  for (let i = 0; i < count; i++) {
    const orderId = derivedId('scenario-order', String(++w.credits));
    await new CreditCollection(w.accounts, derivedId).handle(metadata(customer), {orderId, customerId: customer});
    w.accounts.succeeded();
    w.earned.push(...w.accounts.publications);
  }
}
async function earn(w: W) {
  await credit(w, 3);
  w.accountBeforeReward = structuredClone(w.accounts.loaded);
}
async function issue(w: W, instant: string) {
  w.now = new Date(instant);
  const grant = w.earned[0]?.payload as RewardEarned;
  assert.ok(grant);
  const id = derivedId('reward', grant.grantId);
  await new IssueReward(w.rewards, derivedId, () => w.now).handle(metadata(id), grant);
}
async function redeem(w: W, instant: string) {
  w.now = new Date(instant);
  assert.ok(w.rewards.loaded);
  await new RedeemReward(w.rewards, () => w.now).execute(metadata(w.rewards.loaded.state.id), {orderId: selectedOrder});
}

Given('a customer has no credited collections', function(this: W) {
  assert.equal(this.accounts.loaded, undefined);
});
When('{int} different orders are credited', function(this: W, count: number) { return credit(this, count); });
Then('the account has {int} stamps and {int} earned grants', function(this: W, balance: number, grants: number) {
  assert.equal(this.accounts.loaded?.state.collections, this.credits);
  assert.equal(this.accounts.loaded.state.stampBalance, balance);
  assert.equal(this.accounts.loaded.state.grantsEarned, grants);
});
Then('exactly {int} private RewardEarned publications are produced', function(this: W, count: number) {
  assert.equal(this.earned.length, count);
  assert.ok(this.earned.every(event => event.name === 'loyalty.reward-earned'));
  assert.equal(new Set(this.earned.map(event => (event.payload as RewardEarned).grantId)).size, count);
});
Then('no Reward has been created by the credit commands', function(this: W) {
  assert.equal(this.rewards.calls, 0);
  assert.equal(this.rewards.loaded, undefined);
});
Then('the private grant promises {string} valid for {int} days', function(this: W, benefit: string, days: number) {
  assert.equal(this.earned.length, 1);
  const grant = this.earned[0]!.payload as RewardEarned;
  assert.equal(grant.accountId, customer);
  assert.equal(grant.benefit, benefit);
  assert.equal(grant.validDays, days);
});
Then('the recorded grant matches the published grant identity', function(this: W) {
  const grant = this.earned[0]!.payload as RewardEarned;
  assert.equal(grant.grantId, this.accounts.loaded?.state.lastGrant?.id);
});
Given('the customer has earned a grant', function(this: W) { return earn(this); });
When('the grant is handled at {string}', function(this: W, instant: string) { return issue(this, instant); });
Then('one public RewardIssued publication expires at {string}', function(this: W, expiresAt: string) {
  this.rewards.succeeded();
  const state = this.rewards.loaded!.state;
  assert.equal(state.status, 'issued');
  assert.equal(state.grantId, (this.earned[0]!.payload as RewardEarned).grantId);
  assert.deepEqual(this.rewards.publications, [{name: 'loyalty.reward-issued', payload: {
    rewardId: state.id, customerId: customer, benefit: 'one free drink', expiresAt,
  }}]);
  assert.equal(state.expiresAt, expiresAt);
});
Then('the LoyaltyAccount is unchanged by the reward command', function(this: W) {
  assert.deepEqual(this.accounts.loaded, this.accountBeforeReward);
  assert.equal(this.accounts.calls, 3);
});
Given('a reward was issued at {string}', async function(this: W, instant: string) {
  await earn(this); await issue(this, instant); this.rewards.succeeded();
});
Then('the reward command succeeds', function(this: W) { this.rewards.succeeded(); });
Then('the Reward and its outgoing events are unchanged', function(this: W) { this.rewards.unchanged(); });
When('the reward is redeemed at {string}', function(this: W, instant: string) { return redeem(this, instant); });
Then('the reward is redeemed for the selected order', function(this: W) {
  this.rewards.succeeded();
  assert.equal(this.rewards.loaded?.state.status, 'redeemed');
  assert.equal(this.rewards.loaded.state.redeemedFor, selectedOrder);
  assert.deepEqual(this.rewards.publications, []);
});
Then('the reward command is rejected with {string}', function(this: W, code: string) {
  assert.equal(this.rewards.outcome?.rejection?.code, code);
});
Given('the reward has been redeemed', async function(this: W) {
  await redeem(this, '2026-01-02T11:00:00.000Z'); this.rewards.succeeded();
});
