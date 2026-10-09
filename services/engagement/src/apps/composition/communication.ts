import type {AggregateCommandPort, DeliveryPort} from '../../foundation/application.js';
import type {PagedQueryPort} from '../../foundation/pagination.js';
import {asFunction, createContainer} from 'awilix';
import {PostgresContextDatabase, PostgresAggregateCommandStore, PostgresAggregateQueries} from '../../adaptors/postgres.js';
import {restoreNotification} from '../../adaptors/restore.js';
import {InternalCommandCodec, PostgresDurableCommandOutbox, commandPublication} from '../../adaptors/internal-commands.js';
import {HttpNotificationDelivery} from '../../adaptors/provider.js';
import {secret} from '../../foundation/secrets.js';
import {derivedId} from '../../foundation/identity.js';
import {identifier} from '../../foundation/domain.js';
import {RequestNotificationCommandHandler, DeliverNotificationCommandHandler,
  type RequestNotificationCommand, type DeliverNotificationCommand} from '../../contexts/communication/application/commands.js';
import {PickupOpenedIntegrationEventHandler, RewardIssuedIntegrationEventHandler,
  NotificationRequestedDomainEventHandler} from '../../contexts/communication/application/event-handlers.js';
import schema from '../../contexts/communication/adaptors/messaging/generated/internal_commands.json' with {type: 'json'};
import definitions from '../../contexts/communication/adaptors/messaging/subscriptions.json' with {type: 'json'};
import {commandSubscription, completeSubscriptions} from '../runtime.js';

/** Stable owner-local subscriptions; each installs a separate command queue. */
export enum CommunicationSubscription {
  PickupNotice = 'communication.pickup-notice', RewardNotice = 'communication.reward-notice',
  DeliverNotice = 'communication.deliver-notice',
}
function fields(value: unknown): Record<string, unknown> {
  if (!value || typeof value !== 'object' || Array.isArray(value)) throw new Error('Command object required');
  return value as Record<string, unknown>;
}
function noticeCodec(consumer: CommunicationSubscription) {
  return new InternalCommandCodec('communication', consumer, 'requestNotification', schema, (value): RequestNotificationCommand => {
    const p = fields(value);
    if (typeof p.recipient !== 'string' || typeof p.subject !== 'string' || typeof p.body !== 'string') throw new Error('Invalid notification');
    return {recipient: p.recipient, subject: p.subject, body: p.body};
  });
}
function deliveryCodec() {
  return new InternalCommandCodec('communication', CommunicationSubscription.DeliverNotice, 'deliverNotification', schema,
    (value): DeliverNotificationCommand => ({notificationId: identifier(String(fields(value).notificationId))}));
}

/** Typed provider bindings for the Communication context. */
export type CommunicationDependencies = {
    database: PostgresContextDatabase; notices: AggregateCommandPort<ReturnType<typeof restoreNotification>>;
    queries: PagedQueryPort<ReturnType<typeof restoreNotification>>; provider: DeliveryPort;
    request: RequestNotificationCommandHandler; deliver: DeliverNotificationCommandHandler;
    pickupCodec: InternalCommandCodec<RequestNotificationCommand>; rewardCodec: InternalCommandCodec<RequestNotificationCommand>;
    deliveryCodec: InternalCommandCodec<DeliverNotificationCommand>;
    pickup: PickupOpenedIntegrationEventHandler; reward: RewardIssuedIntegrationEventHandler; requested: NotificationRequestedDomainEventHandler;
  };

/** Owns Communication's resource lifetime and plain-constructor application graph. */
export function createCommunicationContainer(databaseURL: string, deliveryURL: string, deliveryKey: string) {
  const container = createContainer<CommunicationDependencies>({strict: true});
  container.register({
    database: asFunction(() => new PostgresContextDatabase('communication', databaseURL)).singleton().disposer(db => db.pool.end()),
    notices: asFunction((c: CommunicationDependencies) => new PostgresAggregateCommandStore(c.database, 'notification', restoreNotification)).singleton(),
    queries: asFunction((c: CommunicationDependencies) => new PostgresAggregateQueries(c.database, 'notification', restoreNotification)).singleton(),
    provider: asFunction(() => new HttpNotificationDelivery(deliveryURL, deliveryKey)).singleton(),
    request: asFunction((c: CommunicationDependencies) => new RequestNotificationCommandHandler(c.notices)).singleton(),
    deliver: asFunction((c: CommunicationDependencies) => new DeliverNotificationCommandHandler(c.notices, c.queries, c.provider)).singleton(),
    pickupCodec: asFunction(() => noticeCodec(CommunicationSubscription.PickupNotice)).singleton(),
    rewardCodec: asFunction(() => noticeCodec(CommunicationSubscription.RewardNotice)).singleton(),
    deliveryCodec: asFunction(deliveryCodec).singleton(),
    pickup: asFunction((c: CommunicationDependencies) => new PickupOpenedIntegrationEventHandler(new PostgresDurableCommandOutbox(c.database, c.pickupCodec))).singleton(),
    reward: asFunction((c: CommunicationDependencies) => new RewardIssuedIntegrationEventHandler(new PostgresDurableCommandOutbox(c.database, c.rewardCodec))).singleton(),
    requested: asFunction((c: CommunicationDependencies) => new NotificationRequestedDomainEventHandler(new PostgresDurableCommandOutbox(c.database, c.deliveryCodec))).singleton(),
  });
  return container;
}
export async function composeCommunication() {
  const container = createCommunicationContainer(secret('COMMUNICATION_DATABASE_URL'), secret('DELIVERY_URL'), secret('DELIVERY_KEY'));
  try {
  const c = container.cradle;
  return {database: c.database, queries: c.queries, commandHeader: commandPublication(schema, 'communication'),
    dispose: () => container.dispose(), subscriptions: completeSubscriptions('communication', definitions, [
      ...commandSubscription('collection.pickup-opened', c.pickupCodec, definitions, event => derivedId('pickup-notice', event.pickupId), c.pickup, c.request),
      ...commandSubscription('loyalty.reward-issued', c.rewardCodec, definitions, event => derivedId('reward-notice', event.rewardId), c.reward, c.request),
      ...commandSubscription('communication.notification-requested', c.deliveryCodec, definitions, event => event.notificationId, c.requested, c.deliver),
    ])};
  } catch (error) { await container.dispose(); throw error; }
}
