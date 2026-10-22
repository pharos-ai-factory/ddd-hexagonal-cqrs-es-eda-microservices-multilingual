import { test } from 'node:test';
import assert from 'node:assert/strict';
import protobuf from 'protobufjs/light';
import schema from '../generated/realtime.json';
import { decodeRealtimeMessage } from './message-codec';

const wire = protobuf.Root.fromJSON(schema).lookupType('cafe.realtime.v1.Publication');
const id = '00000000-0000-4000-8000-000000000001';

void test('operation rejections are delivered only on their principal target', () => {
  const body = wire
    .encode(
      wire.create({
        eventId: id,
        contractVersion: 1,
        context: 'ordering',
        aggregateKind: 'order',
        aggregateId: id,
        revision: 0,
        target: { principal: id },
        operationRejected: {
          commandId: id,
          correlationId: id,
          code: 'not_found',
          message: 'The order does not exist',
        },
      }),
    )
    .finish();
  const result = decodeRealtimeMessage('principal:' + id, body);
  assert.equal(result.type, 'operationRejected');
  if (result.type === 'operationRejected') assert.equal(result.commandId, id);
  assert.throws(() => decodeRealtimeMessage('cafe:ordering', body), /rejection/);
});

void test('state messages require a positive revision and matching context target', () => {
  const body = wire
    .encode(
      wire.create({
        eventId: id,
        contractVersion: 1,
        context: 'ordering',
        aggregateKind: 'order',
        aggregateId: id,
        revision: 0,
        target: { context: 'ordering' },
        order: { id },
      }),
    )
    .finish();
  assert.throws(() => decodeRealtimeMessage('cafe:ordering', body), /state message/);
});
