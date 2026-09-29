import {ApplicationError, type CommandPort, type IdentityFactory, type Metadata, type Publication} from '../../../foundation/application.js';
import type {OrderCollected, RewardEarned} from '../../../contracts/events.js';
import {LoyaltyAccount, type AccountState} from '../domain/account.js';
import {Reward, type RewardState} from '../domain/reward.js';

export class CreditCollection {
  constructor(private accounts: CommandPort<AccountState>, private ids: IdentityFactory) {}
  handle(metadata: Metadata, event: OrderCollected) {
    const m = {...metadata, id: this.ids('collection-credit', event.orderId)};
    return this.accounts.execute(m, loaded => {
      const account = loaded ? new LoyaltyAccount(loaded.state) : LoyaltyAccount.open(event.customerId);
      account.credit(event.orderId, this.ids('earned-grant', event.orderId));
      const publications: Publication[] = account.events().flatMap(fact => fact.type === 'RewardEarned'
        ? [{name: 'loyalty.reward-earned', payload: {accountId: fact.grant.accountId, grantId: fact.grant.id,
            benefit: fact.grant.benefit, validDays: fact.grant.validDays}}] : []);
      return {state: account.snapshot(), status: 'active', changed: true, publications};
    });
  }
}
export class IssueReward {
  constructor(private rewards: CommandPort<RewardState>, private ids: IdentityFactory, private clock: () => Date) {}
  handle(metadata: Metadata, event: RewardEarned) {
    return this.rewards.execute({...metadata, id: this.ids('issue-reward', event.grantId)}, loaded => {
      if (loaded) {
        const state = new Reward(loaded.state).snapshot();
        return {state, status: state.status, changed: false};
      }
      const reward = Reward.issue(metadata.target, {id: event.grantId, accountId: event.accountId,
        benefit: event.benefit, validDays: event.validDays}, this.clock());
      const state = reward.snapshot();
      return {state, status: 'issued', changed: true, publications: [{name: 'loyalty.reward-issued', payload: {
        rewardId: state.id, customerId: state.customerId, benefit: state.benefit, expiresAt: state.expiresAt}}]};
    });
  }
}
export class RedeemReward {
  constructor(private rewards: CommandPort<RewardState>, private clock: () => Date) {}
  execute(m: Metadata, command: {orderId: string}) {
    return this.rewards.execute(m, loaded => {
      if (!loaded) throw new RewardNotFoundApplicationError();
      const reward = new Reward(loaded.state);
      reward.redeem(command.orderId, this.clock());
      return {state: reward.snapshot(), status: 'redeemed', changed: true};
    });
  }
}

export class RewardNotFoundApplicationError extends ApplicationError {
  constructor() { super('not_found', 'The reward does not exist'); }
}
