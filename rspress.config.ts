import { defineConfig } from '@rspress/core';
import mermaid from 'rspress-plugin-mermaid';
import { navigation, sidebar } from './scripts/docs/navigation';
import { repositoryLinks, repositoryUrl } from './scripts/docs/repository-links';
import { documentationWatcher } from './scripts/docs/watch-docs';

export default defineConfig({
  root: '.',
  title: 'Café architecture reference',
  description: 'DDD, hexagonal architecture, CQRS and durable workflows across Go, Python and TypeScript.',
  lang: 'en',
  outDir: '.local/docs/generated',
  route: {
    extensions: ['.md', '.mdx'],
    exclude: [
      '**/node_modules/**',
      '**/.venv/**',
      '**/generated/**',
      '**/dist/**',
      '**/out/**',
      '**/.next/**',
      '.local/**',
      '.tools/**',
      '.rspress/**',
      'test-results/**',
      'playwright-report/**',
    ],
    cleanUrls: true,
  },
  llms: true,
  plugins: [mermaid({ mermaidConfig: { securityLevel: 'strict' } })],
  builderConfig: {
    plugins: [repositoryLinks(), documentationWatcher()],
    server: { host: '127.0.0.1', port: 3001, strictPort: true },
  },
  themeConfig: {
    nav: navigation,
    sidebar,
    editLink: { docRepoBaseUrl: `${repositoryUrl}/edit/main` },
    socialLinks: [{ icon: 'github', mode: 'link', content: repositoryUrl }],
  },
});
