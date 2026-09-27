import assert from 'node:assert/strict';
import {randomUUID} from 'node:crypto';
import {When, Then} from '@cucumber/cucumber';
import type {WorkflowWorld as W} from './world.js';

When('two different commands redeem that reward version concurrently', async function(this: W) {
  this.accountBefore = await this.client.get('loyalty/accounts/'+this.customer);
  this.redemptionReplies = await Promise.all([randomUUID(), randomUUID()].map(orderId =>
    this.client.request(`loyalty/rewards/${this.reward.state.id}/redeem`, {orderId}, this.reward.version)));
});
Then('one redemption succeeds and one reports a version conflict', function(this: W) {
  assert.deepEqual(this.redemptionReplies.map(reply => reply.status).sort(), [200, 409]);
  assert.equal(this.redemptionReplies.find(reply => reply.status === 409)!.body.rejection?.code, 'version_conflict');
});
Then('the reward is redeemed once without changing the account', async function(this: W) {
  const reward = await this.client.get('loyalty/rewards/'+this.reward.state.id);
  assert.equal(reward.state.status, 'redeemed');
  assert.equal(reward.version, this.reward.version + 1);
  assert.ok(reward.state.redeemedFor);
  assert.deepEqual(await this.client.get('loyalty/accounts/'+this.customer), this.accountBefore);
  const another = await this.client.request(`loyalty/rewards/${reward.state.id}/redeem`, {orderId: randomUUID()}, reward.version);
  assert.equal(another.status, 422);
  assert.equal(another.body.rejection?.code, 'reward_unavailable');
  assert.deepEqual(await this.client.get('loyalty/rewards/'+reward.state.id), reward);
});
