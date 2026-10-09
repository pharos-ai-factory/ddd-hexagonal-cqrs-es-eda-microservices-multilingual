import {request, success} from '../../adaptors/http/client';
import type {ListOperation, Success} from '../../adaptors/generated/http';
import type {Projection} from './model';

type Items<K extends ListOperation> = Extract<Success<K>, {items: unknown}>['items'];
export async function scan<K extends ListOperation>(operation: K, signal: AbortSignal): Promise<Items<K>> {
  const items: Array<Items<K>[number]> = [], seen = new Set<string>();
  let cursor: string | null = null;
  do {
    const response = await request<ListOperation>(operation, {query: {limit: 100,
      ...(cursor === null ? {} : {cursor})}}, {signal});
    const result = await success(operation, response);
    if (Array.isArray(result)) throw new Error('A paginated HTTP query returned an ordinary list');
    const page = result as Extract<Success<K>, {items: unknown}>;
    items.push(...page.items);
    cursor = page.nextCursor;
    if (cursor !== null) {
      if (!cursor || seen.has(cursor)) throw new Error('A paginated query did not advance');
      seen.add(cursor);
    }
  } while (cursor !== null);
  return items as Items<K>;
}
export async function projections(signal: AbortSignal): Promise<Projection[]> {
  const [drinks, editions, orders, tickets, pickups, accounts, rewards, notifications] = await Promise.all([
    scan('listDrinks', signal), scan('listEditions', signal), scan('listOrders', signal), scan('listTickets', signal),
    scan('listPickups', signal), scan('listAccounts', signal), scan('listRewards', signal), scan('listNotifications', signal),
  ]);
  return [
    ...drinks.map(row => ({kind: 'drink' as const, version: row.version, state: row.state})),
    ...editions.map(row => ({kind: 'edition' as const, version: row.version, state: {...row.state, offers: row.state.offers ?? []}})),
    ...orders.map(row => ({kind: 'order' as const, version: row.version, state: row.state})),
    ...tickets.map(row => ({kind: 'ticket' as const, version: row.version, state: row.state})),
    ...pickups.map(row => ({kind: 'pickup' as const, version: row.version, state: row.state})),
    ...accounts.map(row => ({kind: 'account' as const, version: row.version, state: row.state})),
    ...rewards.map(row => ({kind: 'reward' as const, version: row.version, state: row.state})),
    ...notifications.map(row => ({kind: 'notification' as const, version: row.version, state: row.state})),
  ];
}
