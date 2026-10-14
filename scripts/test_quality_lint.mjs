import assert from 'node:assert/strict';
import { mkdtemp, mkdir, writeFile, rm } from 'node:fs/promises';
import { tmpdir } from 'node:os';
import path from 'node:path';
import { test } from 'node:test';
import { ESLint } from 'eslint';
import configuration from '../eslint.config.mjs';

test('typed lint rejects unhandled promises, unsafe values and incomplete switches', async () => {
  const root = await mkdtemp(path.join(tmpdir(), 'cafe-lint-'));
  try {
    const folder = path.join(root, 'services/web/src');
    await mkdir(folder, { recursive: true });
    const file = path.join(folder, 'fixture.ts');
    const project = path.join(root, 'tsconfig.json');
    await writeFile(
      project,
      JSON.stringify({ compilerOptions: { strict: true, target: 'ES2023' }, include: ['services/**/*.ts'] }),
    );
    await writeFile(
      file,
      `declare function deliver(): Promise<void>;
deliver();
declare const untyped: any;
export const leaked: string = untyped;
export function label(state: 'ready' | 'sent') {
  switch (state) { case 'ready': return 'Ready'; }
}
`,
    );
    const eslint = new ESLint({
      cwd: root,
      overrideConfigFile: true,
      overrideConfig: configuration.map((config) =>
        config.languageOptions?.parserOptions
          ? {
              ...config,
              languageOptions: {
                ...config.languageOptions,
                parserOptions: { project, tsconfigRootDir: root },
              },
            }
          : config,
      ),
    });
    const [result] = await eslint.lintFiles([file]);
    assert.equal(result.fatalErrorCount, 0, JSON.stringify(result.messages));
    const rules = new Set(result.messages.map((message) => message.ruleId));
    for (const rule of ['no-floating-promises', 'no-unsafe-assignment', 'switch-exhaustiveness-check']) {
      assert.ok(rules.has('@typescript-eslint/' + rule), JSON.stringify(result.messages));
    }
  } finally {
    await rm(root, { recursive: true, force: true });
  }
});
