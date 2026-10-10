import { create } from 'zustand'
import { useSidebarStore } from '@/features/sidebar/store'
import { useTabsStore } from '@/features/tabs/store'
import { useWorkspacesStore } from '@/features/workspaces/store'
import { toAppError, type AppError } from '@/platform/errors'
import type { ImportProgress } from '@/platform/events'
import type { ClashResolution, ImportClash, ImportPreview, ImportResult } from '@/platform/mappers/import'
import { detectImport, discardImport, findImportClashes, pickImportSources, runImport, type ImportInput, type PendingImport } from '@/platform/services/import'
import { getTree } from '@/platform/services/tree'
import { assignIncluded, buildPlan, containerLocations, defaultTargets, pendingClashes, withRequestLocations } from './lib/plan'
import { ALL_INCLUDED, type ImportLocation, type ImportProgressState, type ImportStep, type ImportTarget } from './model'

interface ImportState {
  step: ImportStep
  preview: ImportPreview | null
  targets: Record<string, ImportTarget>
  locations: Record<string, ImportLocation[]>
  clashes: ImportClash[]
  resolutions: Record<string, ClashResolution>
  applyToAll: ClashResolution | null
  progress: ImportProgressState | null
  result: ImportResult | null
  error: AppError | null
  newWorkspaceFor: string | null
  open: () => void
  close: () => void
  back: () => void
  pickSources: () => Promise<void>
  detect: (input: ImportInput) => Promise<void>
  setTarget: (key: string, patch: Partial<ImportTarget>) => void
  loadLocations: (localId: string) => Promise<void>
  assignIncluded: (localId: string) => void
  startNewWorkspace: (forKey: string) => void
  cancelNewWorkspace: () => void
  createWorkspace: (name: string, path: string) => Promise<boolean>
  start: () => Promise<void>
  resolve: (resolution: ClashResolution, applyToAll: boolean) => Promise<void>
  cancel: () => void
  applyProgress: (progress: ImportProgress) => void
  openItem: (localId: string, itemId: string) => Promise<void>
  reset: () => void
}

let pending: PendingImport | null = null

const initialState = {
  step: 'closed' as ImportStep,
  preview: null as ImportPreview | null,
  targets: {} as Record<string, ImportTarget>,
  locations: {} as Record<string, ImportLocation[]>,
  clashes: [] as ImportClash[],
  resolutions: {} as Record<string, ClashResolution>,
  applyToAll: null as ClashResolution | null,
  progress: null as ImportProgressState | null,
  result: null as ImportResult | null,
  error: null as AppError | null,
  newWorkspaceFor: null as string | null,
}

function discard(preview: ImportPreview | null): void {
  if (preview) {
    void discardImport(preview.sourceId).catch(() => undefined)
  }
}

export const useImportStore = create<ImportState>()((set, get) => {
  async function run() {
    const { preview, targets, resolutions, applyToAll } = get()
    if (!preview) {
      return
    }
    set({ step: 'importing', progress: null, error: null })
    pending = runImport(buildPlan(preview.sourceId, targets, resolutions, applyToAll))
    try {
      const result = await pending.result
      set({ step: 'summary', result })
      for (const localId of new Set(result.items.map((item) => item.localId))) {
        if (localId === useWorkspacesStore.getState().activeLocalId) {
          void useSidebarStore.getState().load()
        }
      }
    } catch (error) {
      set({ step: 'preview', error: toAppError(error) })
    } finally {
      pending = null
    }
  }

  return {
    ...initialState,

    open: () => {
      if (get().step === 'closed') {
        set({ ...initialState, step: 'choose' })
      }
    },

    close: () => {
      pending?.cancel()
      discard(get().preview)
      set({ ...initialState })
    },

    back: () => {
      discard(get().preview)
      set({ ...initialState, step: 'choose' })
    },

    pickSources: async () => {
      try {
        const paths = await pickImportSources()
        if (paths.length > 0) {
          await get().detect({ paths })
        }
      } catch (error) {
        set({ error: toAppError(error) })
      }
    },

    detect: async (input) => {
      discard(get().preview)
      set({ step: 'detecting', preview: null, error: null })
      try {
        const preview = await detectImport(input)
        const localId = useWorkspacesStore.getState().activeLocalId ?? ''
        set({ step: 'preview', preview, targets: defaultTargets(preview, localId), clashes: [], resolutions: {}, applyToAll: null })
        if (localId) {
          void get().loadLocations(localId)
        }
      } catch (error) {
        set({ step: 'choose', error: toAppError(error) })
      }
    },

    setTarget: (key, patch) => {
      const current = get().targets[key]
      if (!current) {
        return
      }
      const next = { ...current, ...patch }
      if (patch.localId !== undefined && patch.localId !== current.localId) {
        next.parentId = ''
        void get().loadLocations(patch.localId)
      }
      set((state) => ({ targets: { ...state.targets, [key]: next } }))
    },

    loadLocations: async (localId) => {
      try {
        const locations = containerLocations(await getTree(localId))
        set((state) => ({
          locations: { ...state.locations, [localId]: locations },
          targets: state.preview ? withRequestLocations(state.preview, state.targets, localId, locations) : state.targets,
        }))
      } catch (error) {
        set({ error: toAppError(error) })
      }
    },

    assignIncluded: (localId) => {
      set((state) => ({ targets: assignIncluded(state.targets, localId) }))
      void get().loadLocations(localId)
    },

    startNewWorkspace: (forKey) => {
      useWorkspacesStore.getState().clearError()
      set({ newWorkspaceFor: forKey })
    },

    cancelNewWorkspace: () => set({ newWorkspaceFor: null }),

    createWorkspace: async (name, path) => {
      const forKey = get().newWorkspaceFor
      const created = await useWorkspacesStore.getState().add(name, path)
      if (created === null) {
        return false
      }
      set({ newWorkspaceFor: null })
      if (forKey === ALL_INCLUDED) {
        get().assignIncluded(created.localId)
      } else if (forKey !== null) {
        get().setTarget(forKey, { localId: created.localId })
      }
      return true
    },

    start: async () => {
      const { preview, targets } = get()
      if (!preview) {
        return
      }
      set({ error: null })
      try {
        const clashes = await findImportClashes(buildPlan(preview.sourceId, targets, {}, null))
        if (clashes.length > 0) {
          set({ step: 'clashes', clashes, resolutions: {}, applyToAll: null })
          return
        }
      } catch (error) {
        set({ error: toAppError(error) })
        return
      }
      await run()
    },

    resolve: async (resolution, applyToAll) => {
      const { clashes, resolutions } = get()
      const [current] = pendingClashes(clashes, resolutions)
      if (applyToAll || !current) {
        set({ applyToAll: resolution })
        await run()
        return
      }
      const next = { ...resolutions, [current.itemKey]: resolution }
      set({ resolutions: next })
      if (pendingClashes(clashes, next).length === 0) {
        await run()
      }
    },

    cancel: () => {
      pending?.cancel()
    },

    applyProgress: (progress) => {
      if (get().step === 'importing' && progress.sourceId === get().preview?.sourceId) {
        set({ progress: { done: progress.done, total: progress.total, currentItem: progress.currentItem } })
      }
    },

    openItem: async (localId, itemId) => {
      get().close()
      if (useWorkspacesStore.getState().activeLocalId !== localId) {
        await useWorkspacesStore.getState().activate(localId)
      }
      await useSidebarStore.getState().load()
      const node = useSidebarStore.getState().reveal(itemId)
      if (node?.kind === 'request') {
        await useTabsStore.getState().openPinned(itemId)
      }
    },

    reset: () => set({ ...initialState }),
  }
})
