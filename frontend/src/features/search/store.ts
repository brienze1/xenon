import { create } from 'zustand'
import { useTabsStore } from '@/features/tabs/store'
import { useWorkspacesStore } from '@/features/workspaces/store'
import { toAppError, type AppError } from '@/platform/errors'
import type { QuickSearchHit } from '@/platform/mappers/tree'
import { quickSearch } from '@/platform/services/tree'
import type { QuickSearchStatus } from './model'

export const QUICK_SEARCH_DELAY_MS = 80
export const QUICK_SEARCH_LIMIT = 50

interface SearchState {
  open: boolean
  query: string
  results: QuickSearchHit[]
  status: QuickSearchStatus
  error: AppError | null
  activeIndex: number
  openSearch: () => void
  close: () => void
  setQuery: (query: string) => void
  /** Moves the highlighted result, wrapping around. */
  move: (step: 1 | -1) => void
  /** Opens the highlighted (or given) result in a tab and closes the search. */
  choose: (index?: number) => Promise<void>
  reset: () => void
}

const initialState = {
  open: false,
  query: '',
  results: [] as QuickSearchHit[],
  status: 'idle' as QuickSearchStatus,
  error: null as AppError | null,
  activeIndex: 0,
}

let searchTimer: ReturnType<typeof setTimeout> | undefined
let generation = 0

function cancelSearch(): void {
  clearTimeout(searchTimer)
  searchTimer = undefined
  generation++
}

export const useSearchStore = create<SearchState>()((set, get) => {
  async function run(query: string) {
    const localId = useWorkspacesStore.getState().activeLocalId
    const current = ++generation
    if (!localId || query.trim() === '') {
      set({ results: [], status: 'idle', error: null, activeIndex: 0 })
      return
    }
    try {
      const results = await quickSearch(localId, query, QUICK_SEARCH_LIMIT)
      if (current === generation) {
        set({ results, status: 'ready', error: null, activeIndex: 0 })
      }
    } catch (error) {
      if (current === generation) {
        set({ results: [], status: 'error', error: toAppError(error), activeIndex: 0 })
      }
    }
  }

  return {
    ...initialState,

    openSearch: () => {
      cancelSearch()
      set({ ...initialState, open: true })
    },

    close: () => {
      cancelSearch()
      set(initialState)
    },

    setQuery: (query) => {
      cancelSearch()
      set({ query, status: query.trim() === '' ? 'idle' : 'loading' })
      searchTimer = setTimeout(() => {
        searchTimer = undefined
        void run(query)
      }, QUICK_SEARCH_DELAY_MS)
    },

    move: (step) => {
      const count = get().results.length
      if (count > 0) {
        set((state) => ({ activeIndex: (state.activeIndex + step + count) % count }))
      }
    },

    choose: async (index) => {
      const hit = get().results[index ?? get().activeIndex]
      if (!hit) {
        return
      }
      get().close()
      await useTabsStore.getState().openPinned(hit.id)
    },

    reset: () => get().close(),
  }
})
