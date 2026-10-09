import js from '@eslint/js'
import boundaries from 'eslint-plugin-boundaries'
import reactHooks from 'eslint-plugin-react-hooks'
import globals from 'globals'
import tseslint from 'typescript-eslint'

const WAILS_RUNTIME_MESSAGE = 'Only src/platform/ may import @wailsio/runtime or the generated bindings.'

export default tseslint.config(
  { ignores: ['dist', 'bindings', 'node_modules', 'public'] },
  js.configs.recommended,
  ...tseslint.configs.recommended,
  {
    files: ['**/*.{ts,tsx}'],
    languageOptions: {
      ecmaVersion: 2022,
      globals: globals.browser,
    },
    plugins: {
      'react-hooks': reactHooks,
      boundaries,
    },
    settings: {
      'import/resolver': {
        typescript: { project: './tsconfig.json' },
        node: true,
      },
      'boundaries/include': ['src/**/*', 'bindings/**/*'],
      'boundaries/ignore': ['src/main.tsx'],
      'boundaries/elements': [
        { type: 'test-support', pattern: 'src/test' },
        { type: 'app', pattern: 'src/app' },
        { type: 'feature', pattern: 'src/features/*', capture: ['feature'] },
        { type: 'platform', pattern: 'src/platform' },
        { type: 'shared', pattern: 'src/shared' },
        { type: 'bindings', pattern: 'bindings' },
      ],
      'boundaries/files': [{ category: 'store', pattern: 'src/features/*/store.ts' }],
    },
    rules: {
      ...reactHooks.configs.recommended.rules,
      'boundaries/no-unknown-files': 'error',
      'boundaries/dependencies': [
        'error',
        {
          default: 'disallow',
          policies: [
            {
              from: { element: { type: 'test-support' } },
              allow: { to: { element: { type: ['test-support', 'shared'] } } },
            },
            {
              from: { element: { type: 'app' } },
              allow: { to: { element: { type: ['app', 'feature', 'platform', 'shared'] } } },
            },
            {
              from: { element: { type: 'feature' } },
              allow: {
                to: [
                  { element: { type: 'feature', captured: { feature: '{{ from.element.captured.feature }}' } } },
                  { element: { type: 'feature' }, file: { categories: 'store' } },
                  { element: { type: ['platform', 'shared'] } },
                ],
              },
            },
            {
              from: { element: { type: 'platform' } },
              allow: { to: { element: { type: ['platform', 'shared', 'bindings'] } } },
            },
            {
              from: { element: { type: 'shared' } },
              allow: { to: { element: { type: 'shared' } } },
            },
            {
              from: { element: { type: 'bindings' } },
              allow: { to: { element: { type: 'bindings' } } },
            },
          ],
        },
      ],
      'no-restricted-imports': [
        'error',
        {
          patterns: [
            { group: ['@wailsio/runtime', '@wailsio/runtime/*'], message: WAILS_RUNTIME_MESSAGE },
            { group: ['**/bindings/**'], message: WAILS_RUNTIME_MESSAGE },
          ],
        },
      ],
    },
  },
  {
    files: ['src/platform/**/*.{ts,tsx}'],
    rules: {
      'no-restricted-imports': 'off',
    },
  },
  {
    files: ['src/**/index.{ts,tsx}'],
    rules: {
      'no-restricted-syntax': [
        'error',
        { selector: 'Program', message: 'No barrel files: import the module file directly.' },
      ],
    },
  },
  {
    files: ['vite.config.ts', 'eslint.config.js'],
    languageOptions: {
      globals: globals.node,
    },
    rules: {
      'no-restricted-imports': 'off',
    },
  },
)
