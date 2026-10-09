import { create } from 'zustand'
import { useWorkspacesStore } from '@/features/workspaces/store'
import { toAppError, type AppError } from '@/platform/errors'
import type { HistoryEntry } from '@/platform/mappers/history'
import type { KeyValueRow, RequestDraft } from '@/platform/mappers/request'
import type { ItemSettings, TreeNode, TreeSearchResult } from '@/platform/mappers/tree'
import type { SidebarView } from '@/platform/mappers/ui-state'
import { clearHistory, getHistoryEntry, listHistory } from '@/platform/services/history'
import { formatBulk as formatBulkRows, parseBulk as parseBulkText } from '@/platform/services/requests'
import {
  createCollection,
  createFolder,
  createRequest,
  deleteItem,
  duplicateItem,
  getItemSettings,
  getTree,
  renameItem,
  saveItemSettings,
  searchTree,
} from '@/platform/services/tree'
import { findPath } from './lib/tree-rows'

export const SEARCH_DELAY_MS = 100
export const NEW_COLLECTION_NAME = 'New Collection'
export const NEW_FOLDER_NAME = 'New Folder'
export const NEW_REQUEST_NAME = 'New Request'

export type SidebarStatus = 'idle' | 'loading' | 'ready'

export interface SidebarSnapshot {
  expanded: Record<string, boolean>
  selectedId: string | null
  query: string
  view: SidebarView
}

interface SidebarState {
  view: SidebarView
  history: HistoryEntry[]
  historyStatus: SidebarStatus
  historyQuery: string
  tree: TreeNode[]
  status: SidebarStatus
  error: AppError | null
  expanded: Record<string, boolean>
  selectedId: string | null
  renamingId: string | null
  query: string
  search: TreeSearchResult | null
  load: () => Promise<void>
  setTree: (tree: TreeNode[]) => void
  applyState: (state: SidebarSnapshot) => void
  setView: (view: SidebarView) => void
  loadHistory: () => Promise<void>
  setHistoryQuery: (query: string) => void
  clearHistory: () => Promise<void>
  openHistoryEntry: (id: string) => Promise<RequestDraft | null>
  setQuery: (query: string) => void
  toggle: (id: string) => void
  select: (id: string) => void
  createCollection: () => Promise<void>
  createFolder: (parentId: string) => Promise<void>
  createRequest: (parentId: string) => Promise<void>
  startRename: (id: string) => void
  cancelRename: () => void
  rename: (id: string, name: string) => Promise<boolean>
  duplicate: (id: string) => Promise<void>
  remove: (id: string) => Promise<string[]>
  loadSettings: (id: string) => Promise<ItemSettings | null>
  saveSettings: (settings: ItemSettings) => Promise<boolean>
  formatBulk: (rows: KeyValueRow[]) => Promise<string | null>
  parseBulk: (text: string) => Promise<KeyValueRow[] | null>
  clearError: () => void
  reset: () => void
}

let searchTimer: ReturnType<typeof setTimeout> | undefined
let historyTimer: ReturnType<typeof setTimeout> | undefined

function activeLocalId(): string | null {
  return useWorkspacesStore.getState().activeLocalId
}

const initialState = {
  view: 'collections' as SidebarView,
  history: [] as HistoryEntry[],
  historyStatus: 'idle' as SidebarStatus,
  historyQuery: '',
  tree: [] as TreeNode[],
  status: 'idle' as SidebarStatus,
  error: null as AppError | null,
  expanded: {} as Record<string, boolean>,
  selectedId: null as string | null,
  renamingId: null as string | null,
  query: '',
  search: null as TreeSearchResult | null,
}

export const useSidebarStore = create<SidebarState>()((set, get) => {
  async function attempt<T>(operation: () => Promise<T>): Promise<T | null> {
    try {
      return await operation()
    } catch (error) {
      set({ error: toAppError(error) })
      return null
    }
  }

  async function run<T>(operation: (localId: string) => Promise<T>): Promise<T | null> {
    const localId = activeLocalId()
    if (!localId) {
      return null
    }
    try {
      return await operation(localId)
    } catch (error) {
      set({ error: toAppError(error) })
      return null
    }
  }

  async function refresh(localId: string) {
    const tree = await getTree(localId)
    if (activeLocalId() === localId) {
      set({ tree, status: 'ready' })
    }
  }

  async function create(parentId: string | null, operation: (localId: string) => Promise<TreeNode>) {
    await run(async (localId) => {
      const created = await operation(localId)
      await refresh(localId)
      set((state) => ({
        expanded: parentId ? { ...state.expanded, [parentId]: true } : state.expanded,
        selectedId: created.id,
        renamingId: created.id,
      }))
    })
  }

  return {
    ...initialState,

    load: async () => {
      set({ status: 'loading', error: null })
      await run(refresh)
      if (get().status === 'loading') {
        set({ status: 'ready' })
      }
    },

    setTree: (tree) => set({ tree, status: 'ready', error: null }),

    applyState: ({ expanded, selectedId, query, view }) => {
      set({ expanded, selectedId })
      get().setQuery(query)
      get().setView(view)
    },

    setView: (view) => {
      set({ view })
      if (view === 'history') {
        void get().loadHistory()
      }
    },

    loadHistory: async () => {
      const query = get().historyQuery
      set((state) => ({ historyStatus: state.historyStatus === 'ready' ? 'ready' : 'loading' }))
      await run(async (localId) => {
        const history = await listHistory(localId, { query })
        if (activeLocalId() === localId && get().historyQuery === query) {
          set({ history, historyStatus: 'ready' })
        }
      })
      if (get().historyStatus === 'loading') {
        set({ historyStatus: 'ready' })
      }
    },

    setHistoryQuery: (query) => {
      clearTimeout(historyTimer)
      set({ historyQuery: query })
      historyTimer = setTimeout(() => void get().loadHistory(), SEARCH_DELAY_MS)
    },

    clearHistory: async () => {
      await run(async (localId) => {
        await clearHistory(localId)
        set({ history: [] })
      })
    },

    openHistoryEntry: (id) =>
      run(async (localId) => {
        const detail = await getHistoryEntry(localId, id)
        return { ...detail.request, id: '' }
      }),

    setQuery: (query) => {
      clearTimeout(searchTimer)
      set({ query })
      if (query.trim() === '') {
        set({ search: null })
        return
      }
      searchTimer = setTimeout(() => {
        void run(async (localId) => {
          const search = await searchTree(localId, query)
          if (get().query === query) {
            set({ search })
          }
        })
      }, SEARCH_DELAY_MS)
    },

    toggle: (id) => set((state) => ({ expanded: { ...state.expanded, [id]: !state.expanded[id] } })),

    select: (id) => set({ selectedId: id }),

    createCollection: () => create(null, (localId) => createCollection(localId, NEW_COLLECTION_NAME)),

    createFolder: (parentId) => create(parentId, (localId) => createFolder(localId, parentId, NEW_FOLDER_NAME)),

    createRequest: (parentId) => create(parentId, (localId) => createRequest(localId, parentId, NEW_REQUEST_NAME)),

    startRename: (id) => set({ renamingId: id }),

    cancelRename: () => set({ renamingId: null }),

    rename: async (id, name) => {
      set({ renamingId: null })
      const current = findPath(get().tree, id).at(-1)
      if (name.trim() === '' || current?.name === name) {
        return false
      }
      const renamed = await run(async (localId) => {
        await renameItem(localId, id, name)
        await refresh(localId)
        return true
      })
      return renamed === true
    },

    duplicate: async (id) => {
      await run(async (localId) => {
        const copy = await duplicateItem(localId, id)
        await refresh(localId)
        set({ selectedId: copy.id })
      })
    },

    remove: async (id) => {
      const deleted = await run(async (localId) => {
        const ids = await deleteItem(localId, id)
        await refresh(localId)
        return ids
      })
      if (get().selectedId === id) {
        set({ selectedId: null })
      }
      return deleted ?? []
    },

    loadSettings: (id) => run((localId) => getItemSettings(localId, id)),

    saveSettings: async (settings) => {
      const saved = await run(async (localId) => {
        await saveItemSettings(localId, settings)
        return true
      })
      return saved === true
    },

    formatBulk: (rows) => attempt(() => formatBulkRows(rows)),

    parseBulk: (text) => attempt(() => parseBulkText(text)),

    clearError: () => set({ error: null }),

    reset: () => {
      clearTimeout(searchTimer)
      clearTimeout(historyTimer)
      set(initialState)
    },
  }
})

export function selectLocation(tree: TreeNode[], id: string): string[] {
  return findPath(tree, id).map((node) => node.name)
}
