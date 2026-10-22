import type { CreditCollectionCommand } from '../../../../contracts/requests/generated/loyalty.js';
import {
  realtimeMessage,
  realtimeTarget,
  type RealtimeBusPort,
  type RealtimeMessage,
} from '../../../../foundation/application.js';
import type { CommandContext } from '../../../../foundation/write-repository.js';
import type { AccountWriteRepository } from '../ports/account-write-repository.js';
import type { IdentityFactory } from '../../../../foundation/application.js';
import { LoyaltyAccount } from '../../domain/loyalty-account.js';

/** Owner-local intent, independent of the Collection event envelope. */

/** Credits one account and records any earned grant atomically. */
export class CreditCollectionCommandHandler {
  constructor(
    private repository: AccountWriteRepository,
    private realtime: RealtimeBusPort<RealtimeMessage>,
    private ids: IdentityFactory,
  ) {}
  async execute(context: CommandContext, command: CreditCollectionCommand) {
    const loaded = await this.repository.get(context.target);
    const account = loaded ? loaded.state : LoyaltyAccount.open(command.customerId);
    account.credit(command.orderId, this.ids('earned-grant', command.orderId));
    await this.repository.save(account);
    await this.realtime.publish(
      realtimeMessage.snapshot(realtimeTarget.context('loyalty'), account.snapshot()),
    );
    return { status: 'active' };
  }
}
