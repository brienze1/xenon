import { create } from 'zustand'
import { useWorkspacesStore } from '@/features/workspaces/store'
import { toAppError, type AppError } from '@/platform/errors'
import type { EnvironmentSummary } from '@/platform/mappers/environment'
import { listEnvironments, setActiveEnvironment } from '@/platform/services/environments'
import type { EnvironmentsStatus } from './model'

interface EnvironmentsState {
  environments: EnvironmentSummary[]
  activeId: string | null
  status: EnvironmentsStatus
  error: AppError | null
  switcherOpen: boolean
  load: () => Promise<void>
  setActive: (id: string | null) => Promise<void>
  openSwitcher: () => void
  setSwitcherOpen: (open: boolean) => void
  clearError: () => void
  reset: () => void
}

const INITIAL_STATE = {
  environments: [],
  activeId: null,
  status: 'idle',
  error: null,
  switcherOpen: false,
} satisfies Partial<EnvironmentsState>

function activeLocalId(): string | null {
  return useWorkspacesStore.getState().activeLocalId
}

export const useEnvironmentsStore = create<EnvironmentsState>()((set) => ({
  ...INITIAL_STATE,

  load: async () => {
    const localId = activeLocalId()
    if (localId === null) {
      return
    }
    set({ status: 'loading', error: null })
    try {
      const list = await listEnvironments(localId)
      if (activeLocalId() === localId) {
        set({ environments: list.environments, activeId: list.activeId, status: 'ready' })
      }
    } catch (error) {
      if (activeLocalId() === localId) {
        set({ status: 'ready', error: toAppError(error) })
      }
    }
  },

  setActive: async (id) => {
    const localId = activeLocalId()
    if (localId === null) {
      return
    }
    set({ error: null })
    try {
      const activeId = await setActiveEnvironment(localId, id)
      if (activeLocalId() === localId) {
        set({ activeId })
      }
    } catch (error) {
      set({ error: toAppError(error) })
    }
  },

  openSwitcher: () => set({ switcherOpen: true }),

  setSwitcherOpen: (open) => set({ switcherOpen: open }),

  clearError: () => set({ error: null }),

  reset: () => set(INITIAL_STATE),
}))
