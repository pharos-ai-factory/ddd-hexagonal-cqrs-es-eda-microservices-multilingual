export type Offer = {code: string; drinkId: string; drinkRevision: number; name: string; minor: number; currency: string};
export type Models = {
  drink: {id: string; name: string; revision: number; published: boolean};
  edition: {id: string; currency: string; status: string; offers: Offer[]};
  order: {id: string; customerId: string; editionId: string; currency: string; status: string;
    lines: {id: string; quantity: number; selection: {offerCode: string; name: string; minor: number}}[]};
  ticket: {id: string; orderId: string; customerId: string; instructions: string; status: string};
  pickup: {id: string; orderId: string; customerId: string; code: string; status: string};
  account: {id: string; stampBalance: number; collections: number; grantsEarned: number;
    lastGrant?: {id: string; accountId: string; benefit: string; validDays: number}};
  reward: {id: string; grantId: string; customerId: string; benefit: string; status: string; expiresAt: string; redeemedFor?: string};
  notification: {id: string; recipient: string; subject: string; body: string; status: string; providerReceipt?: string};
};
export type Kind = keyof Models;
export type Projection<K extends Kind = Kind> = {kind: K; version: number; state: Models[K]};
export type Snapshot = Record<string, Projection>;
export const owners: Record<Kind, string> = {
  drink: 'menu', edition: 'menu', order: 'ordering', ticket: 'preparation', pickup: 'collection',
  account: 'loyalty', reward: 'loyalty', notification: 'communication',
};

export function merge(current: Snapshot, update: Projection): Snapshot {
  const key = `${update.kind}/${update.state.id}`;
  if ((current[key]?.version ?? 0) >= update.version) return current;
  return {...current, [key]: update};
}
export function items<K extends Kind>(snapshot: Snapshot, kind: K): Projection<K>[] {
  return Object.values(snapshot).filter(value => value.kind === kind) as Projection<K>[];
}
