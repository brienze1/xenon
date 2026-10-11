import { create } from 'zustand'
import { toAppError, type AppError } from '@/platform/errors'
import type { ConsoleEntriesEvent, ConsoleEntry, ConsoleFilter } from '@/platform/mappers/console'
import { clearConsole, getConsoleEntry, listConsole } from '@/platform/services/console'
import { appendEntries, CONSOLE_CAPACITY } from './lib/filter'

type LoadStatus = 'idle' | 'loading' | 'ready'

interface ConsoleState {
  entries: ConsoleEntry[]
  status: LoadStatus
  filter: ConsoleFilter
  search: string
  selectedId: string | null
  detail: ConsoleEntry | null
  failure: AppError | null
  /** Loads the session console once; later entries arrive as events. */
  load: () => Promise<void>
  applyEvent: (event: ConsoleEntriesEvent) => void
  setFilter: (filter: ConsoleFilter) => void
  setSearch: (search: string) => void
  select: (id: string | null) => Promise<void>
  clear: () => Promise<void>
}

export const useConsoleStore = create<ConsoleState>()((set, get) => ({
  entries: [],
  status: 'idle',
  filter: 'all',
  search: '',
  selectedId: null,
  detail: null,
  failure: null,

  load: async () => {
    if (get().status !== 'idle') {
      return
    }
    set({ status: 'loading' })
    try {
      const listed = await listConsole({ limit: CONSOLE_CAPACITY })
      set((state) => ({ entries: appendEntries(listed, state.entries), status: 'ready' }))
    } catch (error) {
      set({ status: 'idle', failure: toAppError(error) })
    }
  },

  applyEvent: (event) =>
    set((state) => {
      const base = event.cleared ? [] : state.entries
      const entries = appendEntries(base, event.entries)
      const kept = state.selectedId !== null && entries.some((entry) => entry.id === state.selectedId)
      return { entries, ...(kept ? {} : { selectedId: null, detail: null }) }
    }),

  setFilter: (filter) => set({ filter }),

  setSearch: (search) => set({ search }),

  select: async (id) => {
    set({ selectedId: id, detail: null, failure: null })
    if (id === null) {
      return
    }
    try {
      const detail = await getConsoleEntry(id)
      if (get().selectedId === id) {
        set({ detail })
      }
    } catch (error) {
      if (get().selectedId === id) {
        set({ failure: toAppError(error) })
      }
    }
  },

  clear: async () => {
    set({ entries: [], selectedId: null, detail: null, failure: null })
    try {
      await clearConsole()
    } catch (error) {
      set({ failure: toAppError(error) })
    }
  },
}))
