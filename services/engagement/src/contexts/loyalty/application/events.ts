// Owner-private application delivery values.
export type RewardEarnedDomainEventPayload = {
  accountId: string;
  grantId: string;
  benefit: string;
  validDays: number;
};

/** Private delivery value for the committed RewardIssuedDomainEvent. */
export type RewardIssuedDomainEventPayload = Readonly<{
  rewardId: string;
  customerId: string;
  benefit: string;
  expiresAt: string;
}>;
