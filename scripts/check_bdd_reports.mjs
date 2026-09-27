import assert from 'node:assert/strict';
import {readFileSync} from 'node:fs';
import {catalogue} from './check_specifications.mjs';

const lane = process.argv[2];
assert.ok(['fast', 'infrastructure'].includes(lane), 'Choose fast or infrastructure');
const reportNames = lane === 'fast' ? ['menu', 'ordering', 'operations', 'engagement'] : ['ordering-postgres', 'workflows'];
const expected = catalogue().filter(row => lane === 'fast' ? row.lane === 'fast' : row.lane !== 'fast');
const executed = new Map();
for (const name of reportNames) {
  const report = JSON.parse(readFileSync(`.local/bdd/${name}.json`, 'utf8'));
  assert.ok(report.length, name+': empty report');
  for (const feature of report) {
    for (const element of feature.elements ?? []) {
      if (element.type === 'background') continue;
      const identity = element.tags.map(tag => tag.name.replace(/^@/, '')).filter(tag => /^[A-Z]+_\d{3}$/.test(tag));
      assert.equal(identity.length, 1, name+': missing scenario identity');
      const id = '@'+identity[0];
      assert.ok(element.steps?.length, id+': no executed steps');
      for (const step of [...element.before ?? [], ...element.steps, ...element.after ?? []]) {
        assert.equal(step.result?.status, 'passed', `${id}: ${step.name ?? 'hook'} did not pass`);
      }
      executed.set(id, (executed.get(id) ?? 0) + 1);
    }
  }
}
const definitions = expected.flatMap(row => row.definitions);
assert.equal(executed.size, definitions.length, 'Executed scenario identities differ from the catalogue');
for (const definition of definitions) {
  assert.equal(executed.get(definition.id), definition.expanded, `${definition.id}: examples were omitted or repeated`);
}
console.log(`${[...executed.values()].reduce((sum, count) => sum + count, 0)} ${lane} Gherkin scenarios passed; reports match the catalogue`);
