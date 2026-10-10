import { create } from 'zustand'
import { useWorkspacesStore } from '@/features/workspaces/store'
import { toAppError, type AppError } from '@/platform/errors'
import type { HistoryEntry } from '@/platform/mappers/history'
import type { KeyValueRow, RequestDraft } from '@/platform/mappers/request'
import type { ItemSettings, TreeNode, TreeSearchResult } from '@/platform/mappers/tree'
import type { SidebarView } from '@/platform/mappers/ui-state'
import type { Result } from '@/shared/types/result'
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
import type { SidebarSnapshot, SidebarStatus } from './model'

export const SEARCH_DELAY_MS = 100
export const NEW_COLLECTION_NAME = 'New Collection'
export const NEW_FOLDER_NAME = 'New Folder'
export const NEW_REQUEST_NAME = 'New Request'

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
  clearHistory: () => Promise<AppError | null>
  openHistoryEntry: (id: string) => Promise<RequestDraft | null>
  setQuery: (query: string) => void
  toggle: (id: string) => void
  select: (id: string) => void
  /** Expands the item's ancestors, selects it and returns it; null when the tree has no such item. */
  reveal: (id: string) => TreeNode | null
  createCollection: () => Promise<void>
  createFolder: (parentId: string) => Promise<void>
  createRequest: (parentId: string) => Promise<void>
  startRename: (id: string) => void
  cancelRename: () => void
  rename: (id: string, name: string) => Promise<Result<boolean, AppError>>
  duplicate: (id: string) => Promise<void>
  remove: (id: string) => Promise<Result<string[], AppError>>
  loadSettings: (id: string) => Promise<Result<ItemSettings | null, AppError>>
  saveSettings: (settings: ItemSettings) => Promise<AppError | null>
  formatBulk: (rows: KeyValueRow[]) => Promise<Result<string, AppError>>
  parseBulk: (text: string) => Promise<Result<KeyValueRow[], AppError>>
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
  async function attempt<T>(operation: () => Promise<T>): Promise<Result<T, AppError>> {
    try {
      return { ok: true, value: await operation() }
    } catch (error) {
      return { ok: false, error: toAppError(error) }
    }
  }

  function settle<T>(operation: (localId: string) => Promise<T>): Promise<Result<T | null, AppError>> {
    const localId = activeLocalId()
    return localId ? attempt(() => operation(localId)) : Promise.resolve<Result<T | null, AppError>>({ ok: true, value: null })
  }

  function endRename(id: string) {
    set((state) => (state.renamingId === id ? { renamingId: null } : state))
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
      const cleared = await settle(async (localId) => {
        await clearHistory(localId)
        set({ history: [] })
      })
      return cleared.ok ? null : cleared.error
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

    reveal: (id) => {
      const path = findPath(get().tree, id)
      const node = path.at(-1) ?? null
      if (node) {
        const ancestors = Object.fromEntries(path.slice(0, -1).map((ancestor) => [ancestor.id, true]))
        set((state) => ({ expanded: { ...state.expanded, ...ancestors }, selectedId: id }))
      }
      return node
    },

    createCollection: () => create(null, (localId) => createCollection(localId, NEW_COLLECTION_NAME)),

    createFolder: (parentId) => create(parentId, (localId) => createFolder(localId, parentId, NEW_FOLDER_NAME)),

    createRequest: (parentId) => create(parentId, (localId) => createRequest(localId, parentId, NEW_REQUEST_NAME)),

    startRename: (id) => set({ renamingId: id }),

    cancelRename: () => set({ renamingId: null }),

    rename: async (id, name) => {
      const current = findPath(get().tree, id).at(-1)
      if (name.trim() === '' || current?.name === name) {
        endRename(id)
        return { ok: true, value: false }
      }
      const renamed = await settle(async (localId) => {
        await renameItem(localId, id, name)
        await refresh(localId)
        return true
      })
      if (!renamed.ok) {
        return renamed
      }
      endRename(id)
      return { ok: true, value: renamed.value === true }
    },

    duplicate: async (id) => {
      await run(async (localId) => {
        const copy = await duplicateItem(localId, id)
        await refresh(localId)
        set({ selectedId: copy.id })
      })
    },

    remove: async (id) => {
      const deleted = await settle(async (localId) => {
        const ids = await deleteItem(localId, id)
        await refresh(localId)
        return ids
      })
      if (!deleted.ok) {
        return deleted
      }
      if (get().selectedId === id) {
        set({ selectedId: null })
      }
      return { ok: true, value: deleted.value ?? [] }
    },

    loadSettings: (id) => settle((localId) => getItemSettings(localId, id)),

    saveSettings: async (settings) => {
      const saved = await settle((localId) => saveItemSettings(localId, settings))
      return saved.ok ? null : saved.error
    },

    formatBulk: (rows) => attempt(() => formatBulkRows(rows)),

    parseBulk: (text) => attempt(() => parseBulkText(text)),

    clearError: () => set({ error: null }),

    reset: () => {
      clearTimeout(searchTimer)
      clearTimeout(historyTimer)
      searchTimer = undefined
      historyTimer = undefined
      set(initialState)
    },
  }
})

export function selectLocation(tree: TreeNode[], id: string): string[] {
  return findPath(tree, id).map((node) => node.name)
}
