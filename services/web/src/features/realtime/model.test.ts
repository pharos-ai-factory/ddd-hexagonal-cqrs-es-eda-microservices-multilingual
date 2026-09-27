import {test} from 'node:test';
import assert from 'node:assert/strict';
import {merge, type Projection} from './model';
test('late query responses and duplicate deliveries cannot roll back a newer root', () => {
  const latest: Projection<'drink'> = {kind: 'drink', version: 3,
    state: {id: 'one', name: 'Coffee', revision: 1, published: true}};
  const state = merge({}, latest);
  assert.equal(merge(state, {...latest, version: 2}), state);
  assert.equal(merge(state, latest), state);
  assert.equal(Object.keys(merge(state, {...latest, state: {...latest.state, id: 'two'}})).length, 2);
});
