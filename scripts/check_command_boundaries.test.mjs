import {test} from 'node:test';
import assert from 'node:assert/strict';
import ts from 'typescript';
import path from 'node:path';
import {violations, productionProgram} from './check_command_boundaries.mjs';

const base = productionProgram();
const fixture = path.resolve('services/engagement/src/contexts/loyalty/application/boundary-fixture.ts');
function check(body) {
  const source = `import {LoyaltyAccount as Account} from '../domain/account.js';
import type {AggregateCommandPort, Metadata} from '../../../foundation/application.js';
import type {AccountState} from '../domain/account.js';\n${body}`;
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
  assert.ok(check('class Reaction { constructor(private store: AggregateCommandPort<AccountState>) {} handle() { const execute = this.store.execute; } }').length);
  assert.ok(check('class CreditCommandHandler { handle(account: Account) { account.credit("order", "grant"); } }').length);
});
test('allow aggregate behaviour only inside command execution, and read-only access elsewhere', () => {
  assert.deepEqual(check('class CreditCommandHandler { execute(account: Account) { account.credit("order", "grant"); } }'), []);
  assert.deepEqual(check('class AccountQueryHandler { handle(account: Account) { return account.snapshot(); } }'), []);
});
