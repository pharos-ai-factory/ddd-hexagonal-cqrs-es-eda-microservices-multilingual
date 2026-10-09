import type {Server} from 'node:http';
import {composeLoyalty} from './composition/loyalty.js';
import {composeCommunication} from './composition/communication.js';
import {diagnostics} from '../adaptors/diagnostics.js';
import {redeemInput} from '../adaptors/inputs.js';
import {relay, consume} from '../adaptors/broker.js';
import {realtimeRelay} from '../adaptors/dispatch.js';
import {server} from '../adaptors/http.js';
import {replyRelay} from '../adaptors/replies.js';
import {RabbitMQRequestRegistry} from '../adaptors/requests.js';
import * as loyaltyWire from '../contexts/loyalty/adaptors/messaging/requests.js';
import {notification as notificationWire} from '../contexts/communication/adaptors/messaging/requests.js';

import {secret as required} from '../foundation/secrets.js';
if (!['local', 'development'].includes(process.env.APP_ENV ?? '')) throw new Error('Development environments only');
const loyalty = composeLoyalty();
const communication = composeCommunication();
const contexts = [loyalty, communication];
const databases = {loyalty: loyalty.database, communication: communication.database};
const noticeQueries = communication.queries;
const controller = new AbortController();
const tasks: Promise<void>[] = [];
let http: Server | undefined;
let stopping: Promise<void> | undefined;
function stop() {
  return stopping ??= (async () => {
    controller.abort();
    const closing = new Promise<void>(resolve => { if (http?.listening) http.close(() => resolve()); else resolve(); });
    await Promise.allSettled([...tasks, closing]);
    await Promise.all(contexts.map(c => c.dispose()));
  })();
}
try {
await Promise.all(contexts.map(c => c.database.verify()));
// Validate all credentials before starting any worker.
for (const name of ['API_KEY', 'LOYALTY_BROKER_URL', 'COMMUNICATION_BROKER_URL', 'REALTIME_GATEWAY_URL',
  'LOYALTY_REALTIME_KEY', 'COMMUNICATION_REALTIME_KEY']) required(name);
for (const [owner, db] of Object.entries(databases)) {
  tasks.push(replyRelay(db, required(owner.toUpperCase()+'_BROKER_URL'), controller.signal));
  tasks.push(relay(db, required(owner.toUpperCase()+'_BROKER_URL'), controller.signal));
  tasks.push(realtimeRelay(db, required('REALTIME_GATEWAY_URL'), required(owner.toUpperCase()+'_REALTIME_KEY'), controller.signal));
}
for (const context of contexts) {
  const url = required(context.database.owner.toUpperCase()+'_BROKER_URL');
  tasks.push(relay(context.database, url, controller.signal, context.commandHeader));
  for (const subscription of context.subscriptions) {
    if ((process.env.PAUSED_CONSUMERS ?? '').split(',').includes(subscription.consumer)) continue;
    tasks.push(consume(url, subscription, controller.signal));
  }
}
const redeem = loyalty.redeem;
const loyaltyRequests = new RabbitMQRequestRegistry('loyalty');
loyaltyRequests.command('redeemReward', loyaltyWire.redeem, (m, input) => redeem.execute(m, input));
loyaltyRequests.queries('account', 'accounts', loyalty.accountQueries, loyaltyWire.account);
loyaltyRequests.queries('reward', 'rewards', loyalty.rewardQueries, loyaltyWire.reward);
const communicationRequests = new RabbitMQRequestRegistry('communication');
communicationRequests.queries('notification', 'notifications', noticeQueries, notificationWire);
tasks.push(loyaltyRequests.run(required('LOYALTY_BROKER_URL'), controller.signal));
tasks.push(communicationRequests.run(required('COMMUNICATION_BROKER_URL'), controller.signal));
http = server(required('API_KEY'), [
  {resource: '/v1/loyalty/accounts', queries: loyalty.accountQueries},
  {resource: '/v1/loyalty/rewards', queries: loyalty.rewardQueries,
    commands: {redeem: {name: 'loyalty.RedeemReward', invoke: (m, value) => {
      const input = redeemInput(value);
      return redeem.execute({...m, input}, input);
    }}}},
  {resource: '/v1/communication/notifications', queries: noticeQueries},
], () => diagnostics(databases));
await new Promise<void>((resolve, reject) => {
  http!.once('error', reject);
  http!.listen(8080, '0.0.0.0', () => { console.info('Engagement ready (TypeScript)'); resolve(); });
});
await new Promise<void>(resolve => {
  const halted = () => { process.removeListener('SIGTERM', halted); process.removeListener('SIGINT', halted); resolve(); };
  process.once('SIGTERM', halted);
  process.once('SIGINT', halted);
});
} finally { await stop(); }
