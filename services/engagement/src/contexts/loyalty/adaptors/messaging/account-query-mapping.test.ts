import { test } from 'node:test';
import assert from 'node:assert/strict';
import { MappedReadRepository } from '../../../../adaptors/persistence/mapped-read-repository.js';
import { restoreAccountSnapshot } from '../persistence/account-snapshot.js';
import { accountView } from '../persistence/account-read-repository.js';
import { GetAccountQueryHandler } from '../../application/queries/get-account.js';
import { ListAccountsQueryHandler } from '../../application/queries/list-accounts.js';
import * as queries from './account-query-mapping.js';
import type { PageRequest } from '../../../../foundation/pagination.js';

void test('query entry points preserve missing rows, revisions, page requests and storage failures', async () => {
  const id = '00000000-0000-4000-8000-000000000001';
  const row = {
    exists: true,
    version: 7,
    state: restoreAccountSnapshot({ id, stampBalance: 1, collections: 1, grantsEarned: 0 }),
  };
  const page = { limit: 2, after: id },
    failure = new Error('storage unavailable');
  const source = {
    get: async (key: string) => {
      if (key === 'failed') throw failure;
      return key === id ? row : undefined;
    },
    list: async () => [row],
    page: async (request: PageRequest) => {
      assert.deepEqual(request, page);
      return { items: [row], nextId: id };
    },
  };
  const reader = new MappedReadRepository(source, accountView);
  const get = new GetAccountQueryHandler(reader);
  const list = new ListAccountsQueryHandler(reader);
  assert.deepEqual(await get.execute({ id: 'missing' }), { error: { code: 'not_found', message: '' } });
  assert.deepEqual(await get.execute(queries.getAccount({ id })), { account: row });
  assert.deepEqual(await list.execute(queries.listAccounts({})), {
    accounts: { items: [row], paged: false },
  });
  assert.deepEqual(await list.execute(queries.listAccounts({ page })), {
    accounts: { items: [row], nextId: id, paged: true },
  });
  await assert.rejects(get.execute({ id: 'failed' }), failure);
});
