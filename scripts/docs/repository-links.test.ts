import assert from 'node:assert/strict';
import { mkdtempSync, mkdirSync, rmSync, writeFileSync } from 'node:fs';
import os from 'node:os';
import path from 'node:path';
import { test } from 'node:test';
import { repositoryUrl, rewriteRepositoryLinks } from './repository-links';

void test('Markdown and code links retain their destinations across the site and repository', () => {
  const root = mkdtempSync(path.join(os.tmpdir(), 'cafe-docs-'));
  try {
    mkdirSync(path.join(root, 'docs'));
    mkdirSync(path.join(root, 'source'));
    writeFileSync(path.join(root, 'README.md'), '# Home');
    writeFileSync(path.join(root, 'docs/guide.md'), '# Guide');
    writeFileSync(path.join(root, 'source/example.ts'), 'export {};');
    writeFileSync(path.join(root, 'source/space name.ts'), 'export {};');
    const source = [
      '---\ntitle: Example\n---',
      '[Guide](guide.md#usage)',
      '[Home](..)',
      '[Code](../source/example.ts#L1)',
      '[Folder](../source)',
      '[Space](../source/space%20name.ts)',
      '[External](https://example.com/guide)',
      '[Section](#usage)',
      '[Reference][code]',
      '[code]: ../source/example.ts?plain=1#L1',
      '| Column |\n| --- |\n| Value |',
      '```mermaid\nflowchart LR\n  A --> B\n```',
      '```md\n[Example](missing.md)\n```',
    ].join('\n\n');
    const output = rewriteRepositoryLinks(source, path.join(root, 'docs/guide.md'), root);
    for (const expected of [
      '---\ntitle: Example\n---',
      '[Guide](guide.md#usage)',
      '[Home](/README.md)',
      `[Code](${repositoryUrl}/blob/main/source/example.ts#L1)`,
      `[Folder](${repositoryUrl}/tree/main/source)`,
      `[Space](${repositoryUrl}/blob/main/source/space%20name.ts)`,
      `[code]: ${repositoryUrl}/blob/main/source/example.ts?plain=1#L1`,
      '[External](https://example.com/guide)',
      '[Section](#usage)',
      '| Value',
      '```mermaid\nflowchart LR\n  A --> B\n```',
      '[Example](missing.md)',
    ]) {
      assert.ok(output.includes(expected), expected);
    }
    for (const target of ['missing.md', '../source/missing.ts', '../missing-directory', '../../outside']) {
      assert.throws(
        () => rewriteRepositoryLinks(`[Missing](${target})`, path.join(root, 'docs/guide.md'), root),
        /Missing repository link/,
      );
    }
  } finally {
    rmSync(root, { recursive: true, force: true });
  }
});
