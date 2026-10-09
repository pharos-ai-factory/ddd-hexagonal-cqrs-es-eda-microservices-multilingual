import type {AggregateCommandPort} from '../../foundation/application.js';
import type {PagedQueryPort} from '../../foundation/pagination.js';
import {asFunction, asValue, createContainer} from 'awilix';
import {PostgresContextDatabase, PostgresAggregateCommandStore, PostgresAggregateQueries} from '../../adaptors/postgres.js';
import {restoreAccount, restoreReward} from '../../adaptors/restore.js';
import {InternalCommandCodec, PostgresDurableCommandOutbox, commandPublication} from '../../adaptors/internal-commands.js';
import {derivedId} from '../../foundation/identity.js';
import {secret} from '../../foundation/secrets.js';
import {identifier} from '../../foundation/domain.js';
import {CreditCollectionCommandHandler, IssueRewardCommandHandler, RedeemRewardCommandHandler,
  type CreditCollectionCommand, type IssueRewardCommand} from '../../contexts/loyalty/application/commands.js';
import {OrderCollectedIntegrationEventHandler, RewardEarnedDomainEventHandler} from '../../contexts/loyalty/application/event-handlers.js';
import schema from '../../contexts/loyalty/adaptors/messaging/generated/internal_commands.json' with {type: 'json'};
import definitions from '../../contexts/loyalty/adaptors/messaging/subscriptions.json' with {type: 'json'};
import {commandSubscription, completeSubscriptions} from '../runtime.js';

/** Stable queue identities survive class renames, releases and replay. */
export enum LoyaltySubscription {
  CreditCollection = 'loyalty.credit-collection',
  IssueReward = 'loyalty.issue-reward',
}
function fields(value: unknown): Record<string, unknown> {
  if (!value || typeof value !== 'object' || Array.isArray(value)) throw new Error('Command object required');
  return value as Record<string, unknown>;
}
const creditCodec = () => new InternalCommandCodec('loyalty', LoyaltySubscription.CreditCollection, 'creditCollection', schema,
  (value): CreditCollectionCommand => {const p = fields(value); return {orderId: identifier(String(p.orderId)), customerId: identifier(String(p.customerId))};});
const issueCodec = () => new InternalCommandCodec('loyalty', LoyaltySubscription.IssueReward, 'issueReward', schema,
  (value): IssueRewardCommand => {
    const p = fields(value);
    if (typeof p.benefit !== 'string' || !Number.isInteger(p.validDays) || Number(p.validDays) < 1) throw new Error('Invalid grant');
    return {grantId: identifier(String(p.grantId)), accountId: identifier(String(p.accountId)), benefit: p.benefit, validDays: Number(p.validDays)};
  });

/** Typed provider bindings for the Loyalty context. */
export type LoyaltyDependencies = {
    database: PostgresContextDatabase;
    accounts: AggregateCommandPort<ReturnType<typeof restoreAccount>>;
    rewards: AggregateCommandPort<ReturnType<typeof restoreReward>>;
    accountQueries: PagedQueryPort<ReturnType<typeof restoreAccount>>;
    rewardQueries: PagedQueryPort<ReturnType<typeof restoreReward>>;
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
    accountQueries: asFunction((c: LoyaltyDependencies) => new PostgresAggregateQueries(c.database, 'account', restoreAccount)).singleton(),
    rewardQueries: asFunction((c: LoyaltyDependencies) => new PostgresAggregateQueries(c.database, 'reward', restoreReward)).singleton(),
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
