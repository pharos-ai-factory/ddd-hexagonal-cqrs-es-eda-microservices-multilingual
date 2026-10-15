import type { RsbuildPlugin } from '@rsbuild/core';

const excludedDirectories =
  /(^|[/\\])(node_modules|\.git|\.venv|\.local|\.tools|\.rspress|\.mypy_cache|\.ruff_cache|\.pytest_cache|__pycache__|generated|dist|out|\.next|test-results|playwright-report)([/\\]|$)/;

/** Restrict Rspress's existing repository watcher after Rsbuild merges its watch list. */
export function documentationWatcher(): RsbuildPlugin {
  return {
    name: 'cafe-documentation-watcher',
    setup(api) {
      api.modifyRsbuildConfig((config) => {
        const configured = config.dev?.watchFiles;
        const watchers = configured ? (Array.isArray(configured) ? configured : [configured]) : [];
        for (const watcher of watchers) {
          if (watcher.type === 'restart') {
            watcher.options = { ...watcher.options, ignored: excludedDirectories };
          }
        }
        return config;
      });
    },
  };
}
