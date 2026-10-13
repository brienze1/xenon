import { useCallback, useSyncExternalStore } from 'react'

// A per-machine display preference, not product state (frontend-architecture: localStorage only for these).
const STORAGE_KEY = 'xenon.autoHeadersHidden'

const listeners = new Set<() => void>()

// Used only when the storage cannot be read, so the choice still applies until the app closes.
let unsaved = true

function readSaved(): boolean {
  try {
    return window.localStorage.getItem(STORAGE_KEY) !== 'false'
  } catch {
    return unsaved
  }
}

function subscribe(listener: () => void) {
  listeners.add(listener)
  return () => {
    listeners.delete(listener)
  }
}

function save(next: boolean) {
  unsaved = next
  try {
    window.localStorage.setItem(STORAGE_KEY, String(next))
  } catch {
    // Not remembered this time; the choice still applies until the app closes.
  }
}

/**
 * Whether the auto-generated headers are folded away in every request's Headers tab — hidden by default, as in
 * Postman. Display only: hiding never changes which auto headers are sent.
 */
export function useAutoHeadersHidden(): [boolean, () => void] {
  const value = useSyncExternalStore(subscribe, readSaved)
  const toggle = useCallback(() => {
    save(!readSaved())
    listeners.forEach((listener) => listener())
  }, [])
  return [value, toggle]
}
