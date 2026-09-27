import assert from 'node:assert/strict';
import {readFileSync, readdirSync} from 'node:fs';
import {resolve} from 'node:path';
import {fileURLToPath} from 'node:url';
import {generateMessages} from '@cucumber/gherkin';
import {IdGenerator, SourceMediaType} from '@cucumber/messages';

// Every permitted lane has a runner in verify.py or integration.py.
export const owners = {
  menu: {prefix: 'MENU', lanes: ['fast'], runner: 'Godog'},
  ordering: {prefix: 'ORDER', lanes: ['fast', 'postgres'], runner: 'Godog'},
  preparation: {prefix: 'PREP', lanes: ['fast'], runner: 'pytest-bdd'},
  collection: {prefix: 'COLLECT', lanes: ['fast'], runner: 'pytest-bdd'},
  loyalty: {prefix: 'LOYALTY', lanes: ['fast'], runner: 'Cucumber'},
  communication: {prefix: 'COMMS', lanes: ['fast'], runner: 'Cucumber'},
  workflows: {prefix: 'FLOW', lanes: ['integration'], runner: 'Cucumber'},
};

export function inspectFeature(source, uri) {
  const parts = uri.split('/');
  assert.equal(parts.length, 3, `${uri}: features belong directly in specifications/<owner>`);
  const owner = parts[1], policy = owners[owner];
  assert.ok(policy, `${uri}: unknown owner`);
  const messages = generateMessages(source, uri, SourceMediaType.TEXT_X_CUCUMBER_GHERKIN_PLAIN,
    {newId: IdGenerator.incrementing(), includeGherkinDocument: true, includePickles: true});
  const errors = messages.filter(message => message.parseError).map(message => message.parseError.message);
  assert.deepEqual(errors, [], `${uri}: invalid Gherkin`);
  const feature = messages.find(message => message.gherkinDocument)?.gherkinDocument.feature;
  const pickles = messages.filter(message => message.pickle).map(message => message.pickle);
  assert.ok(feature && pickles.length, `${uri}: empty features cannot pass`);
  const featureTags = feature.tags.map(tag => tag.name);
  assert.ok(featureTags.includes('@'+owner), `${uri}: missing owner tag`);
  const lanes = policy.lanes.filter(lane => featureTags.includes('@'+lane));
  assert.equal(lanes.length, 1, `${uri}: exactly one supported lane is required`);
  assert.equal(featureTags.length, 2, `${uri}: only owner and lane belong on Feature`);
  const definitions = [];
  const visit = children => {
    for (const child of children) {
      if (child.rule) visit(child.rule.children);
      if (!child.scenario) continue;
      const scenario = child.scenario;
      assert.equal(scenario.tags.length, 1, `${uri}: every scenario needs exactly one stable identity tag`);
      const id = scenario.tags[0].name;
      assert.match(id, new RegExp(`^@${policy.prefix}_[0-9]{3}$`), `${uri}: invalid scenario identity`);
      assert.ok(scenario.steps.some(step => step.keywordType === 'Action'), `${id}: missing When`);
      assert.ok(scenario.steps.some(step => step.keywordType === 'Outcome'), `${id}: missing Then`);
      if (['Scenario Outline', 'Scenario Template'].includes(scenario.keyword)) {
        assert.ok(scenario.examples.some(example => example.tableBody.length), `${id}: outline needs example rows`);
      }
      assert.ok(pickles.some(pickle => pickle.astNodeIds.includes(scenario.id)), `${id}: outline has no executable examples`);
      definitions.push({id, name: scenario.name, expanded: pickles.filter(pickle => pickle.astNodeIds.includes(scenario.id)).length});
    }
  };
  visit(feature.children);
  for (const pickle of pickles) {
    assert.deepEqual(pickle.tags.map(tag => tag.name).filter(tag => !featureTags.includes(tag)).length, 1,
      `${uri}: examples/rules must not add execution filters`);
    assert.ok(pickle.steps.every(step => !/<[^>]+>/.test(step.text)), `${uri}: unresolved example parameter`);
  }
  return {uri, owner, lane: lanes[0], runner: policy.runner, definitions, expanded: pickles.length};
}

export function catalogue() {
  const rows = [];
  for (const directory of readdirSync('specifications', {withFileTypes: true})) {
    if (!directory.isDirectory()) continue;
    assert.ok(owners[directory.name], 'Unknown specification owner '+directory.name);
    for (const file of readdirSync('specifications/'+directory.name, {withFileTypes: true})) {
      assert.ok(file.isFile() && file.name.endsWith('.feature'), 'Only feature files belong in context specification directories');
      const uri = `specifications/${directory.name}/${file.name}`;
      rows.push(inspectFeature(readFileSync(uri, 'utf8'), uri));
    }
  }
  const identities = rows.flatMap(row => row.definitions.map(definition => definition.id));
  assert.equal(new Set(identities).size, identities.length, 'Scenario identities must be unique across the catalogue');
  for (const owner of Object.keys(owners)) assert.ok(rows.some(row => row.owner === owner), `Missing scenarios for ${owner}`);
  return rows;
}

if (process.argv[1] && resolve(process.argv[1]) === fileURLToPath(import.meta.url)) {
  const rows = catalogue();
  console.table(rows.map(({uri, lane, runner, definitions, expanded}) => ({file: uri, lane, runner, scenarios: definitions.length, examples: expanded})));
  console.log(`${rows.reduce((count, row) => count + row.expanded, 0)} executable scenarios including outline examples; catalogue valid`);
}
