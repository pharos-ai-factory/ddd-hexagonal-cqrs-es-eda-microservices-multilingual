import type { DeliverNotificationCommand } from '../../../../contracts/requests/generated/communication.js';
import type { RealtimeBusPort, RealtimeEvent } from '../../../../foundation/application.js';
import type { CommandContext } from '../../../../foundation/write-repository.js';
import type { NotificationWriteRepository } from '../ports/notification-write-repository.js';
import { Rejection } from '../../../../foundation/domain.js';

/** Records provider acceptance through the notification aggregate; delivery preparation obtains the receipt before the transaction. */
export class DeliverNotificationCommandHandler {
  constructor(
    private repository: NotificationWriteRepository,
    private realtime: RealtimeBusPort<RealtimeEvent>,
    private receipt: string,
  ) {}
  async execute(context: CommandContext, _command: DeliverNotificationCommand) {
    const current = await this.repository.get(context.target);
    if (!current) throw new Rejection('not_found', 'The notification does not exist');
    const notice = current.state;
    const changed = notice.recordDelivery(this.receipt);
    if (changed) {
      await this.repository.save(notice);
      await this.realtime.publish({ snapshot: notice.snapshot() });
    }
    return { status: 'sent' };
  }
}
