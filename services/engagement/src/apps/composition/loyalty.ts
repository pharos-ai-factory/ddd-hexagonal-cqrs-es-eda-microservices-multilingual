import { PostgresRealtimeEventBus } from '../../adaptors/persistence/realtime-event-bus.js';
import { RewardIssuedRealtimeIntegrationEventHandler } from '../../contexts/loyalty/application/event-handlers/realtime-reward-issued.js';
import { realtimeSubscription } from '../runtime.js';
import {
  CollectionIntegrationEvent,
  LoyaltyDomainEvent,
  LoyaltyIntegrationEvent,
  LoyaltySubscription,
} from '../../adaptors/generated/message-names.js';
import * as rewardQueries from '../../contexts/loyalty/adaptors/messaging/reward-query-mapping.js';
import * as accountQueries from '../../contexts/loyalty/adaptors/messaging/account-query-mapping.js';
import { LoyaltyQuery, LoyaltyCommand } from '../../adaptors/generated/request-names.js';
import { PostgresPublicationOutbox } from '../../adaptors/persistence/integration-event-outbox.js';
import { published } from '../../adaptors/generated/event-types.js';
import { RewardIssuedDomainEventHandler } from '../../contexts/loyalty/application/event-handlers/publish-reward-issued.js';
import { CommandRegistration } from '../command-registration.js';
import { RabbitMQRequestRegistry } from '../../adaptors/messaging/rabbitmq-request-registry.js';
import * as wire from '../../contexts/loyalty/adaptors/messaging/reward-requests.js';
import { accountPublications } from '../../contexts/loyalty/adaptors/messaging/account-publications.js';
import { rewardPublications } from '../../contexts/loyalty/adaptors/messaging/reward-publications.js';
import type { Reward } from '../../contexts/loyalty/domain/reward.js';
import { PostgresRewardWriteRepository } from '../../contexts/loyalty/adaptors/persistence/reward-write-repository.js';
import type { LoyaltyAccount } from '../../contexts/loyalty/domain/loyalty-account.js';
import { PostgresAccountWriteRepository } from '../../contexts/loyalty/adaptors/persistence/account-write-repository.js';
import type { AggregateTransaction } from '../../adaptors/command-execution.js';
import { PostgresAggregateTransaction } from '../../adaptors/persistence/aggregate-transaction.js';
import { ListRewardsQueryHandler } from '../../contexts/loyalty/application/queries/list-rewards.js';
import { GetRewardQueryHandler } from '../../contexts/loyalty/application/queries/get-reward.js';
import { restoreRewardSnapshot } from '../../contexts/loyalty/adaptors/persistence/reward-snapshot.js';
import { PostgresRewardReadRepository } from '../../contexts/loyalty/adaptors/persistence/reward-read-repository.js';
import { ListAccountsQueryHandler } from '../../contexts/loyalty/application/queries/list-accounts.js';
import { GetAccountQueryHandler } from '../../contexts/loyalty/application/queries/get-account.js';
import { restoreAccountSnapshot } from '../../contexts/loyalty/adaptors/persistence/account-snapshot.js';
import { PostgresAccountReadRepository } from '../../contexts/loyalty/adaptors/persistence/account-read-repository.js';
import { creditCodec } from '../../contexts/loyalty/adaptors/messaging/credit-collection-codec.js';
import { issueCodec } from '../../contexts/loyalty/adaptors/messaging/issue-reward-codec.js';
import { asFunction, asValue, createContainer } from 'awilix';
import { PostgresContextDatabase } from '../../adaptors/persistence/postgres-database.js';
import { InternalCommandCodec, commandPublication } from '../../adaptors/messaging/internal-command-codec.js';
import { PostgresDurableCommandOutbox } from '../../adaptors/persistence/command-outbox.js';
import { derivedId } from '../../foundation/identity.js';
import { secret } from '../../foundation/secrets.js';
import {
  CreditCollectionCommandHandler,
  type CreditCollectionCommand,
} from '../../contexts/loyalty/application/commands/credit-collection.js';
import {
  IssueRewardCommandHandler,
  type IssueRewardCommand,
} from '../../contexts/loyalty/application/commands/issue-reward.js';
import { RedeemRewardCommandHandler } from '../../contexts/loyalty/application/commands/redeem-reward.js';
import { OrderCollectedIntegrationEventHandler } from '../../contexts/loyalty/application/event-handlers/order-collected.js';
import { RewardEarnedDomainEventHandler } from '../../contexts/loyalty/application/event-handlers/reward-earned.js';
import schema from '../../contexts/loyalty/adaptors/messaging/generated/internal_commands.json' with { type: 'json' };
import definitions from '../../contexts/loyalty/adaptors/messaging/subscriptions.json' with { type: 'json' };
import { completeSubscriptions, publicationSubscription } from '../runtime.js';

/** Typed provider bindings for the Loyalty context. */
export type LoyaltyDependencies = {
  database: PostgresContextDatabase;
  accountTransaction: AggregateTransaction<LoyaltyAccount>;
  rewardTransaction: AggregateTransaction<Reward>;
  accountReadRepository: PostgresAccountReadRepository;
  getAccount: GetAccountQueryHandler;
  listAccounts: ListAccountsQueryHandler;
  rewardReadRepository: PostgresRewardReadRepository;
  getReward: GetRewardQueryHandler;
  listRewards: ListRewardsQueryHandler;
  accountCommands: CommandRegistration<LoyaltyAccount>;
  rewardCommands: CommandRegistration<Reward>;
  creditCodec: InternalCommandCodec<CreditCollectionCommand>;
  issueCodec: InternalCommandCodec<IssueRewardCommand>;
  collected: OrderCollectedIntegrationEventHandler;
  earned: RewardEarnedDomainEventHandler;
  publishIssued: RewardIssuedDomainEventHandler;
  realtimeIssued: RewardIssuedRealtimeIntegrationEventHandler;
  clock: () => Date;
};

/** Explicit factories keep Awilix and PostgreSQL out of application constructors. */
export function createLoyaltyContainer(databaseURL: string, clock: () => Date = () => new Date()) {
  const container = createContainer<LoyaltyDependencies>({ strict: true });
  container.register({
    database: asFunction(() => new PostgresContextDatabase('loyalty', databaseURL))
      .singleton()
      .disposer((db) => db.pool.end()),
    clock: asValue(clock),
    creditCodec: asFunction(creditCodec).singleton(),
    issueCodec: asFunction(issueCodec).singleton(),
    accountTransaction: asFunction(
      (c: LoyaltyDependencies) =>
        new PostgresAggregateTransaction(
          c.database,
          'account',
          restoreAccountSnapshot,
          (source) => new PostgresAccountWriteRepository(source),
          accountPublications,
        ),
    ).singleton(),
    rewardTransaction: asFunction(
      (c: LoyaltyDependencies) =>
        new PostgresAggregateTransaction(
          c.database,
          'reward',
          restoreRewardSnapshot,
          (source) => new PostgresRewardWriteRepository(source),
          rewardPublications,
        ),
    ).singleton(),
    accountReadRepository: asFunction(
      (c: LoyaltyDependencies) => new PostgresAccountReadRepository(c.database),
    ).singleton(),
    getAccount: asFunction(
      (c: LoyaltyDependencies) => new GetAccountQueryHandler(c.accountReadRepository),
    ).singleton(),
    listAccounts: asFunction(
      (c: LoyaltyDependencies) => new ListAccountsQueryHandler(c.accountReadRepository),
    ).singleton(),
    rewardReadRepository: asFunction(
      (c: LoyaltyDependencies) => new PostgresRewardReadRepository(c.database),
    ).singleton(),
    getReward: asFunction(
      (c: LoyaltyDependencies) => new GetRewardQueryHandler(c.rewardReadRepository),
    ).singleton(),
    listRewards: asFunction(
      (c: LoyaltyDependencies) => new ListRewardsQueryHandler(c.rewardReadRepository),
    ).singleton(),
    accountCommands: asFunction(
      (c: LoyaltyDependencies) => new CommandRegistration(c.accountTransaction),
    ).singleton(),
    rewardCommands: asFunction(
      (c: LoyaltyDependencies) => new CommandRegistration(c.rewardTransaction),
    ).singleton(),
    collected: asFunction(
      (c: LoyaltyDependencies) =>
        new OrderCollectedIntegrationEventHandler(
          new PostgresDurableCommandOutbox(c.database, c.creditCodec),
        ),
    ).singleton(),
    realtimeIssued: asFunction(
      (c: LoyaltyDependencies) =>
        new RewardIssuedRealtimeIntegrationEventHandler(
          c.rewardReadRepository,
          new PostgresRealtimeEventBus(c.database, LoyaltySubscription.REALTIME_REWARD_ISSUED, 'reward'),
        ),
    ).singleton(),
    publishIssued: asFunction(
      (c: LoyaltyDependencies) =>
        new RewardIssuedDomainEventHandler(
          new PostgresPublicationOutbox(
            c.database,
            LoyaltySubscription.PUBLISH_REWARD_ISSUED,
            'reward',
            (payload) => published(LoyaltyIntegrationEvent.REWARD_ISSUED, payload),
          ),
        ),
    ).singleton(),
    earned: asFunction(
      (c: LoyaltyDependencies) =>
        new RewardEarnedDomainEventHandler(new PostgresDurableCommandOutbox(c.database, c.issueCodec)),
    ).singleton(),
  });
  return container;
}
export async function composeLoyalty() {
  const container = createLoyaltyContainer(secret('LOYALTY_DATABASE_URL'));
  try {
    const c = container.cradle;
    return {
      database: c.database,
      requests: loyaltyRequests(c),
      commandHeader: commandPublication(schema, 'loyalty'),
      dispose: () => container.dispose(),
      subscriptions: loyaltySubscriptions(c),
    };
  } catch (error) {
    await container.dispose();
    throw error;
  }
}

/** Keep each published command, parser and handler together at the composition boundary. */
export function loyaltyRequests(c: LoyaltyDependencies) {
  const registry = new RabbitMQRequestRegistry('loyalty');
  const clock = c.clock;
  c.rewardCommands.request(
    registry,
    LoyaltyCommand.REDEEM_REWARD,
    wire.redeem,
    (repository, realtime) => new RedeemRewardCommandHandler(repository, realtime, clock),
  );
  registry.query(LoyaltyQuery.GET_ACCOUNT, accountQueries.getAccount, c.getAccount, accountQueries.getReply);
  registry.query(
    LoyaltyQuery.LIST_ACCOUNTS,
    accountQueries.listAccounts,
    c.listAccounts,
    accountQueries.listReply,
  );
  registry.query(LoyaltyQuery.GET_REWARD, rewardQueries.getReward, c.getReward, rewardQueries.getReply);
  registry.query(
    LoyaltyQuery.LIST_REWARDS,
    rewardQueries.listRewards,
    c.listRewards,
    rewardQueries.listReply,
  );
  return registry;
}

/** Resolve durable reactions and their directly registered command handlers before workers start. */
export function loyaltySubscriptions(c: LoyaltyDependencies) {
  const clock = c.clock;
  return completeSubscriptions('loyalty', definitions, [
    realtimeSubscription(
      'loyalty',
      LoyaltySubscription.REALTIME_REWARD_ISSUED,
      LoyaltyIntegrationEvent.REWARD_ISSUED,
      'reward',
      definitions,
      c.realtimeIssued,
    ),
    publicationSubscription(
      'loyalty',
      LoyaltySubscription.PUBLISH_REWARD_ISSUED,
      LoyaltyDomainEvent.REWARD_ISSUED,
      LoyaltyIntegrationEvent.REWARD_ISSUED,
      definitions,
      c.publishIssued,
    ),
    ...c.accountCommands.subscription(
      CollectionIntegrationEvent.ORDER_COLLECTED,
      c.creditCodec,
      definitions,
      (event) => event.customerId,
      c.collected,
      (repository, realtime) => new CreditCollectionCommandHandler(repository, realtime, derivedId),
      (m, command) => ({ ...m, id: derivedId('collection-credit', command.orderId) }),
    ),
    ...c.rewardCommands.subscription(
      LoyaltyDomainEvent.REWARD_EARNED,
      c.issueCodec,
      definitions,
      (event) => derivedId('reward', event.grantId),
      c.earned,
      (repository) => new IssueRewardCommandHandler(repository, clock),
      (m, command) => ({ ...m, id: derivedId('issue-reward', command.grantId) }),
    ),
  ]);
}
