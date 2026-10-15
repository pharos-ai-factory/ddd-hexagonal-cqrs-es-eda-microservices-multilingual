import assert from 'node:assert/strict';
import { mkdtempSync, rmSync, writeFileSync } from 'node:fs';
import os from 'node:os';
import path from 'node:path';
import { test } from 'node:test';
import { checkConceptLanguages, checkExamples } from './check-examples';

void test('teaching excerpts match real source and detect drift or missing evidence', () => {
  const root = mkdtempSync(path.join(os.tmpdir(), 'cafe-doc-examples-'));
  const filename = path.join(root, 'guide.md');
  const block = (code: string, source = './handler.ts') => `\`\`\`ts source="${source}"\n${code}\n\`\`\``;
  try {
    writeFileSync(path.join(root, 'handler.ts'), 'function save() {\n  return repository.save(root);\n}');
    assert.equal(checkExamples(block('return repository.save( root );'), filename, root), 1);
    assert.equal(checkExamples('```ts\n// Illustrative example\n```', filename, root), 0);
    assert.throws(() => checkExamples(block('return repository.delete(root);'), filename, root), /Stale/);
    assert.throws(() => checkExamples(block(''), filename, root), /Empty/);
    assert.throws(() => checkExamples(block('code', './missing.ts'), filename, root), /ENOENT/);
    assert.throws(() => checkExamples(block('code', '../outside.ts'), filename, root), /External/);
  } finally {
    rmSync(root, { recursive: true, force: true });
  }
});

void test('each concept includes Go, Python and TypeScript examples', () => {
  const content = ['go', 'python', 'ts'].map((language) => '```' + language + '\nexample\n```').join('\n');
  assert.doesNotThrow(() => checkConceptLanguages(content, 'concept.md'));
  assert.throws(() => checkConceptLanguages(content.replace('```go', '```text'), 'concept.md'), /go example/);
});
