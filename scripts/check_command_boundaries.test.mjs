import {test} from 'node:test';
import assert from 'node:assert/strict';
import ts from 'typescript';
import path from 'node:path';
import {violations, productionProgram, commandLayout} from './check_command_boundaries.mjs';

test('require one command and its handler per command module', () => {
  const filename = 'contexts/loyalty/application/commands/credit-collection.ts';
  const pair = 'type CreditCollectionCommand = {}; class CreditCollectionCommandHandler {}';
  const checkLayout = (source, name = filename) => commandLayout(ts.createSourceFile(name, source, ts.ScriptTarget.Latest, true));
  assert.deepEqual(checkLayout(pair), []);
  assert.deepEqual(checkLayout('class RewardNotFoundError {}'), []);
  assert.ok(checkLayout(pair, 'contexts/loyalty/application/commands.ts').length);
  assert.ok(checkLayout('type CreditCollectionCommand = {};').length);
  assert.ok(checkLayout(pair+' type IssueRewardCommand = {}; class IssueRewardCommandHandler {}').length);
});

const base = productionProgram();
function check(body, location = 'application') {
  const fixture = path.resolve('services/engagement/src/contexts/loyalty/'+location+'/boundary-fixture.ts');
  let source = `import {LoyaltyAccount as Account} from '../domain/loyalty-account.js';
import type {Metadata} from '../../../foundation/application.js';
import type {AggregateTransaction} from '../../../adaptors/command-execution.js';
import type {WriteRepository} from '../../../foundation/write-repository.js';
import type {AccountState} from '../domain/loyalty-account.js';\n${body}`;
  if (location !== 'application') source = source.replaceAll("'../domain/", "'../../domain/").replaceAll("'../../../foundation/", "'../../../../foundation/").replaceAll("'../../../adaptors/", "'../../../../adaptors/");
  const host = ts.createCompilerHost(base.getCompilerOptions()), original = host.getSourceFile.bind(host);
  host.getSourceFile = (name, language, onError, create) => name === fixture
    ? ts.createSourceFile(name, source, language, true) : original(name, language, onError, create);
  const program = ts.createProgram([...base.getRootFileNames(), fixture], base.getCompilerOptions(), host);
  const diagnostics = ts.getPreEmitDiagnostics(program).filter(item => item.file?.fileName === fixture);
  assert.deepEqual(diagnostics.map(d => ts.flattenDiagnosticMessageText(d.messageText, '\n')), []);
  return violations(program, name => name === fixture);
}
test('reject event-handler mutation, including renamed imports and captured methods', () => {
  for (const expression of ['account.credit("order", "grant")', 'const action = account.credit; action("order", "grant")', 'account["credit"]("order", "grant")'])
    assert.ok(check(`class CollectedIntegrationEventHandler { handle(account: Account) { ${expression}; } }`).length);
});
test('reject direct aggregate-store access and mutation hidden in an extra handler method', () => {
  assert.ok(check('class Reaction { constructor(private store: AggregateTransaction<AccountState>) {} handle() { const execute = this.store.execute; } }').length);
  assert.ok(check('class CreditCommandHandler { handle(account: Account) { account.credit("order", "grant"); } }').length);
});
test('allow aggregate behaviour only inside command execution, and read-only access elsewhere', () => {
  assert.deepEqual(check('class CreditCommandHandler { execute(account: Account) { account.credit("order", "grant"); } }'), []);
  assert.deepEqual(check('class AccountQueryHandler { handle(account: Account) { return account.snapshot(); } }'), []);
});
test('reject synchronous handler chaining and repeated store execution', () => {
  for (const body of [
    `class CreditCommandHandler { execute() {} }
     class EventHandler { constructor(private next: CreditCommandHandler) {} handle() { this.next.execute(); } }`,
    `class CreditCommandHandler { execute() {} }
     class SecondCommandHandler { execute(next: CreditCommandHandler) { const action = next.execute; action(); } }`,
    `class CreditCommandHandler { execute(store: AggregateTransaction<AccountState>, m: Metadata) {
      store.execute(m, () => {throw Error()}); store.execute(m, () => {throw Error()}); } }`,
    `class CreditCommandHandler { execute(store: AggregateTransaction<AccountState>, m: Metadata) {
      for (const id of [1,2]) store.execute(m, () => {throw Error()}); } }`,
    `class CreditCommandHandler { execute(store: AggregateTransaction<AccountState>) { const run = store.execute; } }`,
  ]) assert.ok(check(body).length, body);
});

test('queries and event reactions remain separate discoverable use cases', () => {
  const check = (source, folder) => commandLayout(ts.createSourceFile(`contexts/loyalty/application/${folder}/fixture.ts`, source, ts.ScriptTarget.Latest, true));
  const pair = 'type GetAccountQuery = {}; class GetAccountQueryHandler {}';
  assert.deepEqual(check(pair, 'queries'), []);
  assert.ok(check(pair, 'commands').length);
  assert.ok(check(pair+' type ListAccountsQuery = {}; class ListAccountsQueryHandler {}', 'queries').length);
  assert.ok(check('class FirstIntegrationEventHandler {} class SecondIntegrationEventHandler {}', 'event-handlers').length);
  assert.deepEqual(check('class OrderCollectedIntegrationEventHandler {}', 'event-handlers'), []);
});

test('persistence can restore aggregate snapshots but cannot change aggregate behaviour', () => {
  const restore = 'const read = (state: AccountState) => new Account(state).snapshot();';
  assert.deepEqual(check(restore, 'adaptors/persistence'), []);
  assert.ok(check('const write = (account: Account) => account.credit("order", "grant");', 'adaptors/persistence').length);
  assert.ok(check(restore).length);
});

test('write repositories cannot grant query or event handlers mutation authority', () => {
  for (const body of [
    'await repository.save(account);',
    'const save = repository.save; await save(account);',
    'const loaded = await repository.get("id"); loaded?.state.credit("order", "grant");',
  ]) assert.ok(check(`class AccountQueryHandler { async execute(repository: WriteRepository<Account>, account: Account) { ${body} } }`).length);
  assert.deepEqual(check('class AccountQueryHandler { async execute(repository: WriteRepository<Account>) { return (await repository.get("id"))?.state.snapshot(); } }'), []);
});
