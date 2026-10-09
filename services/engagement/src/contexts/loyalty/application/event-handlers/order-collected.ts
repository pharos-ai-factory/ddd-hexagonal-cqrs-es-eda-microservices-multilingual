import type {DurableCommandPort, Metadata} from '../../../../foundation/application.js';
import type {OrderCollected} from '../../../../contracts/events.js';
import type {CreditCollectionCommand} from '../commands/credit-collection.js';

/** Maps Collection's fact into a durable intent owned by Loyalty. */
export class OrderCollectedIntegrationEventHandler {
  constructor(private commands: DurableCommandPort<CreditCollectionCommand>) {}
  handle(m: Metadata, event: OrderCollected) {
    return this.commands.enqueue(m, {orderId: event.orderId, customerId: event.customerId});
  }
}
