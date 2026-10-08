import {test} from 'node:test';
import assert from 'node:assert/strict';
import protobuf from 'protobufjs';
import schema from './generated/requests.json' with {type: 'json'};
import {decodeRequest, Registry, requiredInputs} from './requests.js';
import type {Metadata} from '../foundation/application.js';
const root = protobuf.Root.fromJSON(schema), requestType = root.lookupType('cafe.loyalty.requests.v1.Request'), replyType = root.lookupType('cafe.loyalty.requests.v1.Reply');
const id = '11111111-1111-4111-8111-111111111111';
const encode = (value: object) => Buffer.from(requestType.encode(requestType.fromObject(value)).finish());
const wire = {contractVersion: 1, requestId: id, context: 'loyalty', command: {
  metadata: {commandId: id, aggregateId: id, expectedVersion: 0, correlationId: id}, redeemReward: {orderId: id},
}};
test('wire command becomes a plain application command with explicit zero version', async () => {
  const registry = new Registry('loyalty');
  let captured: Metadata | undefined;
  registry.command('redeemReward', (value: unknown) => {
    assert.deepEqual(value, {orderId: id}); return {orderId: id};
  }, async metadata => { captured = metadata; return {aggregateId: id, version: 1, status: 'redeemed'}; });
  const reply = replyType.toObject(replyType.decode(await registry.handle(decodeRequest(encode(wire), 'loyalty'))), {longs: Number});
  assert.equal(captured?.expected, 0);
  assert.equal(captured?.name, 'loyalty.RedeemReward');
  assert.deepEqual(captured?.input, {orderId: id});
  assert.equal(reply.outcome.version, 1);
});
test('foreign context, missing command fields and missing expected version fail before a handler', () => {
  assert.throws(() => decodeRequest(encode(wire), 'communication'));
  assert.throws(() => decodeRequest(encode({...wire, command: {...wire.command, redeemReward: {}}}), 'loyalty'));
  assert.throws(() => decodeRequest(encode({...wire, command: {...wire.command, metadata: {...wire.command.metadata, expectedVersion: undefined}}}), 'loyalty'));
});
test('owner query pagination and an absent next identity remain explicit', async () => {
  const registry = new Registry('communication');
  registry.queries('notification', 'notifications', {
    get: async () => undefined, list: async () => [], page: async request => {
      assert.deepEqual(request, {limit: 1, after: id}); return {items: []};
    },
  }, value => value);
  const type = root.lookupType('cafe.communication.requests.v1.Request');
  const request = decodeRequest(Buffer.from(type.encode(type.fromObject({contractVersion: 1, requestId: id, context: 'communication', query: {listNotifications: {page: {limit: 1, after: id}}}})).finish()), 'communication');
  const replyType = root.lookupType('cafe.communication.requests.v1.Reply');
  const reply = replyType.toObject(replyType.decode(await registry.handle(request)), {longs: Number, defaults: true});
  assert.deepEqual(reply.notifications.items, []);
  assert.equal(reply.notifications.paged, true);
  assert.equal(reply.notifications.nextId, undefined);
});

test('future optional fields remain optional and required zero values remain valid', () => {
  const type = new protobuf.Type('FutureInput')
    .add(new protobuf.Field('price', 1, 'int64', undefined, undefined, {'(cafe.requests.v1.required_input)': true}))
    .add(new protobuf.Field('note', 2, 'string'));
  assert.throws(() => requiredInputs(type, {}), /Missing command field/);
  requiredInputs(type, {price: 0});
  requiredInputs(type, {price: 0, note: 'Optional addition'});
});
