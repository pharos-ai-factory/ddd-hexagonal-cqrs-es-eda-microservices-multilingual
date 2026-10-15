import type { RewardIssuedDomainEventPayload } from '../events.js';
import type { RewardIssuedIntegrationEvent } from '../../../../contracts/events/generated/loyalty.js';
import type { IntegrationEventBusPort, Metadata } from '../../../../foundation/application.js';

/** Publish the committed reward terms through the atomic receipt/outbox port. */
export class RewardIssuedDomainEventHandler {
  constructor(private publications: IntegrationEventBusPort<RewardIssuedIntegrationEvent>) {}
  handle(metadata: Metadata, event: RewardIssuedDomainEventPayload) {
    return this.publications.publish(metadata, {
      rewardId: event.rewardId,
      customerId: event.customerId,
      benefit: event.benefit,
      expiresAt: event.expiresAt,
    });
  }
}
