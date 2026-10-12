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
  /** Every open collection's documentation, by collection id; kept while its tab is hidden. */
  entries: Record<string, DocumentationEntry>
  /** What the last confirmed save wrote; the app refreshes the tree and open requests from it. */
  saved: DocumentationSaved | null
  /** Starts tracking a collection (once) and generates its documentation. */
  open: (collectionId: string) => Promise<void>
  /** Generates the documentation again; unsaved YAML stays and is marked stale when the collection changed. */
  load: (collectionId: string) => Promise<void>
  /** Generates every open documentation again (after the collection or its requests changed). */
  refreshAll: () => Promise<void>
  edit: (collectionId: string, yaml: string) => void
  setView: (collectionId: string, view: DocumentationView) => void
  /** Plans the save; `false` while the plan waits for {@link confirm} (nothing is written yet). */
  save: (collectionId: string) => Promise<Result<boolean, AppError>>
  /** Writes the planned changes; deletions are only written when `confirmDeletions` is set. */
  confirm: (collectionId: string, confirmDeletions: boolean) => Promise<Result<boolean, AppError>>
  cancelPlan: (collectionId: string) => void
  /** Keeps the edited YAML after the collection changed on disk. */
  keepDraft: (collectionId: string) => void
  /** Drops the edited YAML and shows the documentation generated from the collection. */
  discard: (collectionId: string) => void
  drop: (collectionId: string) => void
  reset: () => void
}

/** Whether the collection's documentation has unsaved YAML edits. */
export function isDocumentationDirty(state: Pick<DocumentationState, 'entries'>, collectionId: string): boolean {
  return state.entries[collectionId]?.dirty === true
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

function cancelPreview(collectionId: string) {
  clearTimeout(previewTimers.get(collectionId))
  previewTimers.delete(collectionId)
}

export const useDocumentationStore = create<DocumentationState>()((set, get) => {
  function patch(collectionId: string, update: Partial<DocumentationEntry> | ((entry: DocumentationEntry) => Partial<DocumentationEntry>)) {
    set((state) => {
      const entry = state.entries[collectionId]
      if (!entry) {
        return state
      }
      return { entries: { ...state.entries, [collectionId]: { ...entry, ...(typeof update === 'function' ? update(entry) : update) } } }
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

  async function readDraft(collectionId: string) {
    const draft = get().entries[collectionId]?.draft
    if (draft === undefined) {
      return
    }
    const read = await inWorkspace((localId) => previewDocumentation(localId, collectionId, draft))
    if (get().entries[collectionId]?.draft !== draft || (!read.ok && read.error.code === NO_WORKSPACE)) {
      return
    }
    patch(collectionId, read.ok ? { preview: read.value.document, problem: null } : { problem: read.error })
  }

  return {
    entries: {},
    saved: null,

    open: async (collectionId) => {
      if (get().entries[collectionId]) {
        return
      }
      set((state) => ({ entries: { ...state.entries, [collectionId]: loadingEntry() } }))
      await get().load(collectionId)
    },

    load: async (collectionId) => {
      if (!get().entries[collectionId]) {
        return
      }
      const loaded = await inWorkspace((localId) => getDocumentation(localId, collectionId))
      if (!loaded.ok) {
        if (loaded.error.code !== NO_WORKSPACE) {
          patch(collectionId, { status: 'ready', loadError: loaded.error })
        }
        return
      }
      patch(collectionId, (entry) => {
        if (!entry.dirty) {
          return cleanEntry(loaded.value)
        }
        const changed = entry.base !== null && entry.base.yaml !== loaded.value.yaml
        return { status: 'ready', loadError: null, base: loaded.value, stale: entry.stale || changed }
      })
    },

    refreshAll: async () => {
      await Promise.all(Object.keys(get().entries).map((collectionId) => get().load(collectionId)))
    },

    edit: (collectionId, yaml) => {
      const entry = get().entries[collectionId]
      if (!entry || entry.draft === yaml) {
        return
      }
      patch(collectionId, { draft: yaml, dirty: entry.base === null || entry.base.yaml !== yaml, failure: null })
      cancelPreview(collectionId)
      previewTimers.set(
        collectionId,
        setTimeout(() => {
          previewTimers.delete(collectionId)
          void readDraft(collectionId)
        }, PREVIEW_DELAY_MS),
      )
    },

    setView: (collectionId, view) => patch(collectionId, { view }),

    save: async (collectionId) => {
      const entry = get().entries[collectionId]
      if (!entry?.dirty) {
        return { ok: true, value: true }
      }
      patch(collectionId, { saving: true, failure: null })
      const draft = entry.draft
      const planned = await inWorkspace((localId) => planDocumentation(localId, collectionId, draft))
      if (!planned.ok) {
        patch(collectionId, { saving: false, failure: planned.error, problem: planned.error.details.length > 0 ? planned.error : null })
        return { ok: false, error: planned.error }
      }
      if (isEmptyPlan(planned.value)) {
        patch(collectionId, (current) => (current.base && current.draft === draft ? { ...cleanEntry(current.base), saving: false } : { saving: false }))
        return { ok: true, value: true }
      }
      patch(collectionId, { saving: false, plan: planned.value })
      return { ok: true, value: false }
    },

    confirm: async (collectionId, confirmDeletions) => {
      const entry = get().entries[collectionId]
      if (!entry?.plan) {
        return { ok: true, value: false }
      }
      patch(collectionId, { saving: true, failure: null })
      const draft = entry.draft
      const localId = activeLocalId()
      const applied = await inWorkspace((workspace) => applyDocumentation(workspace, collectionId, draft, confirmDeletions))
      if (!applied.ok) {
        patch(collectionId, { saving: false, failure: applied.error })
        return { ok: false, error: applied.error }
      }
      const edited = get().entries[collectionId]?.draft !== draft
      patch(collectionId, edited ? { saving: false, plan: null, base: applied.value.documentation } : { ...cleanEntry(applied.value.documentation), saving: false, plan: null })
      set({
        saved: {
          localId: localId ?? '',
          collectionId,
          collectionChanged: applied.value.collectionChanged,
          updated: applied.value.updated,
          created: applied.value.created,
          deleted: applied.value.deleted,
        },
      })
      return { ok: true, value: true }
    },

    cancelPlan: (collectionId) => patch(collectionId, { plan: null }),

    keepDraft: (collectionId) => patch(collectionId, { stale: false }),

    discard: (collectionId) => {
      cancelPreview(collectionId)
      patch(collectionId, (entry) => (entry.base ? { ...cleanEntry(entry.base), plan: null, failure: null } : { dirty: false }))
    },

    drop: (collectionId) => {
      cancelPreview(collectionId)
      set((state) => {
        if (!state.entries[collectionId]) {
          return state
        }
        const entries = { ...state.entries }
        delete entries[collectionId]
        return { entries }
      })
    },

    reset: () => {
      for (const collectionId of previewTimers.keys()) {
        cancelPreview(collectionId)
      }
      set({ entries: {}, saved: null })
    },
  }
})
