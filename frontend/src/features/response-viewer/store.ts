import { create } from 'zustand'
import { toAppError, type AppError } from '@/platform/errors'
import type { SendResult } from '@/platform/mappers/request'
import { pickSaveFile } from '@/platform/services/dialogs'
import { lastResult, releaseResponse, saveResponseBody } from '@/platform/services/requests'
import { fetchResponseText } from '@/platform/services/response-body'
import { suggestedFileName } from './lib/body-view'
import { EMPTY_ENTRY, withResult } from './lib/entry'
import type { ResponseEntry } from './model'

interface ResponseViewerState {
  entries: Record<string, ResponseEntry>
  startSending: (tabId: string) => void
  /** Ends a send that Go refused before anything went out, keeping the previous response. */
  stopSending: (tabId: string) => void
  setResult: (tabId: string, result: SendResult) => void
  setFailure: (tabId: string, failure: AppError) => void
  rehydrate: (tabId: string) => Promise<void>
  loadBody: (tabId: string) => Promise<void>
  saveBody: (tabId: string) => Promise<void>
  release: (tabId: string) => Promise<void>
  reset: () => void
}

export const useResponseViewerStore = create<ResponseViewerState>()((set, get) => {
  function patch(tabId: string, update: Partial<ResponseEntry>) {
    set((state) => ({ entries: { ...state.entries, [tabId]: { ...(state.entries[tabId] ?? EMPTY_ENTRY), ...update } } }))
  }

  return {
    entries: {},

    startSending: (tabId) => patch(tabId, { sending: true, failure: null }),

    stopSending: (tabId) => patch(tabId, { sending: false }),

    setResult: (tabId, result) =>
      set((state) => ({ entries: { ...state.entries, [tabId]: withResult(state.entries[tabId], result) } })),

    setFailure: (tabId, failure) => patch(tabId, { sending: false, failure }),

    rehydrate: async (tabId) => {
      if (get().entries[tabId]) {
        return
      }
      try {
        const result = await lastResult(tabId)
        if (result && !get().entries[tabId]) {
          get().setResult(tabId, result)
        }
      } catch (error) {
        patch(tabId, { failure: toAppError(error) })
      }
    },

    loadBody: async (tabId) => {
      const entry = get().entries[tabId]
      if (!entry?.result || entry.bodyStatus !== 'idle') {
        return
      }
      const result = entry.result
      patch(tabId, { bodyStatus: 'loading' })
      try {
        const body = await fetchResponseText(result.bodyUrl)
        if (get().entries[tabId]?.result === result) {
          patch(tabId, { body, bodyStatus: 'ready' })
        }
      } catch (error) {
        patch(tabId, { bodyStatus: 'error', failure: toAppError(error) })
      }
    },

    saveBody: async (tabId) => {
      const result = get().entries[tabId]?.result
      if (!result) {
        return
      }
      try {
        const path = await pickSaveFile('Save response body', suggestedFileName(result.contentType))
        if (path !== null) {
          await saveResponseBody(tabId, path)
        }
      } catch (error) {
        patch(tabId, { failure: toAppError(error) })
      }
    },

    release: async (tabId) => {
      set((state) => {
        const entries = { ...state.entries }
        delete entries[tabId]
        return { entries }
      })
      try {
        await releaseResponse(tabId)
      } catch {
        // The tab is already gone, so there is nowhere left to report the failure.
        return
      }
    },

    reset: () => set({ entries: {} }),
  }
})
