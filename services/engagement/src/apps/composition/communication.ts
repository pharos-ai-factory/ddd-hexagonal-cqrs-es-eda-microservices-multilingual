import {bindCommand, type CommandExecutor} from '../../adaptors/command-execution.js';
import {notificationPublications} from '../../contexts/communication/adaptors/messaging/notification-publications.js';
import {NotificationDeliveryCommandExecutor} from '../../contexts/communication/adaptors/persistence/notification-delivery-command-executor.js';
import type {Notification} from '../../contexts/communication/domain/notification.js';
import {PostgresNotificationWriteRepository} from '../../contexts/communication/adaptors/persistence/notification-write-repository.js';
import type {AggregateTransaction} from '../../adaptors/command-execution.js';
import {PostgresAggregateTransaction} from '../../adaptors/aggregate-transaction.js';
import {notificationQueryEndpoints} from '../../contexts/communication/adaptors/notification-query-endpoints.js';
import {ListNotificationsQueryHandler} from '../../contexts/communication/application/queries/list-notifications.js';
import {GetNotificationQueryHandler} from '../../contexts/communication/application/queries/get-notification.js';
import {restoreNotificationSnapshot} from '../../contexts/communication/adaptors/persistence/notification-snapshot.js';
import {PostgresNotificationReadRepository} from '../../contexts/communication/adaptors/persistence/notification-read-repository.js';
import {noticeCodec, deliveryCodec, CommunicationSubscription} from '../../contexts/communication/adaptors/messaging/command-codecs.js';
import type {DeliveryPort} from '../../foundation/application.js';
import type {PagedQueryPort} from '../../foundation/pagination.js';
import {asFunction, createContainer} from 'awilix';
import {PostgresContextDatabase} from '../../adaptors/postgres.js';
import {PostgresSnapshotReadRepository} from '../../adaptors/snapshot-read-repository.js';
import {InternalCommandCodec, PostgresDurableCommandOutbox, commandPublication} from '../../adaptors/internal-commands.js';
import {HttpNotificationDelivery} from '../../adaptors/provider.js';
import {secret} from '../../foundation/secrets.js';
import {derivedId} from '../../foundation/identity.js';
import {RequestNotificationCommandHandler, type RequestNotificationCommand} from '../../contexts/communication/application/commands/request-notification.js';
import {type DeliverNotificationCommand} from '../../contexts/communication/application/commands/deliver-notification.js';
import {PickupOpenedIntegrationEventHandler} from '../../contexts/communication/application/event-handlers/pickup-opened.js';
import {RewardIssuedIntegrationEventHandler} from '../../contexts/communication/application/event-handlers/reward-issued.js';
import {NotificationRequestedDomainEventHandler} from '../../contexts/communication/application/event-handlers/notification-requested.js';
import schema from '../../contexts/communication/adaptors/messaging/generated/internal_commands.json' with {type: 'json'};
import definitions from '../../contexts/communication/adaptors/messaging/subscriptions.json' with {type: 'json'};
import {commandSubscription, completeSubscriptions} from '../runtime.js';

/** Typed provider bindings for the Communication context. */
export type CommunicationDependencies = {
    database: PostgresContextDatabase; notificationTransaction: AggregateTransaction<Notification>;
    notificationReadRepository: PostgresNotificationReadRepository;
    getNotification: GetNotificationQueryHandler; listNotifications: ListNotificationsQueryHandler;
    queries: ReturnType<typeof notificationQueryEndpoints>; provider: DeliveryPort; notificationSnapshotReadRepository: PagedQueryPort<ReturnType<typeof restoreNotificationSnapshot>>;
    request: CommandExecutor<RequestNotificationCommand>; deliver: CommandExecutor<DeliverNotificationCommand>;
    pickupCodec: InternalCommandCodec<RequestNotificationCommand>; rewardCodec: InternalCommandCodec<RequestNotificationCommand>;
    deliveryCodec: InternalCommandCodec<DeliverNotificationCommand>;
    pickup: PickupOpenedIntegrationEventHandler; reward: RewardIssuedIntegrationEventHandler; requested: NotificationRequestedDomainEventHandler;
  };

/** Owns Communication's resource lifetime and plain-constructor application graph. */
export function createCommunicationContainer(databaseURL: string, deliveryURL: string, deliveryKey: string) {
  const container = createContainer<CommunicationDependencies>({strict: true});
  container.register({
    database: asFunction(() => new PostgresContextDatabase('communication', databaseURL)).singleton().disposer(db => db.pool.end()),
    notificationTransaction: asFunction((c: CommunicationDependencies) => new PostgresAggregateTransaction(c.database, 'notification', restoreNotificationSnapshot, source => new PostgresNotificationWriteRepository(source), notificationPublications)).singleton(),
    notificationReadRepository: asFunction((c: CommunicationDependencies) => new PostgresNotificationReadRepository(c.database)).singleton(),
    getNotification: asFunction((c: CommunicationDependencies) => new GetNotificationQueryHandler(c.notificationReadRepository)).singleton(),
    listNotifications: asFunction((c: CommunicationDependencies) => new ListNotificationsQueryHandler(c.notificationReadRepository)).singleton(),
    queries: asFunction((c: CommunicationDependencies) => notificationQueryEndpoints(c.getNotification, c.listNotifications)).singleton(),
    notificationSnapshotReadRepository: asFunction((c: CommunicationDependencies) => new PostgresSnapshotReadRepository(c.database, 'notification', restoreNotificationSnapshot)).singleton(),
    provider: asFunction(() => new HttpNotificationDelivery(deliveryURL, deliveryKey)).singleton(),
    request: asFunction((c: CommunicationDependencies) => bindCommand(c.notificationTransaction, repository => new RequestNotificationCommandHandler(repository), (m, command) => ({...m, input: command}))).singleton(),
    deliver: asFunction((c: CommunicationDependencies) => new NotificationDeliveryCommandExecutor(c.notificationTransaction, c.notificationSnapshotReadRepository, c.provider)).singleton(),
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
