import {notificationPublications} from '../../src/contexts/communication/adaptors/messaging/notification-publications.js';
import {rewardPublications} from '../../src/contexts/loyalty/adaptors/messaging/reward-publications.js';
import {accountPublications} from '../../src/contexts/loyalty/adaptors/messaging/account-publications.js';
import {Notification} from '../../src/contexts/communication/domain/notification.js';
import {Reward} from '../../src/contexts/loyalty/domain/reward.js';
import {LoyaltyAccount} from '../../src/contexts/loyalty/domain/loyalty-account.js';
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
  accountsTransaction = this.accounts.transaction(state => new LoyaltyAccount(state), root => root.snapshot(), accountPublications);
  rewardsTransaction = this.rewards.transaction(state => new Reward(state), root => root.snapshot(), rewardPublications);
  noticesTransaction = this.notices.transaction(state => new Notification(state), root => root.snapshot(), notificationPublications);
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
