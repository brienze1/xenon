import '@testing-library/jest-dom/vitest'
import { cleanup } from '@testing-library/react'
import { afterAll, afterEach } from 'vitest'

// The Wails runtime (loaded through the generated bindings) polls `window` on a 50 ms interval
// right after it is imported. Let that first tick run before the test environment is torn down,
// otherwise a fast test file ends with an unhandled "window is not defined" error.
const WAILS_RUNTIME_POLL_MS = 60

afterEach(() => {
  cleanup()
})

afterAll(() => new Promise<void>((resolve) => setTimeout(resolve, WAILS_RUNTIME_POLL_MS)))
