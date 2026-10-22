import type { RewardReadRepository } from '../ports/reward-read-repository.js';
import type { RewardIssuedIntegrationEvent } from '../../../../contracts/events/generated/loyalty.js';
import type {
  Metadata,
  Outcome,
  RealtimeBusPort,
  RealtimeMessage,
} from '../../../../foundation/application.js';
import { realtimeTarget } from '../../../../foundation/application.js';

/** Explicitly translate a published reward fact into an owner-authorised browser update. */
export class RewardIssuedRealtimeIntegrationEventHandler {
  constructor(
    private rewards: RewardReadRepository,
    private realtime: RealtimeBusPort<RealtimeMessage>,
  ) {}
  async handle(metadata: Metadata, event: RewardIssuedIntegrationEvent): Promise<Outcome> {
    if (event.rewardId !== metadata.target) throw new Error('Reward event target mismatch');
    const loaded = await this.rewards.get(metadata.target);
    if (!loaded) throw new Error('Committed reward is missing');
    await this.realtime.publish({
      target: realtimeTarget.context('loyalty'),
      source: metadata,
      revision: loaded.version,
      snapshot: loaded.state,
    });
    return { aggregateId: metadata.target, version: loaded.version, status: 'realtime_queued' };
  }
}
