import { useState } from 'react'
import { useEnvironmentsStore } from '@/features/environments/store'

export interface DiscardGuard {
  /** Whether the "Discard unsaved variable changes?" confirmation is showing. */
  confirming: boolean
  /** Runs `proceed` now, or after confirmation when a table of `scopes` has unsaved edits. */
  guard: (proceed: () => void) => void
  discard: () => void
  keepEditing: () => void
}

/** Asks before an action (closing an editor, switching environment) drops unsaved variable table edits. */
export function useDiscardGuard(scopes: readonly string[]): DiscardGuard {
  const [pending, setPending] = useState<(() => void) | null>(null)

  function guard(proceed: () => void) {
    const dirty = useEnvironmentsStore.getState().dirtyScopes.some((scope) => scopes.includes(scope))
    if (dirty) {
      setPending(() => proceed)
      return
    }
    proceed()
  }

  function discard() {
    const proceed = pending
    setPending(null)
    proceed?.()
  }

  return { confirming: pending !== null, guard, discard, keepEditing: () => setPending(null) }
}
