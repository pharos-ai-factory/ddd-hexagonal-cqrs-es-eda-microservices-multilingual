import js from '@eslint/js';
import tseslint from 'typescript-eslint';
import { fileURLToPath } from 'node:url';

export default tseslint.config(
  {
    ignores: [
      '**/generated/**',
      '**/*_generated.*',
      '**/node_modules/**',
      '**/dist/**',
      '**/.next/**',
      '**/out/**',
      '.local/**',
      '.tools/**',
      '**/next-env.d.ts',
    ],
  },
  {
    files: [
      'services/**/*.ts',
      'services/**/*.tsx',
      'services/**/*.mts',
      'tests/**/*.ts',
      'scripts/browser/**/*.ts',
    ],
    extends: [js.configs.recommended, ...tseslint.configs.recommendedTypeChecked],
    languageOptions: {
      parserOptions: {
        project: [
          'services/engagement/tsconfig.check.json',
          'services/web/tsconfig.json',
          'tests/acceptance/tsconfig.json',
          'tests/infrastructure/tsconfig.json',
          'scripts/browser/tsconfig.json',
        ],
        tsconfigRootDir: fileURLToPath(new URL('.', import.meta.url)),
      },
    },
    rules: {
      '@typescript-eslint/switch-exhaustiveness-check': 'error',
      '@typescript-eslint/no-unused-vars': ['error', { argsIgnorePattern: '^_' }],
      // Named application ports deliberately extend generic repository capabilities.
      '@typescript-eslint/no-empty-object-type': ['error', { allowInterfaces: 'with-single-extends' }],
    },
  },
  {
    files: ['**/*.test.ts', '**/*.integration.ts', '**/*.spec.ts', '**/tests/**/*.ts', '**/tests/**/*.mts'],
    // Test doubles implement asynchronous ports without introducing artificial I/O.
    rules: { '@typescript-eslint/require-await': 'off' },
  },
);
