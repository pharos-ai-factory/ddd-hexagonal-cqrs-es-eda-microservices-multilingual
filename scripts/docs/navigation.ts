import { readdirSync, readFileSync } from 'node:fs';
import type { NavItem, Sidebar, SidebarItem } from '@rspress/core';
import contexts from '../../docs/contexts.json';

/** Discover existing document titles without creating a second content catalogue. */
function pages(directory: string): SidebarItem[] {
  return readdirSync(directory)
    .filter((file) => file.endsWith('.md') && file !== 'README.md')
    .sort()
    .map((file) => ({
      text: readFileSync(`${directory}/${file}`, 'utf8').match(/^# (.+)$/m)?.[1] ?? file,
      link: `/${directory}/${file.slice(0, -3)}`,
    }));
}

export const navigation: NavItem[] = [
  { text: 'Guide', link: '/' },
  { text: 'Education', link: '/docs/education/' },
  { text: 'Architecture', link: '/ARCHITECTURE' },
  { text: 'Contexts', link: `/${contexts[0]!.source}/README` },
  { text: 'Contracts', link: '/contracts/README' },
  { text: 'Decisions', link: '/docs/decisions/README' },
];

export const sidebar: Sidebar = {
  '/docs/education/concepts/': [
    { text: 'Concept catalogue', link: '/docs/education/concepts/' },
    { text: 'End-to-end recipe', link: '/docs/education/end-to-end-recipe' },
    ...[
      {
        text: 'Model the business',
        slugs: [
          'bounded-contexts',
          'domain-layer',
          'entities',
          'value-objects',
          'aggregates',
          'invariants',
          'factories',
          'domain-services',
        ],
      },
      {
        text: 'Enter the application',
        slugs: [
          'application-layer',
          'commands',
          'command-handlers',
          'queries',
          'responses',
          'query-handlers',
          'ports',
          'adaptors',
          'dependency-injection',
        ],
      },
      {
        text: 'Restore and persist state',
        slugs: [
          'write-repositories',
          'aggregate-hydration',
          'read-models',
          'read-repositories',
          'transactions',
          'projections',
        ],
      },
      {
        text: 'Deliver facts and updates',
        slugs: [
          'domain-events',
          'domain-event-handlers',
          'integration-events',
          'integration-event-handlers',
          'durable-commands',
          'outbox',
          'receipts-and-idempotency',
          'realtime-events',
          'contracts-and-generation',
        ],
      },
    ].map(({ text, slugs }) => ({
      text,
      items: slugs.map((slug) => ({
        text: readFileSync(`docs/education/concepts/${slug}.md`, 'utf8').match(/^# (.+)$/m)![1]!,
        link: `/docs/education/concepts/${slug}`,
      })),
    })),
    { text: 'Introductory learning path', link: '/docs/education/' },
    { text: 'Implementation coverage and gaps', link: '/docs/education/implementation-coverage' },
  ],
  '/docs/education/': [
    {
      text: 'Learn the architecture',
      items: [
        { text: 'Learning path', link: '/docs/education/' },
        { text: 'End-to-end recipe', link: '/docs/education/end-to-end-recipe' },
        { text: 'Concept reference: code in three languages', link: '/docs/education/concepts/' },
        { text: 'Implementation coverage and gaps', link: '/docs/education/implementation-coverage' },
        { text: '1. Strategic DDD', link: '/docs/education/strategic-ddd' },
        { text: '2. Tactical DDD', link: '/docs/education/tactical-ddd' },
        { text: '3. Hexagonal architecture', link: '/docs/education/hexagonal-architecture' },
        { text: '4. CQRS', link: '/docs/education/cqrs' },
        { text: '5. Event-driven architecture', link: '/docs/education/event-driven-architecture' },
        { text: '6. Consistency and delivery', link: '/docs/education/consistency-and-delivery' },
        { text: '7. Event sourcing', link: '/docs/education/event-sourcing' },
        { text: '8. Modelling and boundaries', link: '/docs/education/modelling-and-boundaries' },
        { text: '9. Testing and exercises', link: '/docs/education/testing-and-exercises' },
        { text: 'Glossary', link: '/docs/education/glossary' },
      ],
    },
    {
      text: 'Explore the implementation',
      items: [
        { text: 'Walkthrough', link: '/docs/walkthrough' },
        { text: 'Reference architecture', link: '/ARCHITECTURE' },
        { text: 'Domain rules', link: '/DDD' },
        { text: 'Accepted decisions', link: '/docs/decisions/README' },
      ],
    },
  ],
  '/': [
    {
      text: 'Start here',
      items: [
        { text: 'Overview', link: '/README' },
        { text: 'Education: learning path', link: '/docs/education/' },
        { text: 'Walkthrough', link: '/docs/walkthrough' },
        { text: 'Contributing', link: '/CONTRIBUTING' },
        { text: 'Developer workflow', link: '/docs/developer-workflow' },
        { text: 'Documentation site', link: '/docs/documentation' },
        { text: 'Testing', link: '/TESTING' },
      ],
    },
    {
      text: 'Architecture',
      items: [
        { text: 'Reference architecture', link: '/ARCHITECTURE' },
        { text: 'Domain model and DDD', link: '/DDD' },
        { text: 'Browser delivery', link: '/CLIENT-SUBSCRIPTIONS' },
        { text: 'Adaptor layout', link: '/docs/adaptor-layout' },
        { text: 'Broker permissions', link: '/docs/broker-permissions' },
        { text: 'Adopting the patterns', link: '/docs/adopting-patterns' },
      ],
    },
    {
      text: 'Bounded contexts',
      items: contexts.map((context) => ({
        text: `${context.name === 'communication' ? 'Customer communication' : context.name[0]!.toUpperCase() + context.name.slice(1)} (${context.service})`,
        link: `/${context.source}/README`,
      })),
    },
    {
      text: 'Contracts and specifications',
      collapsed: true,
      items: [
        { text: 'Published contracts', link: '/contracts/README' },
        { text: 'Generation', link: '/contracts/generation' },
        { text: 'HTTP', link: '/contracts/shared/http_api/README' },
        { text: 'Commands and queries', link: '/contracts/shared/messaging/requests' },
        { text: 'Events', link: '/contracts/shared/messaging/events' },
        { text: 'Realtime', link: '/contracts/shared/realtime/README' },
        { text: 'Executable specifications', link: '/specifications/README' },
      ],
    },
    {
      text: 'Architecture decisions',
      collapsed: true,
      items: [{ text: 'Decision index', link: '/docs/decisions/README' }, ...pages('docs/decisions')],
    },
    {
      text: 'Maintenance',
      collapsed: true,
      items: [
        { text: 'Development secrets', link: '/docs/development-secrets' },
        { text: 'Contract improvements', link: '/docs/contract-improvements' },
        { text: 'File responsibility review', link: '/docs/large-file-review' },
        { text: 'Verification evidence', link: '/docs/verification' },
        { text: 'Changelog', link: '/CHANGELOG' },
        { text: 'Agent instructions', link: '/AGENTS' },
        { text: 'Third-party notices', link: '/THIRD_PARTY_NOTICES' },
      ],
    },
  ],
};
