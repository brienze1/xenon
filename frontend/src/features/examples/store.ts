import { create } from 'zustand'
import { useRequestBuilderStore } from '@/features/request-builder/store'
import { useSidebarStore } from '@/features/sidebar/store'
import { useWorkspacesStore } from '@/features/workspaces/store'
import { SAVE_CONFLICT_CODE, toAppError, type AppError } from '@/platform/errors'
import type { Example, ExampleResponse } from '@/platform/mappers/example'
import type { KeyValueRow } from '@/platform/mappers/request'
import { createExampleFromResponse, getExample, saveExample } from '@/platform/services/examples'
import { fetchResponseText } from '@/platform/services/response-body'
import type { Result } from '@/shared/types/result'
import { headersToText, textToHeaders } from './lib/headers'
import type { ExampleTabState, SaveExampleDialogState } from './model'

interface ExamplesState {
  tabs: Record<string, ExampleTabState>
  saveDialog: SaveExampleDialogState | null
  open: (tabId: string, requestId: string, exampleId: string) => Promise<boolean>
  drop: (tabId: string) => void
  isDirty: (tabId: string) => boolean
  setCode: (tabId: string, code: number) => void
  setStatusText: (tabId: string, status: string) => void
  setHeaders: (tabId: string, headers: KeyValueRow[]) => void
  setBulkHeaders: (tabId: string, bulk: boolean) => void
  setBulkHeadersText: (tabId: string, text: string) => void
  setBody: (tabId: string, body: string) => void
  loadRoutedBody: (tabId: string) => Promise<void>
  save: (tabId: string, force?: boolean) => Promise<Result<boolean, AppError>>
  /** Takes a rename made elsewhere; a tab whose name was still `previousName` takes the new name too. */
  adoptRename: (exampleId: string, previousName: string) => Promise<void>
  openSaveDialog: (tabId: string, requestId: string) => void
  closeSaveDialog: () => void
  /** Creates an example from the dialog's tab response; resolves to the new example id. */
  createFromResponse: (name: string) => Promise<Result<{ requestId: string; exampleId: string }, AppError>>
  reset: () => void
}

function activeLocalId(): string | null {
  return useWorkspacesStore.getState().activeLocalId
}

function builder() {
  return useRequestBuilderStore.getState()
}

function newTab(requestId: string, exampleId: string): ExampleTabState {
  return {
    requestId,
    exampleId,
    example: null,
    response: null,
    bulkHeaders: null,
    routedBody: null,
    bodyStatus: 'idle',
    status: 'loading',
    dirty: false,
    saving: false,
    conflict: false,
    error: null,
  }
}

export const useExamplesStore = create<ExamplesState>()((set, get) => {
  function patchTab(tabId: string, update: Partial<ExampleTabState>) {
    set((state) => {
      const current = state.tabs[tabId]
      return current ? { tabs: { ...state.tabs, [tabId]: { ...current, ...update } } } : state
    })
  }

  function editResponse(tabId: string, update: (response: ExampleResponse) => ExampleResponse) {
    const response = get().tabs[tabId]?.response
    if (response) {
      patchTab(tabId, { response: update(response), dirty: true })
    }
  }

  function loaded(tabId: string, example: Example) {
    builder().loadDraft(tabId, { ...example.request, name: example.name }, null, false)
    patchTab(tabId, {
      example,
      response: example.response,
      bulkHeaders: null,
      routedBody: null,
      bodyStatus: 'idle',
      status: 'ready',
      dirty: false,
      conflict: false,
      error: null,
    })
  }

  return {
    tabs: {},
    saveDialog: null,

    open: async (tabId, requestId, exampleId) => {
      const localId = activeLocalId()
      set((state) => ({ tabs: { ...state.tabs, [tabId]: newTab(requestId, exampleId) } }))
      builder().ensureTab(tabId)
      if (!localId) {
        return false
      }
      try {
        loaded(tabId, await getExample(localId, requestId, exampleId))
        return true
      } catch (error) {
        patchTab(tabId, { status: 'error', error: toAppError(error) })
        return false
      }
    },

    drop: (tabId) =>
      set((state) => {
        if (!state.tabs[tabId]) {
          return state
        }
        const tabs = { ...state.tabs }
        delete tabs[tabId]
        return { tabs }
      }),

    isDirty: (tabId) => (get().tabs[tabId]?.dirty ?? false) || (builder().tabs[tabId]?.dirty ?? false),

    setCode: (tabId, code) => editResponse(tabId, (response) => ({ ...response, code })),

    setStatusText: (tabId, status) => editResponse(tabId, (response) => ({ ...response, status })),

    setHeaders: (tabId, headers) => editResponse(tabId, (response) => ({ ...response, headers })),

    setBulkHeaders: (tabId, bulk) => {
      const tab = get().tabs[tabId]
      if (!tab?.response) {
        return
      }
      if (bulk) {
        patchTab(tabId, { bulkHeaders: headersToText(tab.response.headers) })
        return
      }
      const text = tab.bulkHeaders
      patchTab(tabId, { bulkHeaders: null })
      if (text !== null) {
        editResponse(tabId, (response) => ({ ...response, headers: textToHeaders(text) }))
      }
    },

    setBulkHeadersText: (tabId, text) => {
      patchTab(tabId, { bulkHeaders: text })
      editResponse(tabId, (response) => ({ ...response, headers: textToHeaders(text) }))
    },

    setBody: (tabId, body) => editResponse(tabId, (response) => (response.body === null ? response : { ...response, body })),

    loadRoutedBody: async (tabId) => {
      const tab = get().tabs[tabId]
      const response = tab?.response
      if (!response || response.body !== null || response.isBinary || tab.bodyStatus !== 'idle') {
        return
      }
      patchTab(tabId, { bodyStatus: 'loading' })
      try {
        const body = await fetchResponseText(response.bodyUrl)
        patchTab(tabId, { routedBody: body, bodyStatus: 'ready' })
      } catch (error) {
        patchTab(tabId, { bodyStatus: 'error', error: toAppError(error) })
      }
    },

    save: async (tabId, force = false) => {
      const localId = activeLocalId()
      const tab = get().tabs[tabId]
      if (!localId || !tab?.example || !tab.response || tab.saving) {
        return { ok: true, value: false }
      }
      const request = builder().tabs[tabId]?.draft ?? tab.example.request
      patchTab(tabId, { saving: true })
      try {
        const saved = await saveExample(localId, tab.requestId, { ...tab.example, request, response: tab.response }, tab.example.revision, force)
        const draftName = builder().tabs[tabId]?.draft.name ?? saved.name
        loaded(tabId, saved)
        builder().editName(tabId, draftName)
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

    adoptRename: async (exampleId, previousName) => {
      const localId = activeLocalId()
      const renaming = Object.entries(get().tabs).filter(([, tab]) => tab.exampleId === exampleId)
      if (!localId) {
        return
      }
      await Promise.all(
        renaming.map(async ([tabId, tab]) => {
          try {
            const renamed = await getExample(localId, tab.requestId, exampleId)
            const draftName = builder().tabs[tabId]?.draft.name ?? renamed.name
            const keepsOwnName = draftName.trim() !== previousName
            if (get().isDirty(tabId) || keepsOwnName) {
              patchTab(tabId, { example: tab.example ? { ...tab.example, name: renamed.name, revision: renamed.revision } : renamed })
              builder().editName(tabId, keepsOwnName ? draftName : renamed.name)
            } else {
              loaded(tabId, renamed)
            }
          } catch (error) {
            patchTab(tabId, { error: toAppError(error) })
          }
        }),
      )
    },

    openSaveDialog: (tabId, requestId) => set({ saveDialog: { tabId, requestId } }),

    closeSaveDialog: () => set({ saveDialog: null }),

    createFromResponse: async (name) => {
      const localId = activeLocalId()
      const dialog = get().saveDialog
      if (!localId || !dialog) {
        return { ok: false, error: toAppError(new Error('No response to save')) }
      }
      try {
        const created = await createExampleFromResponse(localId, dialog.requestId, dialog.tabId, name)
        set({ saveDialog: null })
        const sidebar = useSidebarStore.getState()
        await sidebar.load()
        if (!useSidebarStore.getState().expanded[dialog.requestId]) {
          sidebar.toggle(dialog.requestId)
        }
        return { ok: true, value: { requestId: dialog.requestId, exampleId: created.id } }
      } catch (error) {
        return { ok: false, error: toAppError(error) }
      }
    },

    reset: () => set({ tabs: {}, saveDialog: null }),
  }
})
