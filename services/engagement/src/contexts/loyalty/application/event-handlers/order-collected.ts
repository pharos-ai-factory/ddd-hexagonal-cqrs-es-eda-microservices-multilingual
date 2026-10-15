import type { CommandBusPort, Metadata } from '../../../../foundation/application.js';
import type { OrderCollectedIntegrationEvent } from '../../../../contracts/events/generated/collection.js';
import type { CreditCollectionCommand } from '../../../../contracts/requests/generated/loyalty.js';

/** Maps Collection's fact into a durable intent owned by Loyalty. */
export class OrderCollectedIntegrationEventHandler {
  constructor(private commands: CommandBusPort<CreditCollectionCommand>) {}
  handle(m: Metadata, event: OrderCollectedIntegrationEvent) {
    return this.commands.enqueue(m, { orderId: event.orderId, customerId: event.customerId });
  }
}
