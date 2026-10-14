/** Stable owner-local subscriptions; each installs a separate command queue. */
export enum CommunicationSubscription {
  PickupNotice = 'communication.pickup-notice',
  RewardNotice = 'communication.reward-notice',
  DeliverNotice = 'communication.deliver-notice',
}
