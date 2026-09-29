import {identifier, Rejection} from '../foundation/domain.js';
import type {Page, PageRequest} from '../foundation/pagination.js';
export function pageRequest(query: URLSearchParams, resource: string): PageRequest | undefined {
  if (!query.has('limit') && !query.has('cursor')) return;
  const invalid = () => new Rejection('invalid_pagination', 'Supply limit 1–100 and a cursor from this resource');
  const raw = query.get('limit') ?? '';
  if (query.getAll('limit').length !== 1 || query.getAll('cursor').length > 1 || !/^[1-9][0-9]*$/.test(raw)) throw invalid();
  const limit = Number(raw);
  if (limit > 100) throw invalid();
  let after: string | undefined;
  if (query.has('cursor')) {
    const cursor = query.get('cursor')!;
    if (cursor.length > 1024) throw invalid();
    const bytes = Buffer.from(cursor, 'base64url');
    const prefix = '1|'+resource+'|';
    if (bytes.toString('base64url') !== cursor || !bytes.toString().startsWith(prefix)) throw invalid();
    try { after = identifier(bytes.toString().slice(prefix.length)); } catch { throw invalid(); }
  }
  return {limit, after};
}
export function pageResponse<S>(page: Page<S>, resource: string) {
  return {items: page.items, nextCursor: page.nextId ? Buffer.from('1|'+resource+'|'+page.nextId).toString('base64url') : null};
}
