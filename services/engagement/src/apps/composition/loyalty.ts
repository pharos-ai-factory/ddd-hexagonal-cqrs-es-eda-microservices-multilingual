import {rewardQueryEndpoints} from '../../contexts/loyalty/adaptors/reward-query-endpoints.js';
import {ListRewardsQueryHandler} from '../../contexts/loyalty/application/queries/list-rewards.js';
import {GetRewardQueryHandler} from '../../contexts/loyalty/application/queries/get-reward.js';
import {restoreReward, PostgresRewardReader} from '../../contexts/loyalty/adaptors/persistence/rewards.js';
import {accountQueryEndpoints} from '../../contexts/loyalty/adaptors/account-query-endpoints.js';
import {ListAccountsQueryHandler} from '../../contexts/loyalty/application/queries/list-accounts.js';
import {GetAccountQueryHandler} from '../../contexts/loyalty/application/queries/get-account.js';
import {restoreAccount, PostgresAccountReader} from '../../contexts/loyalty/adaptors/persistence/accounts.js';
import {creditCodec, issueCodec} from '../../contexts/loyalty/adaptors/messaging/command-codecs.js';
import type {AggregateCommandPort} from '../../foundation/application.js';
import {asFunction, asValue, createContainer} from 'awilix';
import {PostgresContextDatabase, PostgresAggregateCommandStore} from '../../adaptors/postgres.js';
import {InternalCommandCodec, PostgresDurableCommandOutbox, commandPublication} from '../../adaptors/internal-commands.js';
import {derivedId} from '../../foundation/identity.js';
import {secret} from '../../foundation/secrets.js';
import {CreditCollectionCommandHandler, type CreditCollectionCommand} from '../../contexts/loyalty/application/commands/credit-collection.js';
import {IssueRewardCommandHandler, type IssueRewardCommand} from '../../contexts/loyalty/application/commands/issue-reward.js';
import {RedeemRewardCommandHandler} from '../../contexts/loyalty/application/commands/redeem-reward.js';
import {OrderCollectedIntegrationEventHandler} from '../../contexts/loyalty/application/event-handlers/order-collected.js';
import {RewardEarnedDomainEventHandler} from '../../contexts/loyalty/application/event-handlers/reward-earned.js';
import schema from '../../contexts/loyalty/adaptors/messaging/generated/internal_commands.json' with {type: 'json'};
import definitions from '../../contexts/loyalty/adaptors/messaging/subscriptions.json' with {type: 'json'};
import {commandSubscription, completeSubscriptions} from '../runtime.js';

/** Typed provider bindings for the Loyalty context. */
export type LoyaltyDependencies = {
    database: PostgresContextDatabase;
    accounts: AggregateCommandPort<ReturnType<typeof restoreAccount>>;
    rewards: AggregateCommandPort<ReturnType<typeof restoreReward>>;
    accountReader: PostgresAccountReader;
    getAccount: GetAccountQueryHandler; listAccounts: ListAccountsQueryHandler;
    accountQueries: ReturnType<typeof accountQueryEndpoints>;
    rewardReader: PostgresRewardReader;
    getReward: GetRewardQueryHandler; listRewards: ListRewardsQueryHandler;
    rewardQueries: ReturnType<typeof rewardQueryEndpoints>;
    credit: CreditCollectionCommandHandler; issue: IssueRewardCommandHandler; redeem: RedeemRewardCommandHandler;
    creditCodec: InternalCommandCodec<CreditCollectionCommand>; issueCodec: InternalCommandCodec<IssueRewardCommand>;
    collected: OrderCollectedIntegrationEventHandler; earned: RewardEarnedDomainEventHandler; clock: () => Date;
  };

/** Explicit factories keep Awilix and PostgreSQL out of application constructors. */
export function createLoyaltyContainer(databaseURL: string, clock: () => Date = () => new Date()) {
  const container = createContainer<LoyaltyDependencies>({strict: true});
  container.register({
    database: asFunction(() => new PostgresContextDatabase('loyalty', databaseURL)).singleton().disposer(db => db.pool.end()),
    clock: asValue(clock), creditCodec: asFunction(creditCodec).singleton(), issueCodec: asFunction(issueCodec).singleton(),
    accounts: asFunction((c: LoyaltyDependencies) => new PostgresAggregateCommandStore(c.database, 'account', restoreAccount)).singleton(),
    rewards: asFunction((c: LoyaltyDependencies) => new PostgresAggregateCommandStore(c.database, 'reward', restoreReward)).singleton(),
    accountReader: asFunction((c: LoyaltyDependencies) => new PostgresAccountReader(c.database)).singleton(),
    getAccount: asFunction((c: LoyaltyDependencies) => new GetAccountQueryHandler(c.accountReader)).singleton(),
    listAccounts: asFunction((c: LoyaltyDependencies) => new ListAccountsQueryHandler(c.accountReader)).singleton(),
    accountQueries: asFunction((c: LoyaltyDependencies) => accountQueryEndpoints(c.getAccount, c.listAccounts)).singleton(),
    rewardReader: asFunction((c: LoyaltyDependencies) => new PostgresRewardReader(c.database)).singleton(),
    getReward: asFunction((c: LoyaltyDependencies) => new GetRewardQueryHandler(c.rewardReader)).singleton(),
    listRewards: asFunction((c: LoyaltyDependencies) => new ListRewardsQueryHandler(c.rewardReader)).singleton(),
    rewardQueries: asFunction((c: LoyaltyDependencies) => rewardQueryEndpoints(c.getReward, c.listRewards)).singleton(),
    credit: asFunction((c: LoyaltyDependencies) => new CreditCollectionCommandHandler(c.accounts, derivedId)).singleton(),
    issue: asFunction((c: LoyaltyDependencies) => new IssueRewardCommandHandler(c.rewards, derivedId, c.clock)).singleton(),
    redeem: asFunction((c: LoyaltyDependencies) => new RedeemRewardCommandHandler(c.rewards, c.clock)).singleton(),
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
