import { create } from 'zustand'
import type { ScriptMember, ScriptSnippet, ScriptSyntaxError } from '@/platform/mappers/scripts'
import { checkScript, scriptReference, scriptSnippets } from '@/platform/services/scripts'

type LoadStatus = 'idle' | 'loading' | 'ready' | 'error'

interface ScriptsState {
  /** The sandbox surface for autocomplete, from Go's generated reference. */
  reference: ScriptMember[]
  snippets: ScriptSnippet[]
  status: LoadStatus
  load: () => Promise<void>
  /** Syntax errors of a script; none when Go cannot be asked. */
  check: (text: string) => Promise<ScriptSyntaxError[]>
}

export const useScriptsStore = create<ScriptsState>()((set, get) => ({
  reference: [],
  snippets: [],
  status: 'idle',

  load: async () => {
    if (get().status === 'loading' || get().status === 'ready') {
      return
    }
    set({ status: 'loading' })
    try {
      const [reference, snippets] = await Promise.all([scriptReference(), scriptSnippets()])
      set({ reference, snippets, status: 'ready' })
    } catch {
      set({ status: 'error' })
    }
  },

  check: async (text) => {
    try {
      return await checkScript(text)
    } catch {
      return []
    }
  },
}))
