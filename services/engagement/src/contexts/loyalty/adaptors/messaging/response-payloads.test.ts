import { test } from 'node:test';
import assert from 'node:assert/strict';
import { encodeListAccountsResponse, toReward, registerGetReward } from './generated/response-payloads.js';
import { toOutcome } from '../../../../adaptors/messaging/generated/response-payloads.js';
import type {
  Accounts,
  ListAccountsResponse,
} from '../../../../contracts/requests/generated/loyalty-responses.js';
import type { RabbitMQRequestRegistry } from '../../../../adaptors/messaging/rabbitmq-request-registry.js';

await test('generated responses preserve nested values, optional empty values and numeric bounds', () => {
  const accounts: Accounts = {
    items: [
      {
        exists: true,
        version: 7,
        state: {
          id: 'account',
          stampBalance: 0,
          collections: 3,
          grantsEarned: 1,
          lastGrant: { id: 'grant', accountId: 'account', benefit: 'Coffee', validDays: 7 },
        },
      },
    ],
    paged: true,
    nextId: '',
  };
  const reply = encodeListAccountsResponse({}, { accounts });
  assert.deepEqual(reply, { accounts });
  delete accounts.nextId;
  assert.equal('nextId' in encodeListAccountsResponse({}, { accounts }).accounts!, false);
  assert.throws(() => toOutcome({ aggregateId: 'order', version: 2 ** 53, status: 'placed' }));
  assert.throws(() =>
    toReward({
      id: 'reward',
      grantId: 'grant',
      customerId: 'customer',
      benefit: 'Coffee',
      status: 'issued',
      expiresAt: 12 as unknown as string,
    }),
  );
  assert.throws(() =>
    encodeListAccountsResponse({}, {
      accounts,
      error: { code: 'failed', message: '' },
    } as unknown as ListAccountsResponse),
  );
});

// Compiled negative callers guard the generated request/response association.
function invalidResponse(registry: RabbitMQRequestRegistry) {
  registerGetReward(registry, () => ({ id: 'reward' }), {
    // @ts-expect-error GetReward cannot be registered with an account response handler.
    execute: async (): Promise<ListAccountsResponse> => ({ accounts: { items: [], paged: false } }),
  });
}
void invalidResponse;
