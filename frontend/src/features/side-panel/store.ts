import { create } from 'zustand'
import { useRequestBuilderStore } from '@/features/request-builder/store'
import { useWorkspacesStore } from '@/features/workspaces/store'
import { toAppError, type AppError } from '@/platform/errors'
import type { Snippet, SnippetLanguage } from '@/platform/mappers/snippet'
import { copySnippet, generateSnippet, listSnippetLanguages, type SnippetInput } from '@/platform/services/snippets'
import type { CodeSnippetState, SnippetRequest } from './model'

interface SidePanelState {
  languages: SnippetLanguage[]
  languagesError: AppError | null
  includeSecrets: boolean
  snippets: Record<string, CodeSnippetState>
  loadLanguages: () => Promise<void>
  setIncludeSecrets: (includeSecrets: boolean) => void
  /** Generates the tab's draft as code in Go; only the latest answer per tab is kept. */
  generate: (tabId: string, request: SnippetRequest) => Promise<void>
  /** Go copies the same text it generates, so masked secrets stay masked. */
  copy: (tabId: string, request: SnippetRequest) => Promise<AppError | null>
  reset: () => void
}

const initialState = {
  languages: [] as SnippetLanguage[],
  languagesError: null as AppError | null,
  includeSecrets: false,
  snippets: {} as Record<string, CodeSnippetState>,
}

const generations = new Map<string, number>()

function inputFor(tabId: string, request: SnippetRequest, includeSecrets: boolean): SnippetInput | null {
  const localId = useWorkspacesStore.getState().activeLocalId
  const draft = useRequestBuilderStore.getState().tabs[tabId]?.draft
  if (!localId || !draft) {
    return null
  }
  return { localId, tabId, draft, language: request.language, options: { ...request.options }, includeSecrets }
}

export const useSidePanelStore = create<SidePanelState>()((set, get) => {
  function patch(tabId: string, update: Partial<CodeSnippetState>) {
    set((state) => ({ snippets: { ...state.snippets, [tabId]: { snippet: null, status: 'idle', error: null, ...state.snippets[tabId], ...update } } }))
  }

  return {
    ...initialState,

    loadLanguages: async () => {
      if (get().languages.length > 0) {
        return
      }
      try {
        set({ languages: await listSnippetLanguages(), languagesError: null })
      } catch (error) {
        set({ languagesError: toAppError(error) })
      }
    },

    setIncludeSecrets: (includeSecrets) => set({ includeSecrets }),

    generate: async (tabId, request) => {
      const input = inputFor(tabId, request, get().includeSecrets)
      if (!input) {
        return
      }
      const generation = (generations.get(tabId) ?? 0) + 1
      generations.set(tabId, generation)
      patch(tabId, { status: 'loading' })
      try {
        const snippet: Snippet = await generateSnippet(input)
        if (generations.get(tabId) === generation) {
          patch(tabId, { snippet, status: 'ready', error: null })
        }
      } catch (error) {
        if (generations.get(tabId) === generation) {
          patch(tabId, { status: 'error', error: toAppError(error) })
        }
      }
    },

    copy: async (tabId, request) => {
      const input = inputFor(tabId, request, get().includeSecrets)
      if (!input) {
        return null
      }
      try {
        await copySnippet(input)
        return null
      } catch (error) {
        return toAppError(error)
      }
    },

    reset: () => {
      generations.clear()
      set(initialState)
    },
  }
})
