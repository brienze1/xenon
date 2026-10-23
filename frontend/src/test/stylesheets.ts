import { readFileSync } from 'node:fs'

/**
 * Applies real stylesheets to jsdom so a test can read computed styles. Vitest does not inject CSS-module styles, and
 * a `?raw` import of a stylesheet is not a string under Vitest, so the files are read from disk. Paths are relative to
 * `frontend/`; pass them in the order the app loads them. `setup.ts` removes them after each test.
 */
export function injectStylesheets(...paths: string[]): void {
  for (const path of paths) {
    const style = document.createElement('style')
    style.dataset.testStylesheet = path
    style.textContent = readFileSync(path, 'utf8')
    document.head.append(style)
  }
}
