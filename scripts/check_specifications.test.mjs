import assert from 'node:assert/strict';
import test from 'node:test';
import {inspectFeature, catalogue} from './check_specifications.mjs';

const feature = `@fast @menu
Feature: A publishable menu
  @MENU_999
  Scenario: Reject an empty menu
    When the menu is published
    Then publication is rejected
`;
test('the catalogue covers every owner with unique executable scenario identities', () => {
  assert.equal(new Set(catalogue().map(row => row.owner)).size, 7);
});
test('unrunnable and silently filtered specifications are rejected', () => {
  for (const invalid of [
    feature.replace('@fast', '@integration'),
    feature.replace('@MENU_999', '@pending'),
    feature.replace('@MENU_999', '@MENU_999 @skip'),
    feature.replace('    Then publication is rejected', ''),
    feature.replace('Scenario:', 'Scenario Outline:'),
    '@fast @menu\nFeature: Empty\n',
    feature.replace('Feature:', 'Not Gherkin:'),
  ]) assert.throws(() => inspectFeature(invalid, 'specifications/menu/example.feature'));
});
