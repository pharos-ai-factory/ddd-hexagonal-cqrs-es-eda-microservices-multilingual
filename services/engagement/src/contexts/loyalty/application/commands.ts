import {ApplicationError, type AggregateCommandPort, type IdentityFactory, type Metadata, type Publication} from '../../../foundation/application.js';
import {LoyaltyAccount, type AccountState} from '../domain/account.js';
import {Reward, type RewardState} from '../domain/reward.js';

/** Credits one account and records any earned grant atomically. */
export class CreditCollectionCommandHandler {
  constructor(private accounts: AggregateCommandPort<AccountState>, private ids: IdentityFactory) {}
  execute(metadata: Metadata, command: CreditCollectionCommand) {
    const m = {...metadata, id: this.ids('collection-credit', command.orderId)};
    return this.accounts.execute(m, loaded => {
      const account = loaded ? new LoyaltyAccount(loaded.state) : LoyaltyAccount.open(command.customerId);
      account.credit(command.orderId, this.ids('earned-grant', command.orderId));
      const publications: Publication[] = account.events().flatMap(fact => fact.type === 'RewardEarned'
        ? [{name: 'loyalty.reward-earned', payload: {accountId: fact.grant.accountId, grantId: fact.grant.id,
            benefit: fact.grant.benefit, validDays: fact.grant.validDays}}] : []);
      return {state: account.snapshot(), status: 'active', changed: true, publications};
    });
  }
}
/** Creates one Reward from its immutable earned grant and injected clock. */
export class IssueRewardCommandHandler {
  constructor(private rewards: AggregateCommandPort<RewardState>, private ids: IdentityFactory, private clock: () => Date) {}
  execute(metadata: Metadata, command: IssueRewardCommand) {
    return this.rewards.execute({...metadata, id: this.ids('issue-reward', command.grantId)}, loaded => {
      if (loaded) {
        const state = new Reward(loaded.state).snapshot();
        return {state, status: state.status, changed: false};
      }
      const reward = Reward.issue(metadata.target, {id: command.grantId, accountId: command.accountId,
        benefit: command.benefit, validDays: command.validDays}, this.clock());
      const state = reward.snapshot();
      return {state, status: 'issued', changed: true, publications: [{name: 'loyalty.reward-issued', payload: {
        rewardId: state.id, customerId: state.customerId, benefit: state.benefit, expiresAt: state.expiresAt}}]};
    });
  }
}
/** Applies the reward redemption rule using the injected clock and one aggregate transaction. */
export class RedeemRewardCommandHandler {
  constructor(private rewards: AggregateCommandPort<RewardState>, private clock: () => Date) {}
  execute(m: Metadata, command: RedeemRewardCommand) {
    return this.rewards.execute(m, loaded => {
      if (!loaded) throw new RewardNotFoundApplicationError();
      const reward = new Reward(loaded.state);
      reward.redeem(command.orderId, this.clock());
      return {state: reward.snapshot(), status: 'redeemed', changed: true};
    });
  }
}

/** Reports the expected rejection when the targeted reward has no stored state. */
export class RewardNotFoundApplicationError extends ApplicationError {
  constructor() { super('not_found', 'The reward does not exist'); }
}

/** Owner-local intent, independent of the Collection event envelope. */
export type CreditCollectionCommand = {orderId: string; customerId: string};
/** Carries the immutable grant so reward issuance can recover independently. */
export type IssueRewardCommand = {grantId: string; accountId: string; benefit: string; validDays: number};
export type RedeemRewardCommand = {orderId: string};
