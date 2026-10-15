import { existsSync, statSync } from 'node:fs';
import path from 'node:path';
import type { RsbuildPlugin } from '@rsbuild/core';
import { unified } from 'unified';
import remarkParse from 'remark-parse';
import remarkGfm from 'remark-gfm';
import remarkFrontmatter from 'remark-frontmatter';
import remarkStringify from 'remark-stringify';
import { visit } from 'unist-util-visit';

export const repositoryUrl =
  'https://github.com/pharos-ai-factory/ddd-hexagonal-cqrs-es-eda-microservices-multilingual';
const markdown = unified().use(remarkParse).use(remarkGfm).use(remarkFrontmatter).use(remarkStringify);

/** Keep documentation links local and resolve code/file links against the repository. */
export function rewriteRepositoryLinks(source: string, filename: string, root: string): string {
  const tree = markdown.parse(source);
  visit(tree, (node) => {
    if (node.type !== 'link' && node.type !== 'definition') return;
    if (/^(?:[a-z][a-z\d+.-]*:|\/\/|#)/i.test(node.url)) return;
    const match = /^([^?#]*)(.*)$/.exec(node.url);
    const target = match?.[1];
    if (!target) return;
    const absolute = target.startsWith('/')
      ? path.resolve(root, `.${decodeURIComponent(target)}`)
      : path.resolve(path.dirname(filename), decodeURIComponent(target));
    const relative = path.relative(root, absolute);
    if (relative.startsWith('..') || path.isAbsolute(relative) || !existsSync(absolute)) {
      throw new Error(`Missing repository link in ${filename}: ${node.url}`);
    }
    if (/\.md$/.test(target)) return;
    const directory = statSync(absolute).isDirectory();
    const index = path.join(absolute, 'README.md');
    if (directory && existsSync(index)) {
      node.url = `/${relative ? `${relative}/` : ''}README.md${match?.[2] ?? ''}`;
      return;
    }
    const encoded = relative.split(path.sep).map(encodeURIComponent).join('/');
    node.url = `${repositoryUrl}/${directory ? 'tree' : 'blob'}/main/${encoded}${match?.[2] ?? ''}`;
  });
  return markdown.stringify(tree);
}

/** Transform Markdown before Rspress checks links, for both HTML and Markdown output. */
export function repositoryLinks(): RsbuildPlugin {
  return {
    name: 'cafe-repository-links',
    setup(api) {
      api.transform({ test: /\.md$/, order: 'pre' }, ({ code, resourcePath }) =>
        rewriteRepositoryLinks(code, resourcePath, api.context.rootPath),
      );
    },
  };
}
