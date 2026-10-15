import type { RealtimeBusPort, RealtimeEvent } from '../../../../foundation/application.js';
import type { CommandContext } from '../../../../foundation/write-repository.js';
import type { NotificationWriteRepository } from '../ports/notification-write-repository.js';
import { Notification } from '../../domain/notification.js';

/** Owner-local notification content, prepared before the durable hand-off. */
export type RequestNotificationCommand = { recipient: string; subject: string; body: string };

/** Creates one notification and records its private delivery-requested fact atomically. */
export class RequestNotificationCommandHandler {
  constructor(
    private repository: NotificationWriteRepository,
    private realtime: RealtimeBusPort<RealtimeEvent>,
  ) {}
  async execute(context: CommandContext, command: RequestNotificationCommand) {
    const loaded = await this.repository.get(context.target);
    if (loaded) {
      const state = loaded.state.snapshot();
      return { status: state.status };
    }
    const notification = Notification.request(
      context.target,
      command.recipient,
      command.subject,
      command.body,
    );
    await this.repository.save(notification);
    await this.realtime.publish({ snapshot: notification.snapshot() });
    return { status: 'requested' };
  }
}
