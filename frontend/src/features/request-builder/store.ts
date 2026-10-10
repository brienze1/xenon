import { create } from 'zustand'
import { useResponseViewerStore } from '@/features/response-viewer/store'
import { useSidebarStore } from '@/features/sidebar/store'
import { useWorkspacesStore } from '@/features/workspaces/store'
import { SAVE_CONFLICT_CODE, toAppError, type AppError } from '@/platform/errors'
import {
  cancelledResult,
  emptyDraft,
  type KeyValueRow,
  type PathVariableRow,
  type RequestAuth,
  type RequestBody,
  type RequestDraft,
  type RequestSettings,
} from '@/platform/mappers/request'
import {
  applyParams,
  applyUrl,
  autoHeaders,
  formatBulk,
  getRequest,
  parseBulk,
  saveRequest,
  saveRequestAs,
  sendRequest,
  type PendingSend,
} from '@/platform/services/requests'
import { pickFile } from '@/platform/services/dialogs'
import { keepRowIds } from '@/shared/lib/rows'
import type { Result } from '@/shared/types/result'
import { rowsOf, withRows, withSavedRows } from './lib/draft-rows'
import type { BulkTable, SubTab, TabDraft } from './model'

export const SYNC_DELAY_MS = 150

interface RequestBuilderState {
  tabs: Record<string, TabDraft>
  ensureTab: (tabId: string) => void
  openSaved: (tabId: string, requestId: string) => Promise<boolean>
  loadDraft: (tabId: string, draft: RequestDraft, baseRevision: string | null, dirty: boolean) => void
  dropTab: (tabId: string) => void
  detach: (tabId: string) => void
  save: (tabId: string, force?: boolean) => Promise<Result<boolean, AppError>>
  saveAs: (tabId: string, parentId: string, name: string) => Promise<Result<string | null, AppError>>
  adoptRename: (requestId: string) => Promise<void>
  setName: (tabId: string, name: string) => void
  setDescription: (tabId: string, description: string) => void
  setMethod: (tabId: string, method: string) => void
  setUrl: (tabId: string, url: string) => void
  updateParams: (tabId: string, params: KeyValueRow[]) => void
  updatePathVariables: (tabId: string, pathVariables: PathVariableRow[]) => void
  updateHeaders: (tabId: string, headers: KeyValueRow[]) => void
  setAutoHeaderEnabled: (tabId: string, key: string, enabled: boolean) => void
  updateAuth: (tabId: string, auth: Partial<RequestAuth>) => void
  updateBody: (tabId: string, body: Partial<RequestBody>) => void
  updateSettings: (tabId: string, settings: Partial<RequestSettings>) => void
  setSubTab: (tabId: string, subTab: SubTab) => void
  setBulkMode: (tabId: string, table: BulkTable, bulk: boolean) => Promise<void>
  setBulkText: (tabId: string, table: BulkTable, text: string) => void
  chooseBodyFile: (tabId: string, rowId?: string) => Promise<void>
  refreshAutoHeaders: (tabId: string) => Promise<void>
  send: (tabId: string) => Promise<void>
  cancel: (tabId: string) => void
  reset: () => void
}

const timers = new Map<string, ReturnType<typeof setTimeout>>()

function debounce(key: string, action: () => Promise<void>) {
  clearTimeout(timers.get(key))
  timers.set(
    key,
    setTimeout(() => {
      timers.delete(key)
      void action()
    }, SYNC_DELAY_MS),
  )
}

function clearTimers(prefix: string) {
  for (const [key, timer] of timers) {
    if (key.startsWith(prefix)) {
      clearTimeout(timer)
      timers.delete(key)
    }
  }
}

function newTab(): TabDraft {
  return {
    draft: emptyDraft(),
    autoHeaders: [],
    subTab: 'params',
    bulk: {},
    pending: null,
    error: null,
    baseRevision: null,
    conflict: false,
    saving: false,
    dirty: false,
  }
}

export const useRequestBuilderStore = create<RequestBuilderState>()((set, get) => {
  function tab(tabId: string): TabDraft {
    return get().tabs[tabId] ?? newTab()
  }

  function patchTab(tabId: string, update: Partial<TabDraft>) {
    set((state) => ({ tabs: { ...state.tabs, [tabId]: { ...(state.tabs[tabId] ?? newTab()), ...update } } }))
  }

  function patchDraft(tabId: string, update: (draft: RequestDraft) => RequestDraft) {
    patchTab(tabId, { draft: update(tab(tabId).draft), dirty: true })
    debounce(`${tabId}:auto-headers`, () => get().refreshAutoHeaders(tabId))
  }

  function fail(tabId: string, error: unknown) {
    patchTab(tabId, { error: toAppError(error) })
  }

  function syncUrl(tabId: string) {
    debounce(`${tabId}:url`, async () => {
      const sent = tab(tabId).draft
      try {
        const derived = await applyUrl(sent)
        const current = tab(tabId).draft
        if (current.url !== sent.url) {
          return
        }
        patchTab(tabId, {
          draft: {
            ...current,
            params: keepRowIds(current.params, derived.params),
            pathVariables: keepRowIds(current.pathVariables, derived.pathVariables),
          },
          error: null,
        })
      } catch (error) {
        fail(tabId, error)
      }
    })
  }

  function syncParams(tabId: string) {
    debounce(`${tabId}:params`, async () => {
      const sent = tab(tabId).draft
      try {
        const built = await applyParams(sent)
        const current = tab(tabId).draft
        if (current.params !== sent.params || current.pathVariables !== sent.pathVariables) {
          return
        }
        patchTab(tabId, { draft: { ...current, url: built.url }, error: null })
      } catch (error) {
        fail(tabId, error)
      }
    })
  }

  function setRows(tabId: string, table: BulkTable, rows: KeyValueRow[]) {
    patchDraft(tabId, (draft) => withRows(draft, table, rows))
    if (table === 'params') {
      syncParams(tabId)
    }
  }

  return {
    tabs: {},

    ensureTab: (tabId) => {
      if (!get().tabs[tabId]) {
        patchTab(tabId, {})
      }
    },

    openSaved: async (tabId, requestId) => {
      const localId = useWorkspacesStore.getState().activeLocalId
      if (!localId) {
        return false
      }
      try {
        const saved = await getRequest(localId, requestId)
        patchTab(tabId, { ...newTab(), subTab: tab(tabId).subTab, draft: saved.draft, baseRevision: saved.revision })
        void get().refreshAutoHeaders(tabId)
        return true
      } catch (error) {
        fail(tabId, error)
        return false
      }
    },

    loadDraft: (tabId, draft, baseRevision, dirty) => {
      patchTab(tabId, { ...newTab(), draft, baseRevision, dirty })
      void get().refreshAutoHeaders(tabId)
    },

    dropTab: (tabId) => {
      tab(tabId).pending?.cancel()
      clearTimers(`${tabId}:`)
      set((state) => {
        const tabs = { ...state.tabs }
        delete tabs[tabId]
        return { tabs }
      })
    },

    detach: (tabId) =>
      patchTab(tabId, { draft: { ...tab(tabId).draft, id: '' }, baseRevision: null, conflict: false, dirty: true }),

    save: async (tabId, force = false) => {
      const localId = useWorkspacesStore.getState().activeLocalId
      const current = tab(tabId)
      if (!localId || current.draft.id === '' || current.baseRevision === null || current.saving) {
        return { ok: true, value: false }
      }
      patchTab(tabId, { saving: true })
      try {
        const saved = await saveRequest(localId, current.draft, current.baseRevision, force)
        patchTab(tabId, {
          draft: withSavedRows(tab(tabId).draft, saved.draft),
          baseRevision: saved.revision,
          conflict: false,
          error: null,
          dirty: false,
        })
        void useSidebarStore.getState().load()
        return { ok: true, value: true }
      } catch (error) {
        const appError = toAppError(error)
        patchTab(tabId, appError.code === SAVE_CONFLICT_CODE ? { conflict: true } : { error: appError })
        return { ok: false, error: appError }
      } finally {
        patchTab(tabId, { saving: false })
      }
    },

    saveAs: async (tabId, parentId, name) => {
      const localId = useWorkspacesStore.getState().activeLocalId
      if (!localId) {
        return { ok: true, value: null }
      }
      patchTab(tabId, { saving: true })
      try {
        const saved = await saveRequestAs(localId, parentId, name, tab(tabId).draft)
        patchTab(tabId, {
          draft: withSavedRows(tab(tabId).draft, saved.draft),
          baseRevision: saved.revision,
          conflict: false,
          error: null,
          dirty: false,
        })
        void useSidebarStore.getState().load()
        return { ok: true, value: saved.draft.id }
      } catch (error) {
        return { ok: false, error: toAppError(error) }
      } finally {
        patchTab(tabId, { saving: false })
      }
    },

    adoptRename: async (requestId) => {
      const localId = useWorkspacesStore.getState().activeLocalId
      const tabIds = Object.keys(get().tabs).filter((tabId) => tab(tabId).draft.id === requestId)
      if (!localId || tabIds.length === 0) {
        return
      }
      try {
        const saved = await getRequest(localId, requestId)
        for (const tabId of tabIds) {
          patchTab(tabId, { draft: { ...tab(tabId).draft, name: saved.draft.name }, baseRevision: saved.revision })
        }
      } catch (error) {
        for (const tabId of tabIds) {
          fail(tabId, error)
        }
      }
    },

    setName: (tabId, name) => {
      if (tabId !== '') {
        patchDraft(tabId, (draft) => ({ ...draft, name }))
      }
    },

    setDescription: (tabId, description) => patchDraft(tabId, (draft) => ({ ...draft, description })),

    setMethod: (tabId, method) => patchDraft(tabId, (draft) => ({ ...draft, method: method.trim().toUpperCase() })),

    setUrl: (tabId, url) => {
      patchDraft(tabId, (draft) => ({ ...draft, url }))
      syncUrl(tabId)
    },

    updateParams: (tabId, params) => setRows(tabId, 'params', params),

    updatePathVariables: (tabId, pathVariables) => {
      patchDraft(tabId, (draft) => ({ ...draft, pathVariables }))
      syncParams(tabId)
    },

    updateHeaders: (tabId, headers) => setRows(tabId, 'headers', headers),

    setAutoHeaderEnabled: (tabId, key, enabled) =>
      patchDraft(tabId, (draft) => {
        const others = draft.disabledAutoHeaders.filter((name) => name.toLowerCase() !== key.toLowerCase())
        return { ...draft, disabledAutoHeaders: enabled ? others : [...others, key] }
      }),

    updateAuth: (tabId, auth) => patchDraft(tabId, (draft) => ({ ...draft, auth: { ...draft.auth, ...auth } })),

    updateBody: (tabId, body) => patchDraft(tabId, (draft) => ({ ...draft, body: { ...draft.body, ...body } })),

    updateSettings: (tabId, settings) =>
      patchDraft(tabId, (draft) => ({ ...draft, settings: { ...draft.settings, ...settings } })),

    setSubTab: (tabId, subTab) => patchTab(tabId, { subTab }),

    setBulkMode: async (tabId, table, bulk) => {
      const current = tab(tabId)
      if (!bulk) {
        patchTab(tabId, { bulk: { ...current.bulk, [table]: undefined } })
        return
      }
      try {
        const text = await formatBulk(rowsOf(current.draft, table))
        patchTab(tabId, { bulk: { ...tab(tabId).bulk, [table]: text } })
      } catch (error) {
        fail(tabId, error)
      }
    },

    setBulkText: (tabId, table, text) => {
      patchTab(tabId, { bulk: { ...tab(tabId).bulk, [table]: text } })
      debounce(`${tabId}:bulk:${table}`, async () => {
        try {
          const rows = await parseBulk(text)
          if (tab(tabId).bulk[table] === text) {
            setRows(tabId, table, keepRowIds(rowsOf(tab(tabId).draft, table), rows))
          }
        } catch (error) {
          fail(tabId, error)
        }
      })
    },

    chooseBodyFile: async (tabId, rowId) => {
      try {
        const path = await pickFile('Choose a file to send')
        if (path === null) {
          return
        }
        get().updateBody(
          tabId,
          rowId === undefined
            ? { binary: { src: path } }
            : { formdata: tab(tabId).draft.body.formdata.map((row) => (row.id === rowId ? { ...row, src: path } : row)) },
        )
      } catch (error) {
        fail(tabId, error)
      }
    },

    refreshAutoHeaders: async (tabId) => {
      const localId = useWorkspacesStore.getState().activeLocalId
      if (!localId) {
        return
      }
      const draft = tab(tabId).draft
      try {
        const headers = await autoHeaders(localId, draft)
        if (tab(tabId).draft === draft) {
          patchTab(tabId, { autoHeaders: headers })
        }
      } catch (error) {
        fail(tabId, error)
      }
    },

    send: async (tabId) => {
      const localId = useWorkspacesStore.getState().activeLocalId
      if (!localId || tab(tabId).pending) {
        return
      }
      const responses = useResponseViewerStore.getState()
      let pending: PendingSend
      try {
        pending = sendRequest(localId, tabId, tab(tabId).draft)
      } catch (error) {
        responses.setFailure(tabId, toAppError(error))
        return
      }
      patchTab(tabId, { pending })
      responses.startSending(tabId)
      try {
        responses.setResult(tabId, await pending.result)
      } catch (error) {
        const appError = toAppError(error)
        if (appError.cancelled) {
          responses.setResult(tabId, cancelledResult(tabId))
        } else {
          responses.setFailure(tabId, appError)
        }
      } finally {
        patchTab(tabId, { pending: null })
      }
    },

    cancel: (tabId) => tab(tabId).pending?.cancel(),

    reset: () => {
      clearTimers('')
      set({ tabs: {} })
    },
  }
})
