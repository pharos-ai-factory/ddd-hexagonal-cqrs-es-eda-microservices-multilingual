import {
  CollectionIntegrationEvent,
  CommunicationDomainEvent,
  CommunicationSubscription,
  LoyaltyIntegrationEvent,
} from '../../adaptors/generated/message-names.js';
import * as notificationQueries from '../../contexts/communication/adaptors/messaging/notification-query-mapping.js';
import { CommunicationQuery } from '../../adaptors/generated/request-names.js';
import { CommandRegistration } from '../command-registration.js';
import { RabbitMQRequestRegistry } from '../../adaptors/messaging/rabbitmq-request-registry.js';
import { notificationPublications } from '../../contexts/communication/adaptors/messaging/notification-publications.js';
import { NotificationDeliveryPreparation } from '../../contexts/communication/adaptors/delivery/notification-delivery-preparation.js';
import type { Notification } from '../../contexts/communication/domain/notification.js';
import { PostgresNotificationWriteRepository } from '../../contexts/communication/adaptors/persistence/notification-write-repository.js';
import type { AggregateTransaction } from '../../adaptors/command-execution.js';
import { PostgresAggregateTransaction } from '../../adaptors/persistence/aggregate-transaction.js';
import { ListNotificationsQueryHandler } from '../../contexts/communication/application/queries/list-notifications.js';
import { GetNotificationQueryHandler } from '../../contexts/communication/application/queries/get-notification.js';
import { restoreNotificationSnapshot } from '../../contexts/communication/adaptors/persistence/notification-snapshot.js';
import { PostgresNotificationReadRepository } from '../../contexts/communication/adaptors/persistence/notification-read-repository.js';
import { noticeCodec } from '../../contexts/communication/adaptors/messaging/request-notification-codec.js';
import { deliveryCodec } from '../../contexts/communication/adaptors/messaging/deliver-notification-codec.js';
import type { DeliveryPort } from '../../foundation/application.js';
import type { PagedQueryPort } from '../../foundation/pagination.js';
import { asFunction, createContainer } from 'awilix';
import { PostgresContextDatabase } from '../../adaptors/persistence/postgres-database.js';
import { PostgresSnapshotReadRepository } from '../../adaptors/persistence/snapshot-read-repository.js';
import { InternalCommandCodec, commandPublication } from '../../adaptors/messaging/internal-command-codec.js';
import { PostgresDurableCommandOutbox } from '../../adaptors/persistence/command-outbox.js';
import { HttpNotificationDelivery } from '../../contexts/communication/adaptors/delivery/http-notification-delivery.js';
import { secret } from '../../foundation/secrets.js';
import { derivedId } from '../../foundation/identity.js';
import {
  RequestNotificationCommandHandler,
  type RequestNotificationCommand,
} from '../../contexts/communication/application/commands/request-notification.js';
import {
  DeliverNotificationCommandHandler,
  type DeliverNotificationCommand,
} from '../../contexts/communication/application/commands/deliver-notification.js';
import { PickupOpenedIntegrationEventHandler } from '../../contexts/communication/application/event-handlers/pickup-opened.js';
import { RewardIssuedIntegrationEventHandler } from '../../contexts/communication/application/event-handlers/reward-issued.js';
import { NotificationRequestedDomainEventHandler } from '../../contexts/communication/application/event-handlers/notification-requested.js';
import schema from '../../contexts/communication/adaptors/messaging/generated/internal_commands.json' with { type: 'json' };
import definitions from '../../contexts/communication/adaptors/messaging/subscriptions.json' with { type: 'json' };
import { completeSubscriptions } from '../runtime.js';

/** Typed provider bindings for the Communication context. */
export type CommunicationDependencies = {
  database: PostgresContextDatabase;
  notificationTransaction: AggregateTransaction<Notification>;
  notificationReadRepository: PostgresNotificationReadRepository;
  getNotification: GetNotificationQueryHandler;
  listNotifications: ListNotificationsQueryHandler;
  provider: DeliveryPort;
  notificationSnapshotReadRepository: PagedQueryPort<ReturnType<typeof restoreNotificationSnapshot>>;
  commands: CommandRegistration<Notification>;
  deliveryPreparation: NotificationDeliveryPreparation;
  pickupCodec: InternalCommandCodec<RequestNotificationCommand>;
  rewardCodec: InternalCommandCodec<RequestNotificationCommand>;
  deliveryCodec: InternalCommandCodec<DeliverNotificationCommand>;
  pickup: PickupOpenedIntegrationEventHandler;
  reward: RewardIssuedIntegrationEventHandler;
  requested: NotificationRequestedDomainEventHandler;
};

/** Owns Communication's resource lifetime and plain-constructor application graph. */
export function createCommunicationContainer(databaseURL: string, deliveryURL: string, deliveryKey: string) {
  const container = createContainer<CommunicationDependencies>({ strict: true });
  container.register({
    database: asFunction(() => new PostgresContextDatabase('communication', databaseURL))
      .singleton()
      .disposer((db) => db.pool.end()),
    notificationTransaction: asFunction(
      (c: CommunicationDependencies) =>
        new PostgresAggregateTransaction(
          c.database,
          'notification',
          restoreNotificationSnapshot,
          (source) => new PostgresNotificationWriteRepository(source),
          notificationPublications,
        ),
    ).singleton(),
    notificationReadRepository: asFunction(
      (c: CommunicationDependencies) => new PostgresNotificationReadRepository(c.database),
    ).singleton(),
    getNotification: asFunction(
      (c: CommunicationDependencies) => new GetNotificationQueryHandler(c.notificationReadRepository),
    ).singleton(),
    listNotifications: asFunction(
      (c: CommunicationDependencies) => new ListNotificationsQueryHandler(c.notificationReadRepository),
    ).singleton(),
    notificationSnapshotReadRepository: asFunction(
      (c: CommunicationDependencies) =>
        new PostgresSnapshotReadRepository(c.database, 'notification', restoreNotificationSnapshot),
    ).singleton(),
    provider: asFunction(() => new HttpNotificationDelivery(deliveryURL, deliveryKey)).singleton(),
    commands: asFunction(
      (c: CommunicationDependencies) => new CommandRegistration(c.notificationTransaction),
    ).singleton(),
    deliveryPreparation: asFunction(
      (c: CommunicationDependencies) =>
        new NotificationDeliveryPreparation(c.notificationSnapshotReadRepository, c.provider),
    ).singleton(),
    pickupCodec: asFunction(() => noticeCodec(CommunicationSubscription.PICKUP_NOTICE)).singleton(),
    rewardCodec: asFunction(() => noticeCodec(CommunicationSubscription.REWARD_NOTICE)).singleton(),
    deliveryCodec: asFunction(deliveryCodec).singleton(),
    pickup: asFunction(
      (c: CommunicationDependencies) =>
        new PickupOpenedIntegrationEventHandler(new PostgresDurableCommandOutbox(c.database, c.pickupCodec)),
    ).singleton(),
    reward: asFunction(
      (c: CommunicationDependencies) =>
        new RewardIssuedIntegrationEventHandler(new PostgresDurableCommandOutbox(c.database, c.rewardCodec)),
    ).singleton(),
    requested: asFunction(
      (c: CommunicationDependencies) =>
        new NotificationRequestedDomainEventHandler(
          new PostgresDurableCommandOutbox(c.database, c.deliveryCodec),
        ),
    ).singleton(),
  });
  return container;
}
export async function composeCommunication() {
  const container = createCommunicationContainer(
    secret('COMMUNICATION_DATABASE_URL'),
    secret('DELIVERY_URL'),
    secret('DELIVERY_KEY'),
  );
  try {
    const c = container.cradle;
    return {
      database: c.database,
      requests: communicationRequests(c),
      commandHeader: commandPublication(schema, 'communication'),
      dispose: () => container.dispose(),
      subscriptions: communicationSubscriptions(c),
    };
  } catch (error) {
    await container.dispose();
    throw error;
  }
}

/** Communication exposes query requests and receives commands through durable subscriptions. */
export function communicationRequests(c: CommunicationDependencies) {
  const registry = new RabbitMQRequestRegistry('communication');
  registry.query(
    CommunicationQuery.GET_NOTIFICATION,
    notificationQueries.getNotification,
    c.getNotification,
    notificationQueries.getReply,
  );
  registry.query(
    CommunicationQuery.LIST_NOTIFICATIONS,
    notificationQueries.listNotifications,
    c.listNotifications,
    notificationQueries.listReply,
  );
  return registry;
}

/** Register each private command codec alongside its concrete handler factory. */
export function communicationSubscriptions(c: CommunicationDependencies) {
  const preparation = c.deliveryPreparation;
  return completeSubscriptions('communication', definitions, [
    ...c.commands.subscription(
      CollectionIntegrationEvent.PICKUP_OPENED,
      c.pickupCodec,
      definitions,
      (event) => derivedId('pickup-notice', event.pickupId),
      c.pickup,
      (repository, realtime) => new RequestNotificationCommandHandler(repository, realtime),
      (m, command) => ({ ...m, input: command }),
    ),
    ...c.commands.subscription(
      LoyaltyIntegrationEvent.REWARD_ISSUED,
      c.rewardCodec,
      definitions,
      (event) => derivedId('reward-notice', event.rewardId),
      c.reward,
      (repository, realtime) => new RequestNotificationCommandHandler(repository, realtime),
      (m, command) => ({ ...m, input: command }),
    ),
    ...c.commands.preparedSubscription(
      CommunicationDomainEvent.NOTIFICATION_REQUESTED,
      c.deliveryCodec,
      definitions,
      (event) => event.notificationId,
      c.requested,
      (command) => preparation.prepare(command),
      (repository, receipt, realtime) => new DeliverNotificationCommandHandler(repository, realtime, receipt),
    ),
  ]);
}
