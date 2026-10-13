import '@testing-library/jest-dom/vitest'
import { cleanup, configure } from '@testing-library/react'
import { afterAll, afterEach } from 'vitest'

// The Wails runtime (loaded through the generated bindings) polls `window` on a 50 ms interval
// right after it is imported. Let that first tick run before the test environment is torn down,
// otherwise a fast test file ends with an unhandled "window is not defined" error.
const WAILS_RUNTIME_POLL_MS = 60

// The full validate pipeline runs these tests right after mutation testing saturates the CPU;
// the default 1 s wait for async UI made findBy*/waitFor flaky under that load.
configure({ asyncUtilTimeout: 5000 })

// jsdom has no layout: CodeMirror measures text ranges for tooltips (autocomplete), so give
// ranges empty geometry instead of failing.
if (typeof Range !== 'undefined' && !Range.prototype.getClientRects) {
  Range.prototype.getClientRects = () => [] as unknown as DOMRectList
  Range.prototype.getBoundingClientRect = () => new DOMRect()
}

afterEach(() => {
  cleanup()
})

afterAll(() => new Promise<void>((resolve) => setTimeout(resolve, WAILS_RUNTIME_POLL_MS)))
