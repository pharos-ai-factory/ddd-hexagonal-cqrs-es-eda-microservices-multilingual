export default {
  paths: ['../../specifications/loyalty/*.feature', '../../specifications/communication/*.feature'],
  import: ['tests/bdd/*.ts'],
  tags: '@fast',
  format: ['progress'],
  strict: true,
  publish: false,
};
