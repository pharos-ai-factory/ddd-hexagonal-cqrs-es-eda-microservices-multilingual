import {CorruptState} from '../foundation/domain.js';
import {LoyaltyAccount, type AccountState} from '../contexts/loyalty/domain/account.js';
import {Reward, type RewardState} from '../contexts/loyalty/domain/reward.js';
import {Notification, type NotificationState} from '../contexts/communication/domain/notification.js';

function record(value: unknown): object {
  if (!value || typeof value !== 'object' || Array.isArray(value)) {
    throw new CorruptState('Stored aggregate snapshot is not an object');
  }
  return value;
}

export const restoreAccount = (value: unknown): AccountState =>
  new LoyaltyAccount(record(value) as AccountState).snapshot();
export const restoreReward = (value: unknown): RewardState =>
  new Reward(record(value) as RewardState).snapshot();
export const restoreNotification = (value: unknown): NotificationState =>
  new Notification(record(value) as NotificationState).snapshot();
