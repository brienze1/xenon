import { create } from 'zustand'
import { toAppError, type AppError } from '@/platform/errors'
import type { GitProgress } from '@/platform/events'
import type { ChosenResolution, CloneInput, LinkInput, SyncChange, SyncSession, SyncStatus } from '@/platform/mappers/sync'
import {
  cancelPull,
  cloneAsWorkspace,
  commitChanges,
  finishPull,
  getSyncSession,
  getSyncStatus,
  linkRepository,
  listChanges,
  pullWorkspace,
  pushWorkspace,
  resolveConflict,
  saveSyncSettings,
  unlinkRepository,
} from '@/platform/services/sync'
import { useWorkspacesStore } from '@/features/workspaces/store'

export type SyncOperation = 'status' | 'link' | 'unlink' | 'commit' | 'push' | 'pull' | 'settings' | 'conflicts' | 'resolve' | 'finish' | 'cancel'

export interface SyncSettingsChange {
  accountId?: string
  autoPull?: boolean
  autoPush?: boolean
}

export const STATUS_REFRESH_MS = 5 * 60 * 1000

interface SyncStoreState {
  localId: string | null
  status: SyncStatus | null
  changes: SyncChange[]
  excluded: string[]
  message: string
  session: SyncSession | null
  selectedConflictId: string | null
  busy: SyncOperation | null
  progress: GitProgress | null
  error: AppError | null
  panelOpen: boolean
  linkOpen: boolean
  conflictsOpen: boolean
  cloneOpen: boolean
  cloning: boolean
  reset: (localId: string | null) => void
  refresh: () => Promise<void>
  loadChanges: () => Promise<void>
  toggle: (itemId: string) => void
  setMessage: (message: string) => void
  commit: (andPush: boolean) => Promise<void>
  push: () => Promise<void>
  pull: () => Promise<void>
  link: (input: Omit<LinkInput, 'localId'>) => Promise<AppError | null>
  /** Clones a repository as a new workspace and switches to it; returns the failure, if any. */
  cloneWorkspace: (input: CloneInput) => Promise<AppError | null>
  saveSettings: (change: SyncSettingsChange) => Promise<void>
  /** Copies the workspace files to the folder, re-points the workspace there and deletes Xenon's clone. */
  unlink: (targetFolder: string) => Promise<AppError | null>
  applyStatus: (localId: string, status: SyncStatus) => void
  applyProgress: (progress: GitProgress) => void
  openPanel: () => void
  closePanel: () => void
  openLink: () => void
  closeLink: () => void
  openClone: () => void
  closeClone: () => void
  openConflicts: () => Promise<void>
  closeConflicts: () => void
  selectConflict: (itemId: string) => void
  resolve: (itemId: string, resolution: ChosenResolution, fieldPath?: string) => Promise<void>
  finishPull: () => Promise<void>
  cancelPull: () => Promise<void>
}

const EMPTY = {
  status: null,
  changes: [],
  excluded: [],
  message: '',
  session: null,
  selectedConflictId: null,
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

  function showSession(localId: string, session: SyncSession | null): void {
    if (get().localId !== localId) {
      return
    }
    const selected = session?.conflicts.find((conflict) => conflict.itemId === get().selectedConflictId) ?? session?.conflicts[0]
    set({ session, selectedConflictId: selected?.itemId ?? null, conflictsOpen: session !== null, panelOpen: session === null && get().panelOpen })
  }

  function closeSession(localId: string, status: SyncStatus): void {
    apply(localId, status)
    set({ session: null, selectedConflictId: null, conflictsOpen: false, panelOpen: true })
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
    conflictsOpen: false,
    cloneOpen: false,
    cloning: false,

    reset: (localId) => set({ localId, ...EMPTY, panelOpen: false, linkOpen: false, conflictsOpen: false }),

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
      await run('push', async (localId) => {
        const status = await pushWorkspace(localId)
        apply(localId, status)
        if (status.state === 'conflict') {
          showSession(localId, await getSyncSession(localId))
        }
      })
    },

    pull: async () => {
      await run('pull', async (localId) => {
        const result = await pullWorkspace(localId)
        apply(localId, result.status)
        showSession(localId, result.session)
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

    cloneWorkspace: async (input) => {
      if (get().cloning) {
        return null
      }
      set({ cloning: true })
      try {
        const summary = await cloneAsWorkspace(input)
        set({ cloneOpen: false })
        await useWorkspacesStore.getState().load()
        await useWorkspacesStore.getState().activate(summary.localId)
        return null
      } catch (error) {
        return toAppError(error)
      } finally {
        set({ cloning: false })
      }
    },

    saveSettings: async (change) => {
      const settings = get().status?.settings
      if (!settings) {
        return
      }
      const next = { ...settings, ...change }
      await run('settings', async (localId) => {
        await saveSyncSettings(localId, next.accountId, next.autoPull, next.autoPush)
        apply(localId, await getSyncStatus(localId))
      })
    },

    unlink: async (targetFolder) =>
      run('unlink', async (localId) => {
        await unlinkRepository(localId, targetFolder)
        set({ changes: [], excluded: [], message: '', session: null })
        await useWorkspacesStore.getState().load()
        await useWorkspacesStore.getState().activate(localId)
        apply(localId, await getSyncStatus(localId))
      }),

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
    openClone: () => set({ cloneOpen: true }),
    closeClone: () => set({ cloneOpen: false }),

    openConflicts: async () => {
      await run('conflicts', async (localId) => {
        const session = await getSyncSession(localId)
        showSession(localId, session)
        if (session === null && get().localId === localId) {
          set({ panelOpen: true })
        }
      })
    },

    closeConflicts: () => set({ conflictsOpen: false }),

    selectConflict: (itemId) => set({ selectedConflictId: itemId }),

    resolve: async (itemId, resolution, fieldPath = '') => {
      const session = get().session
      if (session === null) {
        return
      }
      await run('resolve', async () => {
        const resolved = await resolveConflict(session.id, itemId, resolution, fieldPath)
        const current = get().session
        if (current?.id === session.id) {
          set({ session: { ...current, conflicts: current.conflicts.map((conflict) => (conflict.itemId === itemId ? resolved : conflict)) } })
        }
      })
    },

    finishPull: async () => {
      const session = get().session
      if (session === null) {
        return
      }
      await run('finish', async (localId) => {
        closeSession(localId, await finishPull(session.id))
        await reloadChanges(localId)
      })
    },

    cancelPull: async () => {
      const session = get().session
      if (session === null) {
        return
      }
      await run('cancel', async (localId) => closeSession(localId, await cancelPull(session.id)))
    },
  }
})
