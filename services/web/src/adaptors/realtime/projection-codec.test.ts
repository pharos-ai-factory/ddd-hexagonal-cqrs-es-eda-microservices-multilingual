import { test } from 'node:test';
import assert from 'node:assert/strict';
import protobuf from 'protobufjs/light';
import schema from '../generated/realtime.json';
import { decode } from './projection-codec';

const wire = protobuf.Root.fromJSON(schema).lookupType('cafe.realtime.v1.Publication');
const id = '00000000-0000-4000-8000-000000000001';

void test('operation rejections have command identity and can precede aggregate creation', () => {
  const body = wire
    .encode(
      wire.create({
        eventId: id,
        contractVersion: 1,
        context: 'ordering',
        aggregateKind: 'order',
        aggregateId: id,
        revision: 0,
        operationRejected: {
          commandId: id,
          correlationId: id,
          code: 'not_found',
          message: 'The order does not exist',
        },
      }),
    )
    .finish();
  const result = decode('cafe:ordering', body);
  assert.ok('type' in result);
  assert.equal(result.commandId, id);
  assert.throws(() => decode('cafe:loyalty', body), /source/);
});

void test('snapshot publications still require a positive revision and matching root', () => {
  const body = wire
    .encode(
      wire.create({
        eventId: id,
        contractVersion: 1,
        context: 'ordering',
        aggregateKind: 'order',
        aggregateId: id,
        revision: 0,
        order: { id },
      }),
    )
    .finish();
  assert.throws(() => decode('cafe:ordering', body), /snapshot/);
});
