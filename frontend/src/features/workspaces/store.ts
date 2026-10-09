import { create } from 'zustand'
import { toAppError, type AppError } from '@/platform/errors'
import type { WorkspaceSummary } from '@/platform/mappers/workspace'
import { pickFolder } from '@/platform/services/dialogs'
import { activateWorkspace, createWorkspace, listWorkspaces } from '@/platform/services/workspaces'

const FOLDER_MISSING = 'WKS-06404'

export type WorkspacesStatus = 'idle' | 'loading' | 'ready'

interface WorkspacesState {
  workspaces: WorkspaceSummary[]
  activeLocalId: string | null
  status: WorkspacesStatus
  error: AppError | null
  load: () => Promise<void>
  create: (name: string, path: string) => Promise<boolean>
  activate: (localId: string) => Promise<void>
  chooseFolder: (title: string) => Promise<string | null>
  markMissing: (localId: string) => void
  clearError: () => void
}

function replaceSummary(workspaces: WorkspaceSummary[], summary: WorkspaceSummary): WorkspaceSummary[] {
  return workspaces.map((workspace) => (workspace.localId === summary.localId ? summary : workspace))
}

export const useWorkspacesStore = create<WorkspacesState>()((set) => ({
  workspaces: [],
  activeLocalId: null,
  status: 'idle',
  error: null,

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
    set({ error: null })
    try {
      const summary = await createWorkspace({ name, path })
      set((state) => ({ workspaces: [...state.workspaces, summary], activeLocalId: summary.localId }))
      return true
    } catch (error) {
      set({ error: toAppError(error) })
      return false
    }
  },

  activate: async (localId) => {
    set({ error: null })
    try {
      const snapshot = await activateWorkspace(localId)
      set((state) => ({ workspaces: replaceSummary(state.workspaces, snapshot.summary), activeLocalId: localId }))
    } catch (error) {
      const appError = toAppError(error)
      set((state) => ({
        error: appError,
        activeLocalId: appError.code === FOLDER_MISSING && state.activeLocalId === localId ? null : state.activeLocalId,
        workspaces:
          appError.code === FOLDER_MISSING
            ? state.workspaces.map((workspace) => (workspace.localId === localId ? { ...workspace, missing: true } : workspace))
            : state.workspaces,
      }))
    }
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
    set((state) => ({
      workspaces: state.workspaces.map((workspace) => (workspace.localId === localId ? { ...workspace, missing: true } : workspace)),
    }))
  },

  clearError: () => set({ error: null }),
}))
