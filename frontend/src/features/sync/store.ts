import { create } from 'zustand'
import { toAppError, type AppError } from '@/platform/errors'
import type { GitProgress } from '@/platform/events'
import type { LinkInput, SyncChange, SyncConflict, SyncStatus } from '@/platform/mappers/sync'
import {
  commitChanges,
  getSyncStatus,
  linkRepository,
  listChanges,
  pullWorkspace,
  pushWorkspace,
  saveSyncSettings,
} from '@/platform/services/sync'
import { useWorkspacesStore } from '@/features/workspaces/store'

export type SyncOperation = 'status' | 'link' | 'commit' | 'push' | 'pull' | 'settings'

export const STATUS_REFRESH_MS = 5 * 60 * 1000

interface SyncStoreState {
  localId: string | null
  status: SyncStatus | null
  changes: SyncChange[]
  excluded: string[]
  message: string
  conflicts: SyncConflict[]
  busy: SyncOperation | null
  progress: GitProgress | null
  error: AppError | null
  panelOpen: boolean
  linkOpen: boolean
  reset: (localId: string | null) => void
  refresh: () => Promise<void>
  loadChanges: () => Promise<void>
  toggle: (itemId: string) => void
  setMessage: (message: string) => void
  commit: (andPush: boolean) => Promise<void>
  push: () => Promise<void>
  pull: () => Promise<void>
  link: (input: Omit<LinkInput, 'localId'>) => Promise<AppError | null>
  saveSettings: (accountId: string) => Promise<void>
  applyStatus: (localId: string, status: SyncStatus) => void
  applyProgress: (progress: GitProgress) => void
  openPanel: () => void
  closePanel: () => void
  openLink: () => void
  closeLink: () => void
}

const EMPTY = {
  status: null,
  changes: [],
  excluded: [],
  message: '',
  conflicts: [],
  busy: null,
  progress: null,
  error: null,
} satisfies Partial<SyncStoreState>

export const useSyncStore = create<SyncStoreState>()((set, get) => {
  async function run(operation: SyncOperation, work: (localId: string) => Promise<void>): Promise<AppError | null> {
    const localId = get().localId
    if (localId === null || get().busy !== null) {
      return null
    }
    set({ busy: operation, error: null })
    try {
      await work(localId)
      return null
    } catch (error) {
      const failure = toAppError(error)
      if (get().localId === localId) {
        set({ error: failure })
      }
      return failure
    } finally {
      if (get().localId === localId) {
        set({ busy: null, progress: null })
      }
    }
  }

  function apply(localId: string, status: SyncStatus): void {
    if (get().localId === localId) {
      set({ status })
    }
  }

  async function reloadChanges(localId: string): Promise<void> {
    const changes = await listChanges(localId)
    if (get().localId === localId) {
      set({ changes })
    }
  }

  return {
    localId: null,
    ...EMPTY,
    panelOpen: false,
    linkOpen: false,

    reset: (localId) => set({ localId, ...EMPTY, panelOpen: false, linkOpen: false }),

    refresh: async () => {
      await run('status', async (localId) => apply(localId, await getSyncStatus(localId)))
    },

    loadChanges: async () => {
      const localId = get().localId
      if (localId === null || get().status?.state === 'not_linked') {
        return
      }
      try {
        await reloadChanges(localId)
      } catch (error) {
        set({ error: toAppError(error) })
      }
    },

    toggle: (itemId) =>
      set((state) => ({
        excluded: state.excluded.includes(itemId) ? state.excluded.filter((id) => id !== itemId) : [...state.excluded, itemId],
      })),

    setMessage: (message) => set({ message }),

    commit: async (andPush) => {
      const { changes, excluded, message } = get()
      const itemIds = changes.filter((change) => !excluded.includes(change.itemId)).map((change) => change.itemId)
      await run(andPush ? 'push' : 'commit', async (localId) => {
        apply(localId, await commitChanges(localId, message, itemIds))
        set({ message: '', excluded: [] })
        if (andPush) {
          apply(localId, await pushWorkspace(localId))
        }
        await reloadChanges(localId)
      })
    },

    push: async () => {
      await run('push', async (localId) => apply(localId, await pushWorkspace(localId)))
    },

    pull: async () => {
      await run('pull', async (localId) => {
        const result = await pullWorkspace(localId)
        apply(localId, result.status)
        set({ conflicts: result.conflicts })
        await reloadChanges(localId)
      })
    },

    link: async (input) =>
      run('link', async (localId) => {
        apply(localId, await linkRepository({ ...input, localId }))
        set({ linkOpen: false, panelOpen: true })
        await useWorkspacesStore.getState().load()
        await useWorkspacesStore.getState().activate(localId)
        await reloadChanges(localId)
      }),

    saveSettings: async (accountId) => {
      const settings = get().status?.settings
      if (!settings) {
        return
      }
      await run('settings', async (localId) => {
        await saveSyncSettings(localId, accountId, settings.autoPull, settings.autoPush)
        apply(localId, await getSyncStatus(localId))
      })
    },

    applyStatus: (localId, status) => apply(localId, status),

    applyProgress: (progress) => {
      if (progress.localId === get().localId && get().busy !== null) {
        set({ progress })
      }
    },

    openPanel: () => {
      set({ panelOpen: true, error: null })
      void get().loadChanges()
    },
    closePanel: () => set({ panelOpen: false }),
    openLink: () => set({ linkOpen: true, error: null }),
    closeLink: () => set({ linkOpen: false }),
  }
})
