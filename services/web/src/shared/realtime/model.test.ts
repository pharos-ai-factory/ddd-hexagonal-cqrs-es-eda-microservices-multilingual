import { test } from 'node:test';
import assert from 'node:assert/strict';
import { mergeState, type VersionedState } from './model';
void test('late query responses and duplicate deliveries cannot roll back a newer root', () => {
  const latest: VersionedState<'drink'> = {
    kind: 'drink',
    version: 3,
    state: { id: 'one', name: 'Coffee', revision: 1, published: true },
  };
  const state = mergeState({}, latest);
  assert.equal(mergeState(state, { ...latest, version: 2 }), state);
  assert.equal(mergeState(state, latest), state);
  assert.equal(
    Object.keys(mergeState(state, { ...latest, state: { ...latest.state, id: 'two' } })).length,
    2,
  );
});
