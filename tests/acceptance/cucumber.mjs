export default {
  paths: ['specifications/workflows/*.feature'],
  import: ['tests/acceptance/*.steps.ts', 'tests/acceptance/world.ts'],
  tags: '@integration',
  format: ['progress', ...(process.argv.includes('--dry-run') ? [] : ['json:.local/bdd/workflows.json'])],
  strict: true,
  publish: false,
  parallel: 0,
};
