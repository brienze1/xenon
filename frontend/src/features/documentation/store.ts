import { create } from 'zustand'
import { useWorkspacesStore } from '@/features/workspaces/store'
import { AppError, toAppError } from '@/platform/errors'
import type { Documentation } from '@/platform/mappers/documentation'
import { applyDocumentation, getDocumentation, planDocumentation, previewDocumentation } from '@/platform/services/tree'
import type { Result } from '@/shared/types/result'
import { isEmptyPlan } from './lib/plan-summary'
import type { DocumentationEntry, DocumentationSaved, DocumentationView } from './model'

const NO_WORKSPACE = 'WKS-00000'
export const PREVIEW_DELAY_MS = 300

interface DocumentationState {
  /** Every open collection's or folder's documentation, by item id; kept while its tab is hidden. */
  entries: Record<string, DocumentationEntry>
  /** What the last confirmed save wrote; the app refreshes the tree and open requests from it. */
  saved: DocumentationSaved | null
  /** Starts tracking a collection or folder (once) and generates its documentation. */
  open: (itemId: string) => Promise<void>
  /** Generates the documentation again; unsaved YAML stays and is marked stale when the documented item changed. */
  load: (itemId: string) => Promise<void>
  /** Generates every open documentation again (after a collection, folder or request changed). */
  refreshAll: () => Promise<void>
  edit: (itemId: string, yaml: string) => void
  setView: (itemId: string, view: DocumentationView) => void
  /** Plans the save; `false` while the plan waits for {@link confirm} (nothing is written yet). */
  save: (itemId: string) => Promise<Result<boolean, AppError>>
  /** Writes the planned changes; deletions are only written when `confirmDeletions` is set. */
  confirm: (itemId: string, confirmDeletions: boolean) => Promise<Result<boolean, AppError>>
  cancelPlan: (itemId: string) => void
  /** Keeps the edited YAML after the documented item changed on disk. */
  keepDraft: (itemId: string) => void
  /** Drops the edited YAML and shows the documentation generated from the collection or folder. */
  discard: (itemId: string) => void
  drop: (itemId: string) => void
  reset: () => void
}

/** Whether the collection's or folder's documentation has unsaved YAML edits. */
export function isDocumentationDirty(state: Pick<DocumentationState, 'entries'>, itemId: string): boolean {
  return state.entries[itemId]?.dirty === true
}

function activeLocalId(): string | null {
  return useWorkspacesStore.getState().activeLocalId
}

function loadingEntry(): DocumentationEntry {
  return {
    status: 'loading',
    loadError: null,
    base: null,
    draft: '',
    dirty: false,
    view: 'preview',
    preview: null,
    problem: null,
    plan: null,
    saving: false,
    failure: null,
    stale: false,
  }
}

function cleanEntry(base: Documentation): Partial<DocumentationEntry> {
  return { status: 'ready', loadError: null, base, draft: base.yaml, dirty: false, preview: base.document, problem: null, stale: false }
}

const previewTimers = new Map<string, ReturnType<typeof setTimeout>>()

function cancelPreview(itemId: string) {
  clearTimeout(previewTimers.get(itemId))
  previewTimers.delete(itemId)
}

export const useDocumentationStore = create<DocumentationState>()((set, get) => {
  function patch(itemId: string, update: Partial<DocumentationEntry> | ((entry: DocumentationEntry) => Partial<DocumentationEntry>)) {
    set((state) => {
      const entry = state.entries[itemId]
      if (!entry) {
        return state
      }
      return { entries: { ...state.entries, [itemId]: { ...entry, ...(typeof update === 'function' ? update(entry) : update) } } }
    })
  }

  async function inWorkspace<T>(operation: (localId: string) => Promise<T>): Promise<Result<T, AppError>> {
    const localId = activeLocalId()
    if (localId === null) {
      return { ok: false, error: new AppError(NO_WORKSPACE, 'Open a workspace first') }
    }
    try {
      const value = await operation(localId)
      if (activeLocalId() !== localId) {
        return { ok: false, error: new AppError(NO_WORKSPACE, 'Open a workspace first') }
      }
      return { ok: true, value }
    } catch (error) {
      return { ok: false, error: toAppError(error) }
    }
  }

  async function readDraft(itemId: string) {
    const draft = get().entries[itemId]?.draft
    if (draft === undefined) {
      return
    }
    const read = await inWorkspace((localId) => previewDocumentation(localId, itemId, draft))
    if (get().entries[itemId]?.draft !== draft || (!read.ok && read.error.code === NO_WORKSPACE)) {
      return
    }
    patch(itemId, read.ok ? { preview: read.value.document, problem: null } : { problem: read.error })
  }

  return {
    entries: {},
    saved: null,

    open: async (itemId) => {
      if (get().entries[itemId]) {
        return
      }
      set((state) => ({ entries: { ...state.entries, [itemId]: loadingEntry() } }))
      await get().load(itemId)
    },

    load: async (itemId) => {
      if (!get().entries[itemId]) {
        return
      }
      const loaded = await inWorkspace((localId) => getDocumentation(localId, itemId))
      if (!loaded.ok) {
        if (loaded.error.code !== NO_WORKSPACE) {
          patch(itemId, { status: 'ready', loadError: loaded.error })
        }
        return
      }
      patch(itemId, (entry) => {
        if (!entry.dirty) {
          return cleanEntry(loaded.value)
        }
        const changed = entry.base !== null && entry.base.yaml !== loaded.value.yaml
        return { status: 'ready', loadError: null, base: loaded.value, stale: entry.stale || changed }
      })
    },

    refreshAll: async () => {
      await Promise.all(Object.keys(get().entries).map((itemId) => get().load(itemId)))
    },

    edit: (itemId, yaml) => {
      const entry = get().entries[itemId]
      if (!entry || entry.draft === yaml) {
        return
      }
      patch(itemId, { draft: yaml, dirty: entry.base === null || entry.base.yaml !== yaml, failure: null })
      cancelPreview(itemId)
      previewTimers.set(
        itemId,
        setTimeout(() => {
          previewTimers.delete(itemId)
          void readDraft(itemId)
        }, PREVIEW_DELAY_MS),
      )
    },

    setView: (itemId, view) => patch(itemId, { view }),

    save: async (itemId) => {
      const entry = get().entries[itemId]
      if (!entry?.dirty) {
        return { ok: true, value: true }
      }
      patch(itemId, { saving: true, failure: null })
      const draft = entry.draft
      const planned = await inWorkspace((localId) => planDocumentation(localId, itemId, draft))
      if (!planned.ok) {
        patch(itemId, { saving: false, failure: planned.error, problem: planned.error.details.length > 0 ? planned.error : null })
        return { ok: false, error: planned.error }
      }
      if (isEmptyPlan(planned.value)) {
        patch(itemId, (current) => (current.base && current.draft === draft ? { ...cleanEntry(current.base), saving: false } : { saving: false }))
        return { ok: true, value: true }
      }
      patch(itemId, { saving: false, plan: planned.value })
      return { ok: true, value: false }
    },

    confirm: async (itemId, confirmDeletions) => {
      const entry = get().entries[itemId]
      if (!entry?.plan) {
        return { ok: true, value: false }
      }
      patch(itemId, { saving: true, failure: null })
      const draft = entry.draft
      const localId = activeLocalId()
      const applied = await inWorkspace((workspace) => applyDocumentation(workspace, itemId, draft, confirmDeletions))
      if (!applied.ok) {
        patch(itemId, { saving: false, failure: applied.error })
        return { ok: false, error: applied.error }
      }
      const edited = get().entries[itemId]?.draft !== draft
      patch(itemId, edited ? { saving: false, plan: null, base: applied.value.documentation } : { ...cleanEntry(applied.value.documentation), saving: false, plan: null })
      set({
        saved: {
          localId: localId ?? '',
          itemId,
          collectionChanged: applied.value.collectionChanged,
          updated: applied.value.updated,
          created: applied.value.created,
          deleted: applied.value.deleted,
        },
      })
      return { ok: true, value: true }
    },

    cancelPlan: (itemId) => patch(itemId, { plan: null }),

    keepDraft: (itemId) => patch(itemId, { stale: false }),

    discard: (itemId) => {
      cancelPreview(itemId)
      patch(itemId, (entry) => (entry.base ? { ...cleanEntry(entry.base), plan: null, failure: null } : { dirty: false }))
    },

    drop: (itemId) => {
      cancelPreview(itemId)
      set((state) => {
        if (!state.entries[itemId]) {
          return state
        }
        const entries = { ...state.entries }
        delete entries[itemId]
        return { entries }
      })
    },

    reset: () => {
      for (const itemId of previewTimers.keys()) {
        cancelPreview(itemId)
      }
      set({ entries: {}, saved: null })
    },
  }
})
