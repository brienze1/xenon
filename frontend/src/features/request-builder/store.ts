import { create } from 'zustand'
import { useResponseViewerStore } from '@/features/response-viewer/store'
import { useWorkspacesStore } from '@/features/workspaces/store'
import { toAppError, type AppError } from '@/platform/errors'
import {
  cancelledResult,
  emptyDraft,
  type AutoHeader,
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
  parseBulk,
  sendRequest,
  type PendingSend,
} from '@/platform/services/requests'
import { pickFile } from '@/platform/services/dialogs'
import { keepRowIds } from './lib/rows'

export const SYNC_DELAY_MS = 150

export type SubTab = 'params' | 'auth' | 'headers' | 'body' | 'scripts' | 'settings'
export type BulkTable = 'params' | 'headers' | 'urlencoded'

export interface TabDraft {
  draft: RequestDraft
  autoHeaders: AutoHeader[]
  subTab: SubTab
  bulk: Partial<Record<BulkTable, string>>
  pending: PendingSend | null
  error: AppError | null
}

interface RequestBuilderState {
  activeTabId: string
  tabs: Record<string, TabDraft>
  openTab: (tabId?: string) => string
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

function debounce(key: string, action: () => void) {
  clearTimeout(timers.get(key))
  timers.set(
    key,
    setTimeout(() => {
      timers.delete(key)
      action()
    }, SYNC_DELAY_MS),
  )
}

function newTabId(): string {
  return `tab-${globalThis.crypto.randomUUID()}`
}

function newTab(): TabDraft {
  return { draft: emptyDraft(), autoHeaders: [], subTab: 'params', bulk: {}, pending: null, error: null }
}

function rowsOf(draft: RequestDraft, table: BulkTable): KeyValueRow[] {
  return table === 'urlencoded' ? draft.body.urlencoded : draft[table]
}

function withRows(draft: RequestDraft, table: BulkTable, rows: KeyValueRow[]): RequestDraft {
  return table === 'urlencoded' ? { ...draft, body: { ...draft.body, urlencoded: rows } } : { ...draft, [table]: rows }
}

export const useRequestBuilderStore = create<RequestBuilderState>()((set, get) => {
  function tab(tabId: string): TabDraft {
    return get().tabs[tabId] ?? newTab()
  }

  function patchTab(tabId: string, update: Partial<TabDraft>) {
    set((state) => ({ tabs: { ...state.tabs, [tabId]: { ...(state.tabs[tabId] ?? newTab()), ...update } } }))
  }

  function patchDraft(tabId: string, update: (draft: RequestDraft) => RequestDraft) {
    patchTab(tabId, { draft: update(tab(tabId).draft) })
    debounce(`${tabId}:auto-headers`, () => void get().refreshAutoHeaders(tabId))
  }

  function fail(tabId: string, error: unknown) {
    patchTab(tabId, { error: toAppError(error) })
  }

  function syncUrl(tabId: string) {
    debounce(`${tabId}:url`, () => {
      const sent = tab(tabId).draft
      applyUrl(sent).then(
        (derived) => {
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
        },
        (error) => fail(tabId, error),
      )
    })
  }

  function syncParams(tabId: string) {
    debounce(`${tabId}:params`, () => {
      const sent = tab(tabId).draft
      applyParams(sent).then(
        (built) => {
          const current = tab(tabId).draft
          if (current.params !== sent.params || current.pathVariables !== sent.pathVariables) {
            return
          }
          patchTab(tabId, { draft: { ...current, url: built.url }, error: null })
        },
        (error) => fail(tabId, error),
      )
    })
  }

  function setRows(tabId: string, table: BulkTable, rows: KeyValueRow[]) {
    patchDraft(tabId, (draft) => withRows(draft, table, rows))
    if (table === 'params') {
      syncParams(tabId)
    }
  }

  const initialTabId = newTabId()

  return {
    activeTabId: initialTabId,
    tabs: { [initialTabId]: newTab() },

    openTab: (tabId = newTabId()) => {
      set((state) => ({ activeTabId: tabId, tabs: { ...state.tabs, [tabId]: state.tabs[tabId] ?? newTab() } }))
      return tabId
    },

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
      debounce(`${tabId}:bulk:${table}`, () => {
        parseBulk(text).then(
          (rows) => {
            if (tab(tabId).bulk[table] === text) {
              setRows(tabId, table, keepRowIds(rowsOf(tab(tabId).draft, table), rows))
            }
          },
          (error) => fail(tabId, error),
        )
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
      const pending = sendRequest(localId, tabId, tab(tabId).draft)
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
      for (const timer of timers.values()) {
        clearTimeout(timer)
      }
      timers.clear()
      const tabId = newTabId()
      set({ activeTabId: tabId, tabs: { [tabId]: newTab() } })
    },
  }
})
