/** Application-owned read representation of an account. */
export type AccountView = Readonly<{
  id: string; stampBalance: number; collections: number; grantsEarned: number;
  lastGrant?: Readonly<{id: string; accountId: string; benefit: string; validDays: number}>;
}>;
