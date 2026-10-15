import assert from 'node:assert/strict';
import { execFileSync } from 'node:child_process';
import { existsSync, readFileSync, readdirSync } from 'node:fs';
import path from 'node:path';
import type { SidebarGroup, SidebarItem } from '@rspress/core';
import { navigation, sidebar } from './navigation';
import { checkConceptLanguages, checkExamples } from './check-examples';

const output = path.resolve('.local/docs/generated');
const pages = execFileSync('git', ['ls-files', '--cached', '--others', '--exclude-standard', '-z'], {
  encoding: 'utf8',
})
  .split('\0')
  .filter((file) => file.endsWith('.md') && existsSync(file));

let examples = 0;
for (const page of pages) {
  assert.ok(existsSync(path.join(output, page.replace(/\.md$/, '.html'))), `Missing page: ${page}`);
  const count = checkExamples(readFileSync(page, 'utf8'), path.resolve(page), process.cwd());
  examples += count;
  if (page.startsWith('docs/education/concepts/') && !page.endsWith('/index.md')) {
    assert.ok(count > 0, `Concept needs a source-backed code example: ${page}`);
    checkConceptLanguages(readFileSync(page, 'utf8'), page);
  }
}

/** Check navigation separately because Rspress's Markdown checks only inspect page links. */
function checkLink(link: string): void {
  if (!link.startsWith('/')) return;
  const route = link.split('#')[0]!.replace(/\/$/, '/index');
  assert.ok(existsSync(path.join(output, `${route}.html`)), `Missing navigation route: ${link}`);
}

function checkItems(items: (SidebarGroup | SidebarItem)[]): void {
  for (const item of items) {
    if (item.link) checkLink(item.link);
    if ('items' in item) checkItems(item.items as (SidebarGroup | SidebarItem)[]);
  }
}

for (const item of navigation) {
  if ('link' in item) checkLink(item.link);
}
for (const items of Object.values(sidebar)) checkItems(items as (SidebarGroup | SidebarItem)[]);

assert.match(
  readFileSync(path.join(output, 'index.html'), 'utf8'),
  /Follow a command from browser to domain/,
);
assert.match(readFileSync(path.join(output, 'README.md'), 'utf8'), /```mermaid/);
assert.match(readFileSync(path.join(output, 'llms-full.txt'), 'utf8'), /Café architecture reference/);
assert.ok(readdirSync(path.join(output, 'static')).some((file) => file.startsWith('search_index.')));
console.log(
  `Documentation output verified: ${pages.length} pages, ${examples} source excerpts, navigation, search and Markdown exports.`,
);
