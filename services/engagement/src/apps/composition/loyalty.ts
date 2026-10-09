import {bindCommand, type CommandExecutor} from '../../adaptors/command-execution.js';
import {accountPublications} from '../../contexts/loyalty/adaptors/messaging/account-publications.js';
import {rewardPublications} from '../../contexts/loyalty/adaptors/messaging/reward-publications.js';
import type {Reward} from '../../contexts/loyalty/domain/reward.js';
import {PostgresRewardWriteRepository} from '../../contexts/loyalty/adaptors/persistence/reward-write-repository.js';
import type {LoyaltyAccount} from '../../contexts/loyalty/domain/loyalty-account.js';
import {PostgresAccountWriteRepository} from '../../contexts/loyalty/adaptors/persistence/account-write-repository.js';
import type {AggregateTransaction} from '../../adaptors/command-execution.js';
import {PostgresAggregateTransaction} from '../../adaptors/aggregate-transaction.js';
import {rewardQueryEndpoints} from '../../contexts/loyalty/adaptors/reward-query-endpoints.js';
import {ListRewardsQueryHandler} from '../../contexts/loyalty/application/queries/list-rewards.js';
import {GetRewardQueryHandler} from '../../contexts/loyalty/application/queries/get-reward.js';
import {restoreRewardSnapshot} from '../../contexts/loyalty/adaptors/persistence/reward-snapshot.js';
import {PostgresRewardReadRepository} from '../../contexts/loyalty/adaptors/persistence/reward-read-repository.js';
import {accountQueryEndpoints} from '../../contexts/loyalty/adaptors/account-query-endpoints.js';
import {ListAccountsQueryHandler} from '../../contexts/loyalty/application/queries/list-accounts.js';
import {GetAccountQueryHandler} from '../../contexts/loyalty/application/queries/get-account.js';
import {restoreAccountSnapshot} from '../../contexts/loyalty/adaptors/persistence/account-snapshot.js';
import {PostgresAccountReadRepository} from '../../contexts/loyalty/adaptors/persistence/account-read-repository.js';
import {creditCodec, issueCodec} from '../../contexts/loyalty/adaptors/messaging/command-codecs.js';
import {asFunction, asValue, createContainer} from 'awilix';
import {PostgresContextDatabase} from '../../adaptors/postgres.js';
import {InternalCommandCodec, PostgresDurableCommandOutbox, commandPublication} from '../../adaptors/internal-commands.js';
import {derivedId} from '../../foundation/identity.js';
import {secret} from '../../foundation/secrets.js';
import {CreditCollectionCommandHandler, type CreditCollectionCommand} from '../../contexts/loyalty/application/commands/credit-collection.js';
import {IssueRewardCommandHandler, type IssueRewardCommand} from '../../contexts/loyalty/application/commands/issue-reward.js';
import {RedeemRewardCommandHandler, type RedeemRewardCommand} from '../../contexts/loyalty/application/commands/redeem-reward.js';
import {OrderCollectedIntegrationEventHandler} from '../../contexts/loyalty/application/event-handlers/order-collected.js';
import {RewardEarnedDomainEventHandler} from '../../contexts/loyalty/application/event-handlers/reward-earned.js';
import schema from '../../contexts/loyalty/adaptors/messaging/generated/internal_commands.json' with {type: 'json'};
import definitions from '../../contexts/loyalty/adaptors/messaging/subscriptions.json' with {type: 'json'};
import {commandSubscription, completeSubscriptions} from '../runtime.js';

/** Typed provider bindings for the Loyalty context. */
export type LoyaltyDependencies = {
    database: PostgresContextDatabase;
    accountTransaction: AggregateTransaction<LoyaltyAccount>;
    rewardTransaction: AggregateTransaction<Reward>;
    accountReadRepository: PostgresAccountReadRepository;
    getAccount: GetAccountQueryHandler; listAccounts: ListAccountsQueryHandler;
    accountQueries: ReturnType<typeof accountQueryEndpoints>;
    rewardReadRepository: PostgresRewardReadRepository;
    getReward: GetRewardQueryHandler; listRewards: ListRewardsQueryHandler;
    rewardQueries: ReturnType<typeof rewardQueryEndpoints>;
    credit: CommandExecutor<CreditCollectionCommand>; issue: CommandExecutor<IssueRewardCommand>; redeem: CommandExecutor<RedeemRewardCommand>;
    creditCodec: InternalCommandCodec<CreditCollectionCommand>; issueCodec: InternalCommandCodec<IssueRewardCommand>;
    collected: OrderCollectedIntegrationEventHandler; earned: RewardEarnedDomainEventHandler; clock: () => Date;
  };

/** Explicit factories keep Awilix and PostgreSQL out of application constructors. */
export function createLoyaltyContainer(databaseURL: string, clock: () => Date = () => new Date()) {
  const container = createContainer<LoyaltyDependencies>({strict: true});
  container.register({
    database: asFunction(() => new PostgresContextDatabase('loyalty', databaseURL)).singleton().disposer(db => db.pool.end()),
    clock: asValue(clock), creditCodec: asFunction(creditCodec).singleton(), issueCodec: asFunction(issueCodec).singleton(),
    accountTransaction: asFunction((c: LoyaltyDependencies) => new PostgresAggregateTransaction(c.database, 'account', restoreAccountSnapshot, source => new PostgresAccountWriteRepository(source), accountPublications)).singleton(),
    rewardTransaction: asFunction((c: LoyaltyDependencies) => new PostgresAggregateTransaction(c.database, 'reward', restoreRewardSnapshot, source => new PostgresRewardWriteRepository(source), rewardPublications)).singleton(),
    accountReadRepository: asFunction((c: LoyaltyDependencies) => new PostgresAccountReadRepository(c.database)).singleton(),
    getAccount: asFunction((c: LoyaltyDependencies) => new GetAccountQueryHandler(c.accountReadRepository)).singleton(),
    listAccounts: asFunction((c: LoyaltyDependencies) => new ListAccountsQueryHandler(c.accountReadRepository)).singleton(),
    accountQueries: asFunction((c: LoyaltyDependencies) => accountQueryEndpoints(c.getAccount, c.listAccounts)).singleton(),
    rewardReadRepository: asFunction((c: LoyaltyDependencies) => new PostgresRewardReadRepository(c.database)).singleton(),
    getReward: asFunction((c: LoyaltyDependencies) => new GetRewardQueryHandler(c.rewardReadRepository)).singleton(),
    listRewards: asFunction((c: LoyaltyDependencies) => new ListRewardsQueryHandler(c.rewardReadRepository)).singleton(),
    rewardQueries: asFunction((c: LoyaltyDependencies) => rewardQueryEndpoints(c.getReward, c.listRewards)).singleton(),
    credit: asFunction((c: LoyaltyDependencies) => bindCommand(c.accountTransaction, repository => new CreditCollectionCommandHandler(repository, derivedId), (m, command) => ({...m, id: derivedId('collection-credit', command.orderId)}))).singleton(),
    issue: asFunction((c: LoyaltyDependencies) => bindCommand(c.rewardTransaction, repository => new IssueRewardCommandHandler(repository, c.clock), (m, command) => ({...m, id: derivedId('issue-reward', command.grantId)}))).singleton(),
    redeem: asFunction((c: LoyaltyDependencies) => bindCommand(c.rewardTransaction, repository => new RedeemRewardCommandHandler(repository, c.clock))).singleton(),
    collected: asFunction((c: LoyaltyDependencies) => new OrderCollectedIntegrationEventHandler(new PostgresDurableCommandOutbox(c.database, c.creditCodec))).singleton(),
    earned: asFunction((c: LoyaltyDependencies) => new RewardEarnedDomainEventHandler(new PostgresDurableCommandOutbox(c.database, c.issueCodec))).singleton(),
  });
  return container;
}
export async function composeLoyalty() {
  const container = createLoyaltyContainer(secret('LOYALTY_DATABASE_URL'));
  try {
  const c = container.cradle;
  return {database: c.database, accountQueries: c.accountQueries, rewardQueries: c.rewardQueries, redeem: c.redeem,
    commandHeader: commandPublication(schema, 'loyalty'), dispose: () => container.dispose(), subscriptions: completeSubscriptions('loyalty', definitions, [
      ...commandSubscription('collection.order-collected', c.creditCodec, definitions, event => event.customerId, c.collected, c.credit),
      ...commandSubscription('loyalty.reward-earned', c.issueCodec, definitions, event => derivedId('reward', event.grantId), c.earned, c.issue),
    ])};
  } catch (error) { await container.dispose(); throw error; }
}
