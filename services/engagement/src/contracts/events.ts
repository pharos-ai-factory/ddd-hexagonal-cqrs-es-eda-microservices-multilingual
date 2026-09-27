export type OrderCollected = {orderId: string; customerId: string};
export type RewardEarned = {accountId: string; grantId: string; benefit: string; validDays: number};
export type RewardIssued = {rewardId: string; customerId: string; benefit: string; expiresAt: string};
export type PickupOpened = {pickupId: string; orderId: string; customerId: string; collectionCode: string};
export type NotificationRequested = {notificationId: string};
