import { RuntimeLifecycle } from './lifecycle.js';
import type { Server } from 'node:http';
import { composeLoyalty } from './composition/loyalty.js';
import { composeCommunication } from './composition/communication.js';
import { diagnostics } from '../adaptors/diagnostics.js';
import { relay } from '../adaptors/messaging/rabbitmq-outbox-relay.js';
import { consume } from '../adaptors/messaging/rabbitmq-consumer.js';
import { realtimeRelay } from '../adaptors/realtime/gateway-relay.js';
import { server } from '../adaptors/http/operational-security.js';
import { replyRelay } from '../adaptors/messaging/rabbitmq-reply-relay.js';

import { secret as required } from '../foundation/secrets.js';
if (!['local', 'development'].includes(process.env.APP_ENV ?? ''))
  throw new Error('Development environments only');
const lifecycle = new RuntimeLifecycle();
const controller = lifecycle.controller;
const tasks = { push: (task: Promise<void>) => lifecycle.track(task) };
let http: Server | undefined;
try {
  const loyalty = await lifecycle.acquire(composeLoyalty);
  const communication = await lifecycle.acquire(composeCommunication);
  const contexts = [loyalty, communication];
  const databases = { loyalty: loyalty.database, communication: communication.database };
  await Promise.all(contexts.map((c) => c.database.verify()));
  // Validate all credentials before starting any worker.
  for (const name of [
    'API_KEY',
    'LOYALTY_BROKER_URL',
    'COMMUNICATION_BROKER_URL',
    'REALTIME_GATEWAY_URL',
    'LOYALTY_REALTIME_KEY',
    'COMMUNICATION_REALTIME_KEY',
  ])
    required(name);
  for (const [owner, db] of Object.entries(databases)) {
    tasks.push(replyRelay(db, required(owner.toUpperCase() + '_BROKER_URL'), controller.signal));
    tasks.push(relay(db, required(owner.toUpperCase() + '_BROKER_URL'), controller.signal));
    tasks.push(
      realtimeRelay(
        db,
        required('REALTIME_GATEWAY_URL'),
        required(owner.toUpperCase() + '_REALTIME_KEY'),
        controller.signal,
      ),
    );
  }
  for (const context of contexts) {
    const url = required(context.database.owner.toUpperCase() + '_BROKER_URL');
    tasks.push(relay(context.database, url, controller.signal, context.commandHeader));
    for (const subscription of context.subscriptions) {
      if ((process.env.PAUSED_CONSUMERS ?? '').split(',').includes(subscription.consumer)) continue;
      tasks.push(consume(url, subscription, controller.signal));
    }
  }
  tasks.push(loyalty.requests.run(required('LOYALTY_BROKER_URL'), controller.signal));
  tasks.push(communication.requests.run(required('COMMUNICATION_BROKER_URL'), controller.signal));
  http = server(required('API_KEY'), () => diagnostics(databases));
  await new Promise<void>((resolve, reject) => {
    http!.once('error', reject);
    http!.listen(8080, '0.0.0.0', () => {
      console.info('Engagement HTTP listening (TypeScript)');
      resolve();
    });
  });
  await new Promise<void>((resolve) => {
    const halted = () => {
      process.removeListener('SIGTERM', halted);
      process.removeListener('SIGINT', halted);
      resolve();
    };
    process.once('SIGTERM', halted);
    process.once('SIGINT', halted);
  });
} finally {
  await lifecycle.stop(
    () =>
      new Promise<void>((resolve) => {
        if (http?.listening) http.close(() => resolve());
        else resolve();
      }),
  );
}
