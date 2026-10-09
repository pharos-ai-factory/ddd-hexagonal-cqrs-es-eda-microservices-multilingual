import {test} from 'node:test';
import assert from 'node:assert/strict';
import {MappedReadRepository} from '../../../adaptors/mapped-read-repository.js';
import {restoreAccountSnapshot} from './persistence/account-snapshot.js';
import {accountView} from './persistence/account-read-repository.js';
import {GetAccountQueryHandler} from '../application/queries/get-account.js';
import {ListAccountsQueryHandler} from '../application/queries/list-accounts.js';
import {accountQueryEndpoints} from './account-query-endpoints.js';
import type {PageRequest} from '../../../foundation/pagination.js';

test('query entry points preserve missing rows, revisions, page requests and storage failures', async () => {
  const id = '00000000-0000-4000-8000-000000000001';
  const row = {exists: true, version: 7, state: restoreAccountSnapshot({id, stampBalance: 1, collections: 1, grantsEarned: 0})};
  const page = {limit: 2, after: id}, failure = new Error('storage unavailable');
  const source = {get: async (key: string) => {if (key === 'failed') throw failure; return key === id ? row : undefined;},
    list: async () => [row], page: async (request: PageRequest) => {assert.deepEqual(request, page); return {items: [row], nextId: id};}};
  const reader = new MappedReadRepository(source, accountView);
  const endpoints = accountQueryEndpoints(new GetAccountQueryHandler(reader), new ListAccountsQueryHandler(reader));
  assert.equal(await endpoints.get('missing'), undefined);
  assert.deepEqual(await endpoints.get(id), row);
  assert.deepEqual(await endpoints.list(), [row]);
  assert.deepEqual(await endpoints.page(page), {items: [row], nextId: id});
  await assert.rejects(endpoints.get('failed'), failure);
});
