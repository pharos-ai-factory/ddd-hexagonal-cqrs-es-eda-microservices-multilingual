import {test} from 'node:test';
import assert from 'node:assert/strict';
import {outcome} from './receipts.js';
import {CorruptState} from '../foundation/domain.js';

const id = '00000000-0000-4000-8000-000000000001';
const valid = {aggregateId: id, version: 1, status: 'active'};

test('recorded outcomes require a target, integer version and complete success or rejection', () => {
  for (const value of [
    null, [], {}, {...valid, aggregateId: '00000000-0000-4000-8000-000000000002'},
    ...[null, true, '1', -1, 1.5, Number.MAX_SAFE_INTEGER+1].map(version => ({...valid, version})),
    ...[null, true, ''].map(status => ({...valid, status})),
    ...[null, {}, {code: 'missing'}, {code: '', message: 'Missing'}].map(rejection => ({...valid, rejection})),
  ]) assert.throws(() => outcome(value, id), CorruptState);
  assert.deepEqual(outcome(valid, id), valid);
  const rejected = {...valid, version: 0, status: '', rejection: {code: 'not_found', message: 'Missing root'}};
  assert.deepEqual(outcome(rejected, id), rejected);
});
