export type OrderCollected = {orderId: string; customerId: string};
export type RewardIssued = {rewardId: string; customerId: string; benefit: string; expiresAt: string};
export type PickupOpened = {pickupId: string; orderId: string; customerId: string; collectionCode: string};
