import assert from 'node:assert/strict';
import {Given, When, Then} from '@cucumber/cucumber';
import {PickupNotice, RewardNotice, DeliverNotification} from '../../src/contexts/communication/application/commands.js';
import {customer, metadata, notification, selectedOrder} from './probes.js';
import type {EngagementWorld as W} from './world.js';

async function request(w: W, code = 'ABC123') {
  await new PickupNotice(w.notices).handle(metadata(notification), {
    pickupId: selectedOrder, orderId: selectedOrder, customerId: customer, collectionCode: code,
  });
  w.notices.succeeded();
}
async function deliver(w: W) {
  w.beforeDelivery = structuredClone(w.notices.loaded);
  w.commandsBeforeDelivery = w.notices.calls;
  w.deliveryError = undefined;
  try {
    await new DeliverNotification(w.notices, w.notices, w.provider).handle(metadata(notification), {notificationId: notification});
  } catch (error) { w.deliveryError = error; }
}
When('Communication handles an opened pickup with code {string}', function(this: W, code: string) { return request(this, code); });
When('Communication handles a reward valid until {string}', async function(this: W, expiresAt: string) {
  await new RewardNotice(this.notices).handle(metadata(notification), {
    rewardId: selectedOrder, customerId: customer, benefit: 'one free drink', expiresAt,
  });
  this.notices.succeeded();
});
Then('the requested notification tells the customer {string}', function(this: W, body: string) {
  assert.equal(this.notices.loaded?.state.status, 'requested');
  assert.equal(this.notices.loaded.state.recipient, customer);
  assert.equal(this.notices.loaded.state.body, body);
});
Then('one private NotificationRequested publication identifies the notification', function(this: W) {
  assert.deepEqual(this.notices.publications, [{name: 'communication.notification-requested', payload: {notificationId: notification}}]);
});
Then('no provider call has been made', function(this: W) { assert.equal(this.providerCalls, 0); });
Given('a pickup notification has been requested', function(this: W) { return request(this); });
Then('the Notification and its outgoing events are unchanged', function(this: W) {
  assert.equal(this.deliveryError, undefined);
  this.notices.succeeded();
  this.notices.unchanged();
});
When('the delivery worker handles the request', function(this: W) { return deliver(this); });
Then('the notification is sent with provider receipt {string}', function(this: W, receipt: string) {
  assert.equal(this.deliveryError, undefined);
  this.notices.succeeded();
  assert.equal(this.notices.loaded?.state.status, 'sent');
  assert.equal(this.notices.loaded.state.providerReceipt, receipt);
});
Then('the provider has received exactly {int} call', function(this: W, count: number) { assert.equal(this.providerCalls, count); });
Then('delivery produces no outgoing business events', function(this: W) { assert.deepEqual(this.notices.publications, []); });
Given('the provider is unavailable', function(this: W) { this.unavailable = true; });
Then('the provider failure is available to the retry mechanism', function(this: W) {
  assert.ok(this.deliveryError instanceof Error);
  assert.equal(this.deliveryError.message, 'provider unavailable');
});
Then('the notification is still requested at its original version', function(this: W) {
  assert.equal(this.notices.loaded?.state.status, 'requested');
  assert.deepEqual(this.notices.loaded, this.beforeDelivery);
});
Then('no delivery command has committed', function(this: W) { assert.equal(this.notices.calls, this.commandsBeforeDelivery); });
Given('a pickup notification has been delivered', async function(this: W) {
  await request(this); await deliver(this);
  assert.equal(this.deliveryError, undefined);
  assert.equal(this.notices.loaded?.state.status, 'sent');
});
