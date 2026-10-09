import {type AggregateCommandPort, type IdentityFactory, type Metadata, type Publication} from '../../../../foundation/application.js';
import {LoyaltyAccount, type AccountState} from '../../domain/loyalty-account.js';

/** Owner-local intent, independent of the Collection event envelope. */
export type CreditCollectionCommand = {orderId: string; customerId: string};

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
