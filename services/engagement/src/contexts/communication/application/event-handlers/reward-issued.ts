import type {DurableCommandPort, Metadata} from '../../../../foundation/application.js';
import type {RewardIssued} from '../../../../contracts/events.js';
import type {RequestNotificationCommand} from '../commands/request-notification.js';

/** Translates a published reward into Communication's durable notification intent. */
export class RewardIssuedIntegrationEventHandler {
  constructor(private commands: DurableCommandPort<RequestNotificationCommand>) {}
  handle(m: Metadata, event: RewardIssued) {
    return this.commands.enqueue(m, {recipient: event.customerId, subject: 'You earned a reward',
      body: event.benefit + '; valid until ' + event.expiresAt});
  }
}
