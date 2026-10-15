import { encodeGetAccountResponse } from '../contexts/loyalty/adaptors/messaging/generated/response-payloads.js';
import type { GetAccountResponse } from '../contracts/requests/generated/loyalty-responses.js';
import { test } from 'node:test';
import assert from 'node:assert/strict';
import { asValue } from 'awilix';
import protobuf from 'protobufjs';
import schema from '../adaptors/generated/requests.json' with { type: 'json' };
import { LoyaltyQuery, CommunicationQuery, LoyaltyCommand } from '../adaptors/generated/request-names.js';
import { decodeRequest } from '../adaptors/messaging/request-envelope.js';
import { RabbitMQRequestRegistry } from '../adaptors/messaging/rabbitmq-request-registry.js';
import { createLoyaltyContainer, loyaltyRequests } from './composition/loyalty.js';
import * as mapping from '../contexts/loyalty/adaptors/messaging/account-query-mapping.js';
import type { PageRequest } from '../foundation/pagination.js';

const id = '00000000-0000-4000-8000-000000000001';
const missing = '00000000-0000-4000-8000-000000000002';
const root = protobuf.Root.fromJSON(schema);
const requestType = root.lookupType('cafe.loyalty.requests.v1.Request');
const replyType = root.lookupType('cafe.loyalty.requests.v1.Reply');
const row = { exists: true, version: 7, state: { id, stampBalance: 1, collections: 1, grantsEarned: 0 } };

const wireRow = { ...row, state: { ...row.state, lastGrant: null } };

void test('direct query handlers preserve missing roots, complete reads, pagination and storage failure', async () => {
  const container = createLoyaltyContainer('postgresql://unused:unused@127.0.0.1:1/unused');
  const calls: unknown[] = [];
  let failed = false;
  let continuation: string | undefined = id;
  container.register({
    accountReadRepository: asValue({
      get: async (identity: string) => {
        calls.push(identity);
        if (failed) throw new Error('private database detail');
        return identity === id ? row : undefined;
      },
      list: async () => {
        calls.push('complete');
        return Array.from({ length: 101 }, () => row);
      },
      page: async (page: PageRequest) => {
        calls.push(page);
        return { items: [row], nextId: continuation };
      },
    }),
  });
  try {
    const registry = loyaltyRequests(container.cradle);
    const query = async (name: LoyaltyQuery, value: object) => {
      const request = decodeRequest(
        Buffer.from(
          requestType
            .encode(
              requestType.fromObject({
                contractVersion: 1,
                requestId: id,
                context: 'loyalty',
                query: { [name]: value },
              }),
            )
            .finish(),
        ),
        'loyalty',
      );
      return replyType.toObject(replyType.decode(await registry.handle(request)), {
        defaults: true,
        longs: Number,
      });
    };
    const item = await query(LoyaltyQuery.GET_ACCOUNT, { id });
    assert.deepEqual(item.account, wireRow);
    assert.deepEqual((await query(LoyaltyQuery.GET_ACCOUNT, { id: missing })).error, {
      code: 'not_found',
      message: '',
    });
    const list: unknown = (await query(LoyaltyQuery.LIST_ACCOUNTS, {})).accounts;
    assert.deepEqual(list, { items: Array.from({ length: 101 }, () => wireRow), paged: false });
    const paged: unknown = (await query(LoyaltyQuery.LIST_ACCOUNTS, { page: { limit: 2, after: id } }))
      .accounts;
    assert.deepEqual(paged, { items: [wireRow], paged: true, nextId: id });
    assert.deepEqual(calls.at(-1), { limit: 2, after: id });
    continuation = '';
    assert.deepEqual((await query(LoyaltyQuery.LIST_ACCOUNTS, { page: { limit: 2 } })).accounts, {
      items: [wireRow],
      paged: true,
      nextId: '',
    });
    continuation = undefined;
    assert.deepEqual((await query(LoyaltyQuery.LIST_ACCOUNTS, { page: { limit: 2 } })).accounts, {
      items: [wireRow],
      paged: true,
    });
    const before = calls.length;
    await assert.rejects(query(LoyaltyQuery.LIST_ACCOUNTS, { page: { limit: 0 } }));
    assert.equal(calls.length, before);
    failed = true;
    const error = await query(LoyaltyQuery.GET_ACCOUNT, { id });
    assert.deepEqual(error.error, { code: 'temporarily_unavailable', message: '' });
    assert.equal(error.requestId, id);
  } finally {
    await container.dispose();
  }
});

void test('query registrations reject duplicates and foreign identifiers', () => {
  const registry = new RabbitMQRequestRegistry('loyalty');
  const handler = {
    execute: async (): Promise<GetAccountResponse> => ({ error: { code: 'not_found', message: '' } }),
  };
  registry.query(LoyaltyQuery.GET_ACCOUNT, mapping.getAccount, handler, encodeGetAccountResponse);
  assert.throws(
    () => registry.query(LoyaltyQuery.GET_ACCOUNT, mapping.getAccount, handler, encodeGetAccountResponse),
    /Duplicate/,
  );
  assert.throws(
    () =>
      registry.query(
        CommunicationQuery.GET_NOTIFICATION,
        mapping.getAccount,
        handler,
        encodeGetAccountResponse,
      ),
    /owner/,
  );
});

// These callers are compiled but never invoked; widening the API makes the directives fail.
function rejectedIdentifiers(registry: RabbitMQRequestRegistry) {
  const handler = {
    execute: async (): Promise<GetAccountResponse> => ({ error: { code: 'not_found', message: '' } }),
  };
  // @ts-expect-error Raw wire strings require a generated query identifier.
  registry.query('getAccount', mapping.getAccount, handler, encodeGetAccountResponse);
  // @ts-expect-error A command identifier cannot select a query handler.
  registry.query(LoyaltyCommand.REDEEM_REWARD, mapping.getAccount, handler, encodeGetAccountResponse);
}
void rejectedIdentifiers;
