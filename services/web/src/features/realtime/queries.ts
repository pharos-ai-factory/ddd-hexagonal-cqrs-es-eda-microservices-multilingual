// Query types opt into pagination explicitly; ordinary queries keep their own shape.
export async function query<T>(url: string, signal?: AbortSignal): Promise<T> {
  const response = await fetch(url, {cache: 'no-store', signal});
  if (response.status === 401) throw new SessionEnded();
  if (!response.ok) throw new Error('The current café state could not be loaded. Reconnect to retry.');
  return await response.json() as T;
}
export class SessionEnded extends Error {
  constructor() { super('Session ended'); }
}
export type Page<T> = {items: T[]; nextCursor: string | null};
export async function queryAll<T>(resource: string, read: (url: string) => Promise<Page<T>> = query): Promise<T[]> {
  const items: T[] = [], seen = new Set<string>();
  let cursor: string | null = null;
  do {
    const params = new URLSearchParams({limit: '100'});
    if (cursor) params.set('cursor', cursor);
    const page = await read(resource+(resource.includes('?') ? '&' : '?')+params);
    if (!Array.isArray(page.items) || !(page.nextCursor === null || typeof page.nextCursor === 'string')) {
      throw new Error('Invalid paginated query response');
    }
    items.push(...page.items);
    cursor = page.nextCursor;
    if (cursor !== null) {
      if (!cursor || seen.has(cursor)) throw new Error('A paginated query did not advance');
      seen.add(cursor);
    }
  } while (cursor !== null);
  return items;
}

// A scan starts only after all authorised subscriptions are attached. Publications
// then cover inserts behind a keyset cursor while revision guards cover updates.
export function subscriptionBarrier(channels: readonly string[]) {
  const pending = new Set(channels);
  let needsScan = false;
  return {
    reset(force: boolean) { pending.clear(); channels.forEach(channel => pending.add(channel)); needsScan = force; },
    subscribed(channel: string, recovered: boolean): boolean {
      if (!pending.delete(channel)) return false;
      needsScan ||= !recovered;
      return pending.size === 0 && needsScan;
    },
  };
}
