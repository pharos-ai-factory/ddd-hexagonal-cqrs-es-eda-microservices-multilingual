import {setWorldConstructor, World} from '@cucumber/cucumber';
import type {Publication} from '../../src/foundation/application.js';
import type {AccountState} from '../../src/contexts/loyalty/domain/loyalty-account.js';
import type {RewardState} from '../../src/contexts/loyalty/domain/reward.js';
import type {NotificationState} from '../../src/contexts/communication/domain/notification.js';
import {CommandProbe} from './probes.js';

export class EngagementWorld extends World {
  accounts = new CommandProbe<AccountState>();
  rewards = new CommandProbe<RewardState>();
  notices = new CommandProbe<NotificationState>();
  credits = 0;
  earned: Publication[] = [];
  accountBeforeReward: unknown;
  now = new Date('2026-01-01T12:00:00.000Z');
  providerCalls = 0;
  unavailable = false;
  deliveryError: unknown;
  beforeDelivery: unknown;
  commandsBeforeDelivery = 0;
  provider = {
    deliver: async () => {
      this.providerCalls++;
      if (this.unavailable) throw new Error('provider unavailable');
      return 'accepted-1';
    },
  };
}
setWorldConstructor(EngagementWorld);
