import { create } from 'zustand'
import { useWorkspacesStore } from '@/features/workspaces/store'
import { AppError, toAppError } from '@/platform/errors'
import type { ItemSettings } from '@/platform/mappers/tree'
import { formatBulk, parseBulk } from '@/platform/services/requests'
import { getItemSettings, saveItemSettings } from '@/platform/services/tree'
import type { Result } from '@/shared/types/result'
import { sameSettings } from './lib/same-settings'
import type { ItemEntry, ItemFailure, ItemSection } from './model'

const NO_WORKSPACE = 'WKS-00000'
const SAVE_NEXT_STEP = 'Check that the workspace folder exists and is writable, then try again.'
const FORMAT_NEXT_STEP = 'Try again, or keep editing the headers in the table.'
const BULK_NEXT_STEP = 'Check that each line reads key:value, then try again.'

interface ItemSettingsState {
  /** Every open collection or folder tab's settings, by item id; kept while the tab is hidden. */
  items: Record<string, ItemEntry>
  /** Starts tracking an item (once) and reads its settings. */
  open: (itemId: string) => Promise<void>
  /** Reads the item's settings again; unsaved edits stay and the tab is marked stale if disk changed. */
  load: (itemId: string) => Promise<void>
  /** Reads every open item again (after the workspace tree changed). */
  refreshAll: () => Promise<void>
  edit: (itemId: string, patch: Partial<Omit<ItemSettings, 'id' | 'kind' | 'name'>>) => void
  /** Shows a section; an item not tracked yet starts loading on it (Run… on a closed tab). */
  setSection: (itemId: string, section: ItemSection) => void
  /** Switches the headers between the table and `key:value` lines. */
  setBulk: (itemId: string, bulk: boolean) => Promise<void>
  editBulkText: (itemId: string, text: string) => void
  /** Saves the edits; `false` when there was nothing to save or the item is not loaded. */
  save: (itemId: string) => Promise<Result<boolean, AppError>>
  /** Drops the edits and shows what is on disk. */
  discard: (itemId: string) => void
  /** Edits the name draft; whether it is pending is derived from the saved name, so it does not mark the settings dirty. */
  editName: (itemId: string, name: string) => void
  /** Stops tracking an item (its tab closed). */
  drop: (itemId: string) => void
  reset: () => void
}

/** Whether the open collection or folder has unsaved edits. */
export function isItemDirty(state: Pick<ItemSettingsState, 'items'>, itemId: string): boolean {
  return state.items[itemId]?.dirty === true
}

/** Whether the item's name was edited in its tab and not saved yet; compared with the name it was loaded with. */
export function isItemNamePending(state: Pick<ItemSettingsState, 'items'>, itemId: string): boolean {
  return hasPendingName(state.items[itemId])
}

function hasPendingName(entry: ItemEntry | undefined): boolean {
  return entry?.draft != null && entry.base != null && entry.draft.name.trim() !== entry.base.name
}

function activeLocalId(): string | null {
  return useWorkspacesStore.getState().activeLocalId
}

function loadingEntry(): ItemEntry {
  return { status: 'loading', loadError: null, base: null, draft: null, bulkText: null, dirty: false, saving: false, failure: null, stale: false, section: 'overview' }
}

export const useItemSettingsStore = create<ItemSettingsState>()((set, get) => {
  function patch(itemId: string, update: Partial<ItemEntry> | ((entry: ItemEntry) => Partial<ItemEntry>)) {
    set((state) => {
      const entry = state.items[itemId]
      if (!entry) {
        return state
      }
      return { items: { ...state.items, [itemId]: { ...entry, ...(typeof update === 'function' ? update(entry) : update) } } }
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

  function fail(itemId: string, error: AppError, nextStep: string): ItemFailure {
    const failure = { error, nextStep }
    patch(itemId, { failure, saving: false })
    return failure
  }

  /** A name typed in the tab survives a reload; otherwise the draft takes the loaded name. */
  function nameAfterLoad(entry: ItemEntry, loaded: ItemSettings): string {
    return hasPendingName(entry) ? entry.draft!.name : loaded.name
  }

  function applyLoaded(entry: ItemEntry, loaded: ItemSettings): Partial<ItemEntry> {
    const name = nameAfterLoad(entry, loaded)
    if (!entry.dirty || !entry.draft) {
      const unchanged = entry.base !== null && sameSettings(entry.base, loaded)
      return { status: 'ready', loadError: null, base: loaded, draft: { ...loaded, name }, bulkText: unchanged ? entry.bulkText : null, stale: false }
    }
    const changed = entry.base === null || !sameSettings(entry.base, loaded)
    return { status: 'ready', loadError: null, base: loaded, draft: { ...entry.draft, name }, stale: entry.stale || changed }
  }

  return {
    items: {},

    open: async (itemId) => {
      if (get().items[itemId]) {
        return
      }
      set((state) => ({ items: { ...state.items, [itemId]: loadingEntry() } }))
      await get().load(itemId)
    },

    load: async (itemId) => {
      if (!get().items[itemId]) {
        return
      }
      const loaded = await inWorkspace((localId) => getItemSettings(localId, itemId))
      if (!loaded.ok) {
        if (loaded.error.code !== NO_WORKSPACE) {
          patch(itemId, { status: 'ready', loadError: loaded.error })
        }
        return
      }
      patch(itemId, (entry) => applyLoaded(entry, loaded.value))
    },

    refreshAll: async () => {
      await Promise.all(Object.keys(get().items).map((itemId) => get().load(itemId)))
    },

    edit: (itemId, update) => patch(itemId, (entry) => (entry.draft ? { draft: { ...entry.draft, ...update }, dirty: true } : {})),

    setSection: (itemId, section) => {
      if (get().items[itemId]) {
        patch(itemId, { section })
        return
      }
      // Run… opens the tab on its Runner before the tab mounted: track the item with that section and read it.
      set((state) => ({ items: { ...state.items, [itemId]: { ...loadingEntry(), section } } }))
      void get().load(itemId)
    },

    setBulk: async (itemId, bulk) => {
      const entry = get().items[itemId]
      if (!entry?.draft) {
        return
      }
      if (bulk) {
        const text = await inWorkspace(() => formatBulk(entry.draft!.headers))
        if (!text.ok) {
          fail(itemId, text.error, FORMAT_NEXT_STEP)
          return
        }
        patch(itemId, { bulkText: text.value, failure: null })
        return
      }
      const headers = await inWorkspace(() => parseBulk(entry.bulkText ?? ''))
      if (!headers.ok) {
        fail(itemId, headers.error, BULK_NEXT_STEP)
        return
      }
      patch(itemId, (current) => ({ bulkText: null, failure: null, draft: current.draft ? { ...current.draft, headers: headers.value } : null }))
    },

    editBulkText: (itemId, text) => patch(itemId, { bulkText: text, dirty: true }),

    save: async (itemId) => {
      const entry = get().items[itemId]
      if (!entry?.draft) {
        return { ok: true, value: false }
      }
      if (!entry.dirty) {
        return { ok: true, value: true }
      }
      patch(itemId, { saving: true, failure: null })
      const draft = entry.draft
      const headers = entry.bulkText === null ? { ok: true as const, value: draft.headers } : await inWorkspace(() => parseBulk(entry.bulkText ?? ''))
      if (!headers.ok) {
        return { ok: false, error: fail(itemId, headers.error, BULK_NEXT_STEP).error }
      }
      const name = entry.base?.name ?? draft.name
      const saved = await inWorkspace((localId) => saveItemSettings(localId, { ...draft, name, headers: headers.value }))
      if (!saved.ok) {
        return { ok: false, error: fail(itemId, saved.error, SAVE_NEXT_STEP).error }
      }
      const edited = get().items[itemId]?.draft !== draft
      patch(itemId, edited ? { saving: false, base: saved.value, stale: false } : { saving: false, base: saved.value, draft: { ...saved.value, name: draft.name }, dirty: false, stale: false })
      return { ok: true, value: true }
    },

    discard: (itemId) => patch(itemId, (entry) => ({ draft: entry.base, dirty: false, stale: false, bulkText: null, failure: null })),

    editName: (itemId, name) => patch(itemId, (entry) => (entry.draft ? { draft: { ...entry.draft, name } } : {})),

    drop: (itemId) =>
      set((state) => {
        if (!state.items[itemId]) {
          return state
        }
        const items = { ...state.items }
        delete items[itemId]
        return { items }
      }),

    reset: () => set({ items: {} }),
  }
})
