import assert from 'node:assert/strict';
import {randomUUID} from 'node:crypto';
import {Before, After, setDefaultTimeout, setWorldConstructor, World} from '@cucumber/cucumber';
import {Client, configuration, derived, eventually, type Loaded, type Reply} from './client.js';
import {Evidence} from './evidence.js';

export type Order = {id: string; ticket: string; pickup: string; customer: string; code?: string};
export class WorkflowWorld extends World {
  client!: Client;
  evidence!: Evidence;
  customer = randomUUID();
  otherCustomer = randomUUID();
  edition = randomUUID();
  orders: Order[] = [];
  replay!: {id: string; consumer: string};
  reward!: Loaded;
  before!: Loaded;
  accountBefore!: Loaded;
  noticeBefore!: Loaded;
  redemptionReplies: Reply[] = [];
  completionKey = randomUUID();
  completionReply!: Reply;

  get last(): Order { assert.ok(this.orders.length); return this.orders.at(-1)!; }
  async menu() {
    const drink = randomUUID();
    await this.client.command(`menu/drinks/${drink}`, {name: 'Coffee'}, 0);
    await this.client.command(`menu/drinks/${drink}/publish`, {}, 1);
    await this.client.command(`menu/editions/${this.edition}`, {currency: 'EUR'}, 0);
    await this.client.availableCommand(`menu/editions/${this.edition}/offers`, {code: 'C1', drinkId: drink, drinkRevision: 1, minor: 300}, 1);
    await this.client.command(`menu/editions/${this.edition}/publish`, {}, 2);
  }
  async queued(customer = this.customer): Promise<Order> {
    const id = randomUUID();
    const order = {id, ticket: derived('ticket', id), pickup: derived('pickup', id), customer};
    const base = 'ordering/orders/'+id;
    await this.client.availableCommand(base, {customerId: customer, editionId: this.edition}, 0);
    await this.client.command(base+'/lines', {lineId: randomUUID(), editionId: this.edition, offerCode: 'C1', quantity: 1}, 1);
    await this.client.command(base+'/place', {}, 2);
    await this.client.state('preparation/tickets/'+order.ticket, state => state.status === 'queued');
    this.orders.push(order);
    return order;
  }
  async ready(customer = this.customer): Promise<Order> {
    const order = await this.queued(customer);
    await this.client.command(`preparation/tickets/${order.ticket}/start`, {}, 1);
    await this.client.command(`preparation/tickets/${order.ticket}/complete`, {}, 2);
    const pickup = await this.client.state('collection/pickups/'+order.pickup, state => state.status === 'ready');
    order.code = pickup.state.code;
    return order;
  }
  async collect(order: Order) {
    assert.ok(order.code);
    await this.client.command(`collection/pickups/${order.pickup}/collect`, {code: order.code}, 1);
  }
  async collected(count: number, customer = this.customer) {
    for (let i = 0; i < count; i++) await this.collect(await this.ready(customer));
    await this.account(count, customer);
  }
  async account(collections: number, customer = this.customer) {
    return this.client.state('loyalty/accounts/'+customer, state => state.collections === collections);
  }
  async rewards(count: number): Promise<Loaded[]> {
    return eventually(count+' rewards', async () => {
      const rewards = await this.evidence.roots('loyalty', 'reward', 'customerId', this.customer);
      return rewards.length === count ? rewards : undefined;
    });
  }
  async notice(id: string) {
    return this.client.state('communication/notifications/'+id, state => state.status === 'sent');
  }
  async acceptances(id: string) {
    const messages = await this.client.provider('/messages') as {message: {id: string}}[];
    return messages.filter(item => item.message.id === id);
  }
  async settle() {
    // Positive terminal observations prevent one scenario's unfinished provider
    // work from consuming the next scenario's deliberate lost response.
    for (const customer of new Set(this.orders.map(order => order.customer))) {
      const pickups = await this.evidence.roots('collection', 'pickup', 'customerId', customer);
      const account = await this.client.request('loyalty/accounts/'+customer);
      const grantCount = account.status === 200 ? account.body.state.grantsEarned : 0;
      const rewards = await eventually('all earned rewards', async () => {
        const values = await this.evidence.roots('loyalty', 'reward', 'customerId', customer);
        return values.length === grantCount ? values : undefined;
      });
      for (const pickup of pickups) await this.notice(derived('pickup-notice', pickup.state.id));
      for (const reward of rewards) await this.notice(derived('reward-notice', reward.state.id));
    }
  }
}
setDefaultTimeout(120_000);
setWorldConstructor(WorkflowWorld);
Before(function(this: WorkflowWorld) {
  const values = configuration();
  this.client = new Client(values); this.evidence = new Evidence(values);
});
After(async function(this: WorkflowWorld, {result}) {
  try {
    if (result?.status === 'PASSED') await this.settle();
  } finally { await this.evidence?.close(); }
});
