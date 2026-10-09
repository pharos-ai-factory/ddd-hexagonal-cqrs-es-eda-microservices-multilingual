import {notificationQueryEndpoints} from '../../contexts/communication/adaptors/notification-query-endpoints.js';
import {ListNotificationsQueryHandler} from '../../contexts/communication/application/queries/list-notifications.js';
import {GetNotificationQueryHandler} from '../../contexts/communication/application/queries/get-notification.js';
import {restoreNotification, PostgresNotificationReader} from '../../contexts/communication/adaptors/persistence/notifications.js';
import {noticeCodec, deliveryCodec, CommunicationSubscription} from '../../contexts/communication/adaptors/messaging/command-codecs.js';
import type {AggregateCommandPort, DeliveryPort} from '../../foundation/application.js';
import type {PagedQueryPort} from '../../foundation/pagination.js';
import {asFunction, createContainer} from 'awilix';
import {PostgresContextDatabase, PostgresAggregateCommandStore, PostgresAggregateQueries} from '../../adaptors/postgres.js';
import {InternalCommandCodec, PostgresDurableCommandOutbox, commandPublication} from '../../adaptors/internal-commands.js';
import {HttpNotificationDelivery} from '../../adaptors/provider.js';
import {secret} from '../../foundation/secrets.js';
import {derivedId} from '../../foundation/identity.js';
import {RequestNotificationCommandHandler, type RequestNotificationCommand} from '../../contexts/communication/application/commands/request-notification.js';
import {DeliverNotificationCommandHandler, type DeliverNotificationCommand} from '../../contexts/communication/application/commands/deliver-notification.js';
import {PickupOpenedIntegrationEventHandler} from '../../contexts/communication/application/event-handlers/pickup-opened.js';
import {RewardIssuedIntegrationEventHandler} from '../../contexts/communication/application/event-handlers/reward-issued.js';
import {NotificationRequestedDomainEventHandler} from '../../contexts/communication/application/event-handlers/notification-requested.js';
import schema from '../../contexts/communication/adaptors/messaging/generated/internal_commands.json' with {type: 'json'};
import definitions from '../../contexts/communication/adaptors/messaging/subscriptions.json' with {type: 'json'};
import {commandSubscription, completeSubscriptions} from '../runtime.js';

/** Typed provider bindings for the Communication context. */
export type CommunicationDependencies = {
    database: PostgresContextDatabase; notices: AggregateCommandPort<ReturnType<typeof restoreNotification>>;
    notificationReader: PostgresNotificationReader;
    getNotification: GetNotificationQueryHandler; listNotifications: ListNotificationsQueryHandler;
    queries: ReturnType<typeof notificationQueryEndpoints>; provider: DeliveryPort; deliveryReads: PagedQueryPort<ReturnType<typeof restoreNotification>>;
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
    notificationReader: asFunction((c: CommunicationDependencies) => new PostgresNotificationReader(c.database)).singleton(),
    getNotification: asFunction((c: CommunicationDependencies) => new GetNotificationQueryHandler(c.notificationReader)).singleton(),
    listNotifications: asFunction((c: CommunicationDependencies) => new ListNotificationsQueryHandler(c.notificationReader)).singleton(),
    queries: asFunction((c: CommunicationDependencies) => notificationQueryEndpoints(c.getNotification, c.listNotifications)).singleton(),
    deliveryReads: asFunction((c: CommunicationDependencies) => new PostgresAggregateQueries(c.database, 'notification', restoreNotification)).singleton(),
    provider: asFunction(() => new HttpNotificationDelivery(deliveryURL, deliveryKey)).singleton(),
    request: asFunction((c: CommunicationDependencies) => new RequestNotificationCommandHandler(c.notices)).singleton(),
    deliver: asFunction((c: CommunicationDependencies) => new DeliverNotificationCommandHandler(c.notices, c.deliveryReads, c.provider)).singleton(),
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
