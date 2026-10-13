import { create } from 'zustand'
import { useWorkspacesStore } from '@/features/workspaces/store'
import { toAppError, type AppError } from '@/platform/errors'
import type { ExportProgress } from '@/platform/events'
import type { ExportFormat, ExportItemKind, ExportOptions, ExportShape, ExportSummary } from '@/platform/mappers/export'
import { getExportOptions, pickExportTarget, runExport, showExportInFolder, type PendingExport } from '@/platform/services/export'
import { canExport, defaultChoices, exportRequest, suggestedFileName, withFormat } from './lib/choices'
import type { ExportChoices, ExportProgressState, ExportStep, ExportTarget } from './model'

const CANCELLED_BY_BACKEND = 'EXP-05499'

interface ExportData {
  step: ExportStep
  target: ExportTarget | null
  options: ExportOptions | null
  choices: ExportChoices | null
  progress: ExportProgressState | null
  summary: ExportSummary | null
  error: AppError | null
  cancelled: boolean
}

interface ExportState extends ExportData {
  open: (target: ExportTarget) => Promise<void>
  /** Exports an item of the active workspace. */
  openItem: (kind: Exclude<ExportItemKind, 'workspace' | 'globals'>, itemId: string, name: string) => Promise<void>
  /** Exports the whole active workspace (File → Export Workspace…). */
  openWorkspace: () => Promise<void>
  close: () => void
  setFormat: (format: ExportFormat) => void
  setShape: (shape: ExportShape) => void
  setChoice: (patch: Partial<Pick<ExportChoices, 'includeSecrets' | 'passphrase' | 'passphraseConfirmation' | 'currentValuesAsInitial'>>) => void
  start: () => Promise<void>
  cancel: () => void
  applyProgress: (progress: ExportProgress) => void
  showInFolder: () => Promise<void>
  reset: () => void
}

let pending: PendingExport | null = null

const initialState: ExportData = {
  step: 'closed',
  target: null,
  options: null,
  choices: null,
  progress: null,
  summary: null,
  error: null,
  cancelled: false,
}

function wasCancelled(error: AppError): boolean {
  return error.cancelled || error.code === CANCELLED_BY_BACKEND
}

export const useExportStore = create<ExportState>()((set, get) => ({
  ...initialState,

  open: async (target) => {
    if (get().step !== 'closed') {
      return
    }
    set({ ...initialState, step: 'loading', target })
    try {
      const options = await getExportOptions(target)
      if (get().target !== target) {
        return
      }
      set({ step: 'configure', options, choices: defaultChoices(options) })
    } catch (error) {
      if (get().target === target) {
        set({ step: 'configure', error: toAppError(error) })
      }
    }
  },

  openItem: async (kind, itemId, name) => {
    const localId = useWorkspacesStore.getState().activeLocalId
    if (localId) {
      await get().open({ localId, kind, itemId, name })
    }
  },

  openWorkspace: async () => {
    const { activeLocalId, workspaces } = useWorkspacesStore.getState()
    const active = workspaces.find((workspace) => workspace.localId === activeLocalId)
    if (active && !active.missing) {
      await get().open({ localId: active.localId, kind: 'workspace', itemId: '', name: active.name })
    }
  },

  close: () => {
    pending?.cancel()
    pending = null
    set({ ...initialState })
  },

  setFormat: (format) => {
    const { options, choices } = get()
    if (options && choices) {
      set({ choices: withFormat(options, choices, format), error: null, cancelled: false })
    }
  },

  setShape: (shape) => {
    const choices = get().choices
    if (choices) {
      set({ choices: { ...choices, shape }, error: null, cancelled: false })
    }
  },

  setChoice: (patch) => {
    const choices = get().choices
    if (choices) {
      set({ choices: { ...choices, ...patch }, error: null, cancelled: false })
    }
  },

  start: async () => {
    const { target, options, choices } = get()
    if (!target || !options || !choices || !canExport(options, choices)) {
      return
    }
    set({ error: null, cancelled: false })
    let path: string | null
    try {
      path = await pickExportTarget(suggestedFileName(options, choices), choices.format, choices.shape)
    } catch (error) {
      set({ error: toAppError(error) })
      return
    }
    if (path === null || get().target !== target) {
      return
    }
    set({ step: 'exporting', progress: null })
    const run = runExport(exportRequest(target, options, choices, path))
    pending = run
    try {
      const summary = await run.summary
      if (pending === run) {
        set({ step: 'summary', summary, choices: { ...choices, passphrase: '', passphraseConfirmation: '' } })
      }
    } catch (error) {
      if (pending === run) {
        const failure = toAppError(error)
        set(wasCancelled(failure) ? { step: 'configure', cancelled: true } : { step: 'configure', error: failure })
      }
    } finally {
      if (pending === run) {
        pending = null
      }
    }
  },

  cancel: () => {
    pending?.cancel()
  },

  applyProgress: (progress) => {
    if (get().step === 'exporting') {
      set({ progress: { done: progress.done, total: progress.total } })
    }
  },

  showInFolder: async () => {
    const path = get().summary?.path
    if (!path) {
      return
    }
    try {
      await showExportInFolder(path)
    } catch (error) {
      set({ error: toAppError(error) })
    }
  },

  reset: () => {
    pending?.cancel()
    pending = null
    set({ ...initialState })
  },
}))
