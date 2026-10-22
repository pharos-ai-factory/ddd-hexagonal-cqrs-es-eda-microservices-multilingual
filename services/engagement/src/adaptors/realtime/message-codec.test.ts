import { test } from 'node:test';
import assert from 'node:assert/strict';
import protobuf from 'protobufjs';
import schema from '../generated/realtime.json' with { type: 'json' };
import { realtimeTarget } from '../../foundation/application.js';
import { realtime } from './message-codec.js';

const publication = protobuf.Root.fromJSON(schema).lookupType('cafe.realtime.v1.Publication');
const id = '00000000-0000-4000-8000-000000000001';

void test('context and principal targets are encoded', () => {
  const state = realtime('loyalty', 'reward', id, 1, realtimeTarget.context('loyalty'), { id });
  const stateValue = publication.toObject(publication.decode(state.body), { oneofs: true }) as {
    target: { target: string; context: string };
  };
  assert.deepEqual(stateValue.target, { context: 'loyalty', target: 'context' });

  const rejection = realtime(
    'loyalty',
    'reward',
    id,
    0,
    realtimeTarget.principal(id),
    { commandId: id, correlationId: id, code: 'not_found', message: 'missing' },
    'operationRejected',
  );
  const rejectionValue = publication.toObject(publication.decode(rejection.body), {
    oneofs: true,
  }) as { target: { target: string; principal: string } };
  assert.deepEqual(rejectionValue.target, { principal: id, target: 'principal' });
});
