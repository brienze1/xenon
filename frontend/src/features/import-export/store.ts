import { create } from 'zustand'
import { useSidebarStore } from '@/features/sidebar/store'
import { useTabsStore } from '@/features/tabs/store'
import { useWorkspacesStore } from '@/features/workspaces/store'
import { toAppError, type AppError } from '@/platform/errors'
import type { ImportProgress } from '@/platform/events'
import type { ClashResolution, ImportClash, ImportPreview, ImportResult } from '@/platform/mappers/import'
import { detectImport, discardImport, findImportClashes, pickImportSources, runImport, type ImportInput, type PendingImport } from '@/platform/services/import'
import { pickFolder } from '@/platform/services/dialogs'
import { getTree } from '@/platform/services/tree'
import { assignIncluded, buildPlan, containerLocations, createsWorkspaces, defaultTargets, pendingClashes, withRequestLocations } from './lib/plan'
import { ALL_INCLUDED, type ImportLocation, type ImportProgressState, type ImportStep, type ImportTarget, type PassphraseRequest } from './model'

const PASSPHRASE_MISSING = 'IMP-50401'
const PASSPHRASE_WRONG = 'IMP-08401'
const FOLDER_PICKER_TITLE = 'Choose an empty folder for the imported workspace'

interface ImportData {
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
  /** Opens the export's encrypted secrets; kept only while the dialog is open, never stored or logged. */
  passphrase: string
  skipSecrets: boolean
  passphraseRequest: PassphraseRequest | null
}

interface ImportState extends ImportData {
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
  chooseFolder: (key: string) => Promise<void>
  setPassphrase: (passphrase: string) => void
  submitPassphrase: () => Promise<void>
  importWithoutSecrets: () => Promise<void>
  createWorkspace: (name: string, path: string) => Promise<boolean>
  start: () => Promise<void>
  resolve: (resolution: ClashResolution, applyToAll: boolean) => Promise<void>
  cancel: () => void
  applyProgress: (progress: ImportProgress) => void
  openItem: (localId: string, itemId: string) => Promise<void>
  reset: () => void
}

let pending: PendingImport | null = null

const initialState: ImportData = {
  step: 'closed',
  preview: null,
  targets: {},
  locations: {},
  clashes: [],
  resolutions: {},
  applyToAll: null,
  progress: null,
  result: null,
  error: null,
  newWorkspaceFor: null,
  passphrase: '',
  skipSecrets: false,
  passphraseRequest: null,
}

function discard(preview: ImportPreview | null): void {
  if (preview) {
    void discardImport(preview.sourceId).catch(() => undefined)
  }
}

export const useImportStore = create<ImportState>()((set, get) => {
  async function run() {
    const { preview, targets, resolutions, applyToAll, passphrase, skipSecrets } = get()
    if (!preview) {
      return
    }
    set({ step: 'importing', progress: null, error: null })
    const plan = { ...buildPlan(preview.sourceId, targets, resolutions, applyToAll), ...(passphrase ? { passphrase } : {}), ...(skipSecrets ? { skipSecrets } : {}) }
    pending = runImport(plan)
    try {
      const result = await pending.result
      set({ step: 'summary', result, passphrase: '', passphraseRequest: null })
      if (createsWorkspaces(preview, targets)) {
        void useWorkspacesStore.getState().load()
      }
      const activeLocalId = useWorkspacesStore.getState().activeLocalId
      if (result.items.some((item) => item.localId === activeLocalId)) {
        void useSidebarStore.getState().load()
      }
    } catch (error) {
      const failure = toAppError(error)
      if (failure.code === PASSPHRASE_MISSING || failure.code === PASSPHRASE_WRONG) {
        set({ step: 'passphrase', passphrase: '', passphraseRequest: failure.code === PASSPHRASE_WRONG ? 'wrong' : 'missing' })
      } else {
        set({ step: 'preview', error: failure })
      }
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

    chooseFolder: async (key) => {
      try {
        const folder = await pickFolder(FOLDER_PICKER_TITLE)
        if (folder) {
          get().setTarget(key, { folder })
        }
      } catch (error) {
        set({ error: toAppError(error) })
      }
    },

    setPassphrase: (passphrase) => set({ passphrase }),

    submitPassphrase: async () => {
      if (get().passphrase === '') {
        return
      }
      set({ skipSecrets: false })
      await run()
    },

    importWithoutSecrets: async () => {
      set({ passphrase: '', skipSecrets: true })
      await run()
    },

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
