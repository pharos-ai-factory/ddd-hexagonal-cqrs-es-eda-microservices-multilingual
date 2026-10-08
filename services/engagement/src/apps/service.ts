import {diagnostics} from '../adaptors/diagnostics.js';
import {Database, Commands, Queries} from '../adaptors/postgres.js';
import {restoreAccount, restoreReward, restoreNotification} from '../adaptors/restore.js';
import {redeemInput} from '../adaptors/inputs.js';
import {relay, consume, type Subscription} from '../adaptors/broker.js';
import {realtimeRelay} from '../adaptors/dispatch.js';
import {server} from '../adaptors/http.js';
import {replyRelay} from '../adaptors/replies.js';
import {Registry} from '../adaptors/requests.js';
import * as loyaltyWire from '../contexts/loyalty/adaptors/messaging/requests.js';
import {notification as notificationWire} from '../contexts/communication/adaptors/messaging/requests.js';
import {HttpDelivery} from '../adaptors/provider.js';
import {derivedId} from '../foundation/identity.js';
import type {Metadata, Outcome} from '../foundation/application.js';
import type {AccountState} from '../contexts/loyalty/domain/account.js';
import type {RewardState} from '../contexts/loyalty/domain/reward.js';
import type {NotificationState} from '../contexts/communication/domain/notification.js';
import {CreditCollection, IssueReward, RedeemReward} from '../contexts/loyalty/application/commands.js';
import {PickupNotice, RewardNotice, DeliverNotification} from '../contexts/communication/application/commands.js';
import type {OrderCollected, RewardIssued, PickupOpened} from '../contracts/events.js';
import type {NotificationRequested} from '../contexts/communication/application/events.js';
import type {RewardEarned} from '../contexts/loyalty/application/events.js';

import {secret as required} from '../foundation/secrets.js';
if (!['local', 'development'].includes(process.env.APP_ENV ?? '')) throw new Error('Development environments only');
const databases = {
  loyalty: new Database('loyalty', required('LOYALTY_DATABASE_URL')),
  communication: new Database('communication', required('COMMUNICATION_DATABASE_URL')),
};
await Promise.all(Object.values(databases).map(db => db.verify()));
const accounts = new Commands<AccountState>(databases.loyalty, 'account', restoreAccount);
const rewards = new Commands<RewardState>(databases.loyalty, 'reward', restoreReward);
const notices = new Commands<NotificationState>(databases.communication, 'notification', restoreNotification);
const noticeQueries = new Queries<NotificationState>(databases.communication, 'notification', restoreNotification);
const controller = new AbortController();
const tasks: Promise<void>[] = [];
for (const [owner, db] of Object.entries(databases)) {
  tasks.push(replyRelay(db, required(owner.toUpperCase()+'_BROKER_URL'), controller.signal));
  tasks.push(relay(db, required(owner.toUpperCase()+'_BROKER_URL'), controller.signal));
  tasks.push(realtimeRelay(db, required('REALTIME_GATEWAY_URL'), required(owner.toUpperCase()+'_REALTIME_KEY'), controller.signal));
}
function subscribe<P extends object>(owner: string, consumer: string, event: string, target: (payload: P) => string,
  handler: {handle(m: Metadata, payload: P): Promise<Outcome>}) {
  if ((process.env.PAUSED_CONSUMERS ?? '').split(',').includes(consumer)) return;
  const subscription: Subscription = {owner, consumer, event, target: value => target(value as P),
    handle: (metadata, value) => handler.handle(metadata, value as P)};
  tasks.push(consume(required(owner.toUpperCase()+'_BROKER_URL'), subscription, controller.signal));
}
subscribe<OrderCollected>('loyalty', 'loyalty.credit-collection', 'collection.order-collected',
  event => event.customerId, new CreditCollection(accounts, derivedId));
subscribe<RewardEarned>('loyalty', 'loyalty.issue-reward', 'loyalty.reward-earned',
  event => derivedId('reward', event.grantId), new IssueReward(rewards, derivedId, () => new Date()));
subscribe<PickupOpened>('communication', 'communication.pickup-notice', 'collection.pickup-opened',
  event => derivedId('pickup-notice', event.pickupId), new PickupNotice(notices));
subscribe<RewardIssued>('communication', 'communication.reward-notice', 'loyalty.reward-issued',
  event => derivedId('reward-notice', event.rewardId), new RewardNotice(notices));
subscribe<NotificationRequested>('communication', 'communication.deliver-notice', 'communication.notification-requested',
  event => event.notificationId, new DeliverNotification(notices, noticeQueries,
    new HttpDelivery(required('DELIVERY_URL'), required('DELIVERY_KEY'))));
const redeem = new RedeemReward(rewards, () => new Date());
const loyaltyRequests = new Registry('loyalty');
loyaltyRequests.command('redeemReward', loyaltyWire.redeem, (m, input) => redeem.execute(m, input));
loyaltyRequests.queries('account', 'accounts', new Queries<AccountState>(databases.loyalty, 'account', restoreAccount), loyaltyWire.account);
loyaltyRequests.queries('reward', 'rewards', new Queries<RewardState>(databases.loyalty, 'reward', restoreReward), loyaltyWire.reward);
const communicationRequests = new Registry('communication');
communicationRequests.queries('notification', 'notifications', noticeQueries, notificationWire);
tasks.push(loyaltyRequests.run(required('LOYALTY_BROKER_URL'), controller.signal));
tasks.push(communicationRequests.run(required('COMMUNICATION_BROKER_URL'), controller.signal));
const http = server(required('API_KEY'), [
  {resource: '/v1/loyalty/accounts', queries: new Queries<AccountState>(databases.loyalty, 'account', restoreAccount)},
  {resource: '/v1/loyalty/rewards', queries: new Queries<RewardState>(databases.loyalty, 'reward', restoreReward),
    commands: {redeem: {name: 'loyalty.RedeemReward', invoke: (m, value) => {
      const input = redeemInput(value);
      return redeem.execute({...m, input}, input);
    }}}},
  {resource: '/v1/communication/notifications', queries: noticeQueries},
], () => diagnostics(databases));
http.listen(8080, '0.0.0.0', () => console.info('Engagement ready (TypeScript)'));
async function stop() {
  controller.abort(); http.close();
  await Promise.allSettled(tasks);
  await Promise.all(Object.values(databases).map(db => db.pool.end()));
}
process.once('SIGTERM', () => void stop());
process.once('SIGINT', () => void stop());
