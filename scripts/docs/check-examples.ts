import assert from 'node:assert/strict';
import { readFileSync } from 'node:fs';
import path from 'node:path';
import { unified } from 'unified';
import remarkParse from 'remark-parse';
import { visit } from 'unist-util-visit';

const markdown = unified().use(remarkParse);
const compact = (value: string) => value.replace(/\s+/g, '');

/** Compare source-labelled teaching excerpts with the checkout, allowing presentation whitespace. */
export function checkExamples(content: string, filename: string, root: string): number {
  let count = 0;
  visit(markdown.parse(content), 'code', (node) => {
    const source = /(?:^|\s)source="([^"]+)"/.exec(node.meta ?? '')?.[1];
    if (!source) return;
    const target = path.resolve(path.dirname(filename), source);
    const relative = path.relative(root, target);
    assert.ok(!relative.startsWith('..') && !path.isAbsolute(relative), `External example: ${source}`);
    const code = compact(node.value);
    assert.ok(code.length > 0, `Empty source example in ${filename}`);
    assert.ok(
      compact(readFileSync(target, 'utf8')).includes(code),
      `Stale code example in ${filename}: ${source}`,
    );
    count += 1;
  });
  return count;
}

/** Every concept teaches the three supported implementation languages. */
export function checkConceptLanguages(content: string, filename: string): void {
  const languages = new Set<string>();
  visit(markdown.parse(content), 'code', (node) => {
    if (node.lang) languages.add(node.lang);
  });
  for (const language of ['go', 'python', 'ts']) {
    assert.ok(languages.has(language), `Concept needs a ${language} example: ${filename}`);
  }
}
