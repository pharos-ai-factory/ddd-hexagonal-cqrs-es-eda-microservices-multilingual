/** Application-owned read representation of a reward. */
export type RewardView = Readonly<{
  id: string; grantId: string; customerId: string; benefit: string;
  status: 'issued' | 'redeemed' | 'expired'; expiresAt: string; redeemedFor?: string;
}>;
