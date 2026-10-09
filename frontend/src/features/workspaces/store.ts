import { create } from 'zustand'
import { useRequestBuilderStore } from '@/features/request-builder/store'
import { useResponseViewerStore } from '@/features/response-viewer/store'
import { useSidebarStore } from '@/features/sidebar/store'
import { useTabsStore } from '@/features/tabs/store'
import { toAppError, type AppError } from '@/platform/errors'
import { defaultLayout, type LayoutState } from '@/platform/mappers/ui-state'
import type { WorkspaceSummary } from '@/platform/mappers/workspace'
import { pickFolder } from '@/platform/services/dialogs'
import {
  activateWorkspace,
  createWorkspace,
  listWorkspaces,
  locateWorkspace,
  openWorkspace,
  removeWorkspace,
  renameWorkspace,
} from '@/platform/services/workspaces'
import { clampLayout } from './lib/layout'
import { replaceSummary, withMissing } from './lib/summaries'
import type { WorkspacesStatus } from './model'

export const FOLDER_MISSING = 'WKS-06404'
export const TRASH_FAILED = 'WKS-05500'

interface WorkspacesState {
  workspaces: WorkspaceSummary[]
  activeLocalId: string | null
  missingLocalId: string | null
  status: WorkspacesStatus
  error: AppError | null
  layout: LayoutState
  setLayout: (patch: Partial<LayoutState>) => void
  load: () => Promise<void>
  create: (name: string, path: string) => Promise<boolean>
  open: (path: string) => Promise<boolean>
  activate: (localId: string) => Promise<void>
  rename: (localId: string, name: string) => Promise<boolean>
  remove: (localId: string, moveToTrash: boolean) => Promise<boolean>
  locate: (localId: string) => Promise<void>
  chooseFolder: (title: string) => Promise<string | null>
  markMissing: (localId: string) => void
  clearError: () => void
}

function resetWorkspaceStores(): void {
  useTabsStore.getState().reset()
  useRequestBuilderStore.getState().reset()
  useResponseViewerStore.getState().reset()
  useSidebarStore.getState().reset()
}

export const useWorkspacesStore = create<WorkspacesState>()((set, get) => {
  async function attempt<T>(operation: () => Promise<T>): Promise<T | null> {
    set({ error: null })
    try {
      return await operation()
    } catch (error) {
      set({ error: toAppError(error) })
      return null
    }
  }

  function leaveWorkspace(missingLocalId: string | null) {
    resetWorkspaceStores()
    set({ activeLocalId: null, missingLocalId, layout: defaultLayout() })
  }

  return {
    workspaces: [],
    activeLocalId: null,
    missingLocalId: null,
    status: 'idle',
    error: null,
    layout: defaultLayout(),

    setLayout: (patch) => set((state) => ({ layout: clampLayout({ ...state.layout, ...patch }, window.innerHeight) })),

    load: async () => {
      set({ status: 'loading', error: null })
      try {
        const list = await listWorkspaces()
        set({ workspaces: list.workspaces, activeLocalId: list.activeLocalId, status: 'ready' })
      } catch (error) {
        set({ status: 'ready', error: toAppError(error) })
      }
    },

    create: async (name, path) => {
      const summary = await attempt(() => createWorkspace({ name, path }))
      if (summary === null) {
        return false
      }
      set((state) => ({ workspaces: [...state.workspaces, summary] }))
      await get().activate(summary.localId)
      return true
    },

    open: async (path) => {
      const summary = await attempt(() => openWorkspace(path))
      if (summary === null) {
        return false
      }
      set((state) => ({ workspaces: [...state.workspaces, summary] }))
      await get().activate(summary.localId)
      return true
    },

    activate: async (localId) => {
      set({ error: null })
      try {
        const snapshot = await activateWorkspace(localId)
        resetWorkspaceStores()
        set((state) => ({
          workspaces: replaceSummary(state.workspaces, snapshot.summary),
          activeLocalId: localId,
          missingLocalId: null,
          layout: clampLayout(snapshot.uiState.layout, window.innerHeight),
        }))
        useSidebarStore.getState().setTree(snapshot.tree)
        useTabsStore.getState().hydrate(localId, snapshot.uiState)
      } catch (error) {
        const appError = toAppError(error)
        if (appError.code !== FOLDER_MISSING) {
          set({ error: appError })
          return
        }
        set((state) => ({ workspaces: withMissing(state.workspaces, localId) }))
        leaveWorkspace(localId)
      }
    },

    rename: async (localId, name) => {
      const summary = await attempt(() => renameWorkspace(localId, name))
      if (summary === null) {
        return false
      }
      set((state) => ({ workspaces: replaceSummary(state.workspaces, summary) }))
      return true
    },

    remove: async (localId, moveToTrash) => {
      const wasActive = get().activeLocalId === localId
      const list = await attempt(() => removeWorkspace(localId, moveToTrash))
      if (list === null) {
        return false
      }
      set((state) => ({
        workspaces: list.workspaces,
        missingLocalId: state.missingLocalId === localId ? null : state.missingLocalId,
      }))
      if (!wasActive && get().activeLocalId !== null) {
        return true
      }
      if (list.activeLocalId === null) {
        leaveWorkspace(null)
        return true
      }
      await get().activate(list.activeLocalId)
      return true
    },

    locate: async (localId) => {
      const path = await get().chooseFolder('Locate the workspace folder')
      if (path === null) {
        return
      }
      const summary = await attempt(() => locateWorkspace(localId, path))
      if (summary === null) {
        return
      }
      set((state) => ({ workspaces: replaceSummary(state.workspaces, summary) }))
      await get().activate(localId)
    },

    chooseFolder: async (title) => {
      try {
        return await pickFolder(title)
      } catch (error) {
        set({ error: toAppError(error) })
        return null
      }
    },

    markMissing: (localId) => {
      set((state) => ({ workspaces: withMissing(state.workspaces, localId) }))
      if (get().activeLocalId === localId) {
        leaveWorkspace(localId)
      }
    },

    clearError: () => set({ error: null }),
  }
})
