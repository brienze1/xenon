import { create } from 'zustand'
import { useWorkspacesStore } from '@/features/workspaces/store'
import { toAppError, type AppError } from '@/platform/errors'
import type { HistoryActivity, HistoryEntry, HistorySourceFilter } from '@/platform/mappers/history'
import type { RequestDraft } from '@/platform/mappers/request'
import type { TreeNode, TreeSearchResult } from '@/platform/mappers/tree'
import type { SidebarView } from '@/platform/mappers/ui-state'
import type { Result } from '@/shared/types/result'
import { clearHistory, getHistoryEntry, listHistory } from '@/platform/services/history'
import {
  createCollection,
  createFolder,
  createRequest,
  deleteItem,
  duplicateItem,
  getTree,
  moveItem,
  renameItem,
  searchTree,
} from '@/platform/services/tree'
import { deleteExample, duplicateExample, renameExample } from '@/platform/services/examples'
import type { DropTarget, MoveTarget } from './lib/tree-drop'
import { findExampleRequest, findPath } from './lib/tree-rows'
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
  /** The History view's source filter (All / User / AI), remembered per workspace. */
  historySource: HistorySourceFilter
  tree: TreeNode[]
  status: SidebarStatus
  error: AppError | null
  expanded: Record<string, boolean>
  selectedId: string | null
  renamingId: string | null
  query: string
  search: TreeSearchResult | null
  /** The item being dragged in the tree. */
  draggingId: string | null
  /** Where the dragged item would drop now (the drop indicator). */
  dropTarget: DropTarget | null
  load: () => Promise<void>
  setTree: (tree: TreeNode[]) => void
  applyState: (state: SidebarSnapshot) => void
  setView: (view: SidebarView) => void
  loadHistory: () => Promise<void>
  setHistoryQuery: (query: string) => void
  setHistorySource: (source: HistorySourceFilter) => void
  /** Clears the entries the current source filter shows. */
  clearHistory: () => Promise<AppError | null>
  /** An AI action row's detail: what it did, with its masked arguments. */
  loadHistoryActivity: (id: string) => Promise<Result<HistoryActivity | null, AppError>>
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
  renameExample: (exampleId: string, name: string) => Promise<Result<boolean, AppError>>
  duplicateExample: (exampleId: string) => Promise<void>
  /** Starts renaming the selected item (Cmd/Ctrl+E). */
  renameSelected: () => void
  /** Duplicates the selected item or example (Cmd/Ctrl+D). */
  duplicateSelected: () => Promise<void>
  removeExample: (exampleId: string) => Promise<AppError | null>
  startDrag: (id: string) => void
  setDropTarget: (target: DropTarget | null) => void
  endDrag: () => void
  /** Moves the item to the new parent and index and shows the moved item. */
  move: (id: string, target: MoveTarget) => Promise<void>
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
  historySource: 'all' as HistorySourceFilter,
  tree: [] as TreeNode[],
  status: 'idle' as SidebarStatus,
  error: null as AppError | null,
  expanded: {} as Record<string, boolean>,
  selectedId: null as string | null,
  renamingId: null as string | null,
  query: '',
  search: null as TreeSearchResult | null,
  draggingId: null as string | null,
  dropTarget: null as DropTarget | null,
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

    applyState: ({ expanded, selectedId, query, view, historySource }) => {
      set({ expanded, selectedId, historySource })
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
      const { historyQuery: query, historySource: source } = get()
      set((state) => ({ historyStatus: state.historyStatus === 'ready' ? 'ready' : 'loading' }))
      await run(async (localId) => {
        const history = await listHistory(localId, { query, source })
        if (activeLocalId() === localId && get().historyQuery === query && get().historySource === source) {
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

    setHistorySource: (source) => {
      if (source === get().historySource) {
        return
      }
      set({ historySource: source })
      void get().loadHistory()
    },

    clearHistory: async () => {
      const source = get().historySource
      const cleared = await settle(async (localId) => {
        await clearHistory(localId, source)
        if (get().historySource === source) {
          set({ history: [] })
        }
      })
      return cleared.ok ? null : cleared.error
    },

    loadHistoryActivity: (id) => settle(async (localId) => (await getHistoryEntry(localId, id)).activity),

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

    renameExample: async (exampleId, name) => {
      const request = findExampleRequest(get().tree, exampleId)
      const current = request?.examples?.find((example) => example.id === exampleId)
      if (!request || name.trim() === '' || current?.name === name) {
        endRename(exampleId)
        return { ok: true, value: false }
      }
      const renamed = await settle(async (localId) => {
        await renameExample(localId, request.id, exampleId, name)
        await refresh(localId)
        return true
      })
      if (!renamed.ok) {
        return renamed
      }
      endRename(exampleId)
      return { ok: true, value: renamed.value === true }
    },

    renameSelected: () => {
      const { selectedId, tree } = get()
      if (selectedId !== null && (findPath(tree, selectedId).length > 0 || findExampleRequest(tree, selectedId))) {
        set({ renamingId: selectedId })
      }
    },

    duplicateSelected: async () => {
      const { selectedId, tree } = get()
      if (selectedId === null) {
        return
      }
      if (findPath(tree, selectedId).length > 0) {
        await get().duplicate(selectedId)
      } else if (findExampleRequest(tree, selectedId)) {
        await get().duplicateExample(selectedId)
      }
    },

    duplicateExample: async (exampleId) => {
      const request = findExampleRequest(get().tree, exampleId)
      if (!request) {
        return
      }
      await run(async (localId) => {
        const copy = await duplicateExample(localId, request.id, exampleId)
        await refresh(localId)
        set({ selectedId: copy.id })
      })
    },

    removeExample: async (exampleId) => {
      const request = findExampleRequest(get().tree, exampleId)
      if (!request) {
        return null
      }
      const removed = await settle(async (localId) => {
        await deleteExample(localId, request.id, exampleId)
        await refresh(localId)
      })
      if (!removed.ok) {
        return removed.error
      }
      if (get().selectedId === exampleId) {
        set({ selectedId: null })
      }
      return null
    },

    startDrag: (id) => set({ draggingId: id, dropTarget: null }),

    setDropTarget: (target) =>
      set((state) =>
        state.dropTarget?.id === target?.id && state.dropTarget?.position === target?.position ? state : { dropTarget: target },
      ),

    endDrag: () => set({ draggingId: null, dropTarget: null }),

    move: async (id, target) => {
      await run(async (localId) => {
        const tree = await moveItem(localId, id, target.parentId, target.index)
        if (activeLocalId() !== localId) {
          return
        }
        set((state) => ({
          tree,
          selectedId: id,
          expanded: target.parentId ? { ...state.expanded, [target.parentId]: true } : state.expanded,
        }))
      })
    },

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
