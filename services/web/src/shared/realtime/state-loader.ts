import type { QueryClient, QueryKey } from '@tanstack/react-query';
import { getListPickupsQueryKey, listPickups } from '../../adaptors/generated/http/collection/collection';
import {
  getListNotificationsQueryKey,
  listNotifications,
} from '../../adaptors/generated/http/communication/communication';
import {
  getListAccountsQueryKey,
  getListRewardsQueryKey,
  listAccounts,
  listRewards,
} from '../../adaptors/generated/http/loyalty/loyalty';
import {
  getListDrinksQueryKey,
  getListEditionsQueryKey,
  listDrinks,
  listEditions,
} from '../../adaptors/generated/http/menu/menu';
import { getListOrdersQueryKey, listOrders } from '../../adaptors/generated/http/ordering/ordering';
import { getListTicketsQueryKey, listTickets } from '../../adaptors/generated/http/preparation/preparation';
import type {
  ListAccounts200,
  ListDrinks200,
  ListEditions200,
  ListNotifications200,
  ListOrders200,
  ListPickups200,
  ListRewards200,
  ListTickets200,
} from '../../adaptors/generated/http/model';
import { httpStatus, SessionEnded, type ListOperation } from '../../adaptors/http/operation-client';
import type { VersionedState } from './model';

type Paginated<T> = Extract<T, { items: unknown }>;
type ListRows = {
  listDrinks: Paginated<ListDrinks200>['items'];
  listEditions: Paginated<ListEditions200>['items'];
  listOrders: Paginated<ListOrders200>['items'];
  listTickets: Paginated<ListTickets200>['items'];
  listPickups: Paginated<ListPickups200>['items'];
  listAccounts: Paginated<ListAccounts200>['items'];
  listRewards: Paginated<ListRewards200>['items'];
  listNotifications: Paginated<ListNotifications200>['items'];
};
type Pagination = { limit?: number; cursor?: string };
type ListResult<Rows extends unknown[]> = {
  data: Rows | { items: Rows; nextCursor: string | null };
};
type ListDefinition<Rows extends unknown[]> = {
  queryKey: (params?: Pagination) => QueryKey;
  request: (params?: Pagination, options?: RequestInit) => Promise<ListResult<Rows>>;
};

const listQueries = {
  listDrinks: { queryKey: getListDrinksQueryKey, request: listDrinks },
  listEditions: { queryKey: getListEditionsQueryKey, request: listEditions },
  listOrders: { queryKey: getListOrdersQueryKey, request: listOrders },
  listTickets: { queryKey: getListTicketsQueryKey, request: listTickets },
  listPickups: { queryKey: getListPickupsQueryKey, request: listPickups },
  listAccounts: { queryKey: getListAccountsQueryKey, request: listAccounts },
  listRewards: { queryKey: getListRewardsQueryKey, request: listRewards },
  listNotifications: { queryKey: getListNotificationsQueryKey, request: listNotifications },
} satisfies { [Name in ListOperation]: ListDefinition<ListRows[Name]> };

/** Load every page for one generated owner query through the TanStack cache. */
export async function scan<Name extends ListOperation>(
  operation: Name,
  queryClient: QueryClient,
  signal: AbortSignal,
): Promise<ListRows[Name]> {
  // TypeScript loses the key/value correlation when indexing this heterogeneous map.
  const definition = listQueries[operation] as unknown as ListDefinition<ListRows[Name]>;
  const items: Array<ListRows[Name][number]> = [];
  const seen = new Set<string>();
  let cursor: string | null = null;
  try {
    do {
      const params: Pagination = { limit: 100, ...(cursor === null ? {} : { cursor }) };
      const response: ListResult<ListRows[Name]> = await queryClient.fetchQuery({
        queryKey: definition.queryKey(params),
        queryFn: () => definition.request(params, { signal }),
        staleTime: 0,
      });
      if (Array.isArray(response.data)) throw new Error('A paginated HTTP query returned an ordinary list');
      const page = response.data;
      items.push(...page.items);
      cursor = page.nextCursor;
      if (cursor !== null) {
        if (!cursor || seen.has(cursor)) throw new Error('A paginated query did not advance');
        seen.add(cursor);
      }
    } while (cursor !== null);
  } catch (error) {
    if (httpStatus(error) === 401) throw new SessionEnded();
    throw error;
  }
  return items as ListRows[Name];
}

export async function loadCafeState(
  queryClient: QueryClient,
  signal: AbortSignal,
): Promise<VersionedState[]> {
  const [drinks, editions, orders, tickets, pickups, accounts, rewards, notifications] = await Promise.all([
    scan('listDrinks', queryClient, signal),
    scan('listEditions', queryClient, signal),
    scan('listOrders', queryClient, signal),
    scan('listTickets', queryClient, signal),
    scan('listPickups', queryClient, signal),
    scan('listAccounts', queryClient, signal),
    scan('listRewards', queryClient, signal),
    scan('listNotifications', queryClient, signal),
  ]);
  return [
    ...drinks.map((row) => ({ kind: 'drink' as const, version: row.version, state: row.state })),
    ...editions.map((row) => ({
      kind: 'edition' as const,
      version: row.version,
      state: { ...row.state, offers: row.state.offers ?? [] },
    })),
    ...orders.map((row) => ({ kind: 'order' as const, version: row.version, state: row.state })),
    ...tickets.map((row) => ({ kind: 'ticket' as const, version: row.version, state: row.state })),
    ...pickups.map((row) => ({ kind: 'pickup' as const, version: row.version, state: row.state })),
    ...accounts.map((row) => ({ kind: 'account' as const, version: row.version, state: row.state })),
    ...rewards.map((row) => ({ kind: 'reward' as const, version: row.version, state: row.state })),
    ...notifications.map((row) => ({
      kind: 'notification' as const,
      version: row.version,
      state: row.state,
    })),
  ];
}
