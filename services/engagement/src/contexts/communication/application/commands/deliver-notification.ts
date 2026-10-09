import type {CommandContext} from '../../../../foundation/write-repository.js';
import type {NotificationWriteRepository} from '../ports/notification-write-repository.js';
import {Rejection} from '../../../../foundation/domain.js';

export type DeliverNotificationCommand = {notificationId: string};

/** Records provider acceptance through the notification aggregate; the executor obtains the receipt before the transaction. */
export class DeliverNotificationCommandHandler {
  constructor(private repository: NotificationWriteRepository, private receipt: string) {}
  async execute(context: CommandContext, command: DeliverNotificationCommand) {
    const current = await this.repository.get(context.target);
    if (!current) throw new Rejection('not_found', 'The notification does not exist');
    const notice = current.state;
    const changed = notice.recordDelivery(this.receipt);
    if (changed) await this.repository.save(notice);
    return {status: 'sent'};
  }
}
