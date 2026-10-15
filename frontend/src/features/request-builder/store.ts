import { create } from 'zustand'
import { useResponseViewerStore } from '@/features/response-viewer/store'
import { useSidebarStore } from '@/features/sidebar/store'
import { useWorkspacesStore } from '@/features/workspaces/store'
import { SAVE_CONFLICT_CODE, toAppError, type AppError } from '@/platform/errors'
import type { VariableSuggestion } from '@/platform/mappers/inspection'
import {
  cancelledResult,
  emptyDraft,
  noAcknowledgement,
  type KeyValueRow,
  type PathVariableRow,
  type RequestAuth,
  type RequestBody,
  type RequestDraft,
  type RequestSettings,
  type SavedRequest,
  type ScriptSet,
  type SendAcknowledgement,
} from '@/platform/mappers/request'
import {
  applyParams,
  applyUrl,
  autoHeaders,
  confirmationOf,
  expiredTokenOwnerOf,
  formatBulk,
  getRequest,
  missingSecretsOf,
  parseBulk,
  parseCurl,
  saveRequest,
  saveRequestAs,
  sendRequest,
  type PendingSend,
} from '@/platform/services/requests'
import { createAlias, deleteAlias, saveAlias } from '@/platform/services/aliases'
import { pickFile } from '@/platform/services/dialogs'
import { inspectVariables, suggestVariables } from '@/platform/services/variables'
import { keepRowIds } from '@/shared/lib/rows'
import type { Result } from '@/shared/types/result'
import { withPastedRequest } from './lib/curl'
import { fieldText, withFieldText } from './lib/draft-fields'
import { rowsOf, withRows, withSavedRows } from './lib/draft-rows'
import { DEFAULT_VERSION, isVersionDirty, versionKey } from './lib/versions'
import type { BulkTable, FieldSelection, SubTab, TabDraft, VersionState } from './model'

interface SaveOptions {
  name?: string
}

export const SYNC_DELAY_MS = 150
/** Highlighting and the Variables panel follow typing after this pause (SC-004: within 100 ms of Inspect). */
export const INSPECT_DELAY_MS = 120

interface RequestBuilderState {
  tabs: Record<string, TabDraft>
  /** Bumped by the Focus URL shortcut; the URL bar of the active request focuses itself when it changes. */
  urlFocusTick: number
  focusUrl: () => void
  ensureTab: (tabId: string) => void
  /** Opens the saved request on `aliasId` (default: the alias the tab shows when it already shows this request, else the Default). */
  openSaved: (tabId: string, requestId: string, aliasId?: string | null) => Promise<boolean>
  /** Shows another version of the request in the tab; the version on screen keeps its unsaved edits aside. */
  selectVersion: (tabId: string, aliasId: string | null) => void
  /** Creates an alias as a copy of the saved Default and shows it; returns why it could not, or null. */
  createAlias: (tabId: string, name: string) => Promise<AppError | null>
  /** Deletes an alias and its unsaved edits; a tab showing it goes back to the Default. Returns why it could not, or null. */
  deleteAlias: (tabId: string, aliasId: string) => Promise<AppError | null>
  /** Puts back, after a restart, the alias a tab showed and the drafts of its other versions, then reads its aliases. */
  restoreVersions: (tabId: string, aliasId: string | null, versionDrafts: Record<string, RequestDraft>) => Promise<void>
  /** Gives a duplicated tab the versions of the tab it copies. */
  copyVersions: (fromTabId: string, toTabId: string) => void
  loadDraft: (tabId: string, draft: RequestDraft, baseRevision: string | null, dirty: boolean, savedName?: string | null) => void
  dropTab: (tabId: string) => void
  detach: (tabId: string) => void
  /** Saves the draft; `options.name` replaces the name written (a pending name goes through a rename instead). */
  save: (tabId: string, force?: boolean, options?: SaveOptions) => Promise<Result<boolean, AppError>>
  saveAs: (tabId: string, parentId: string, name: string) => Promise<Result<string | null, AppError>>
  /** Takes the revision of a rename made elsewhere; tabs whose name was still `previousName` take the new name too. */
  adoptRename: (requestId: string, previousName: string) => Promise<void>
  /** A content edit of the name, for a request that was never saved. */
  setName: (tabId: string, name: string) => void
  /** Edits the name draft of a saved request or an example; whether it is pending is derived from the saved name. */
  editName: (tabId: string, name: string) => void
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
  updateScripts: (tabId: string, scripts: ScriptSet) => void
  setInheritScripts: (tabId: string, inheritScripts: boolean) => void
  setSubTab: (tabId: string, subTab: SubTab) => void
  setBulkMode: (tabId: string, table: BulkTable, bulk: boolean) => Promise<void>
  setBulkText: (tabId: string, table: BulkTable, text: string) => void
  /** Replaces the tab's request with a pasted cURL command as one change that can be undone. */
  pasteCurl: (tabId: string, text: string) => Promise<void>
  undoCurlPaste: (tabId: string) => void
  dismissCurlPaste: (tabId: string) => void
  chooseBodyFile: (tabId: string, rowId?: string) => Promise<void>
  refreshAutoHeaders: (tabId: string) => Promise<void>
  send: (tabId: string, acknowledgement?: Partial<SendAcknowledgement>) => Promise<void>
  dismissMissingSecrets: (tabId: string) => void
  dismissConfirmation: (tabId: string) => void
  dismissExpiredToken: (tabId: string) => void
  inspect: (tabId: string) => Promise<void>
  refreshInspections: () => void
  suggest: (tabId: string, prefix: string) => Promise<VariableSuggestion[]>
  setSelection: (tabId: string, selection: FieldSelection | null) => void
  replaceSelection: (tabId: string, text: string) => boolean
  cancel: (tabId: string) => void
  reset: () => void
}

const timers = new Map<string, ReturnType<typeof setTimeout>>()

function debounce(key: string, action: () => Promise<void>, delay = SYNC_DELAY_MS) {
  clearTimeout(timers.get(key))
  timers.set(
    key,
    setTimeout(() => {
      timers.delete(key)
      void action()
    }, delay),
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
    missingSecrets: null,
    confirmation: null,
    expiredTokenOwner: null,
    acknowledgement: noAcknowledgement(),
    inspection: null,
    selection: null,
    error: null,
    curlPaste: null,
    baseRevision: null,
    savedName: null,
    conflict: false,
    saving: false,
    dirty: false,
    aliasId: null,
    aliases: [],
    savedDefault: null,
    versions: {},
  }
}

/** What the tab knows about the request on disk after Go returned it. */
function fromDisk(saved: SavedRequest): Pick<TabDraft, 'baseRevision' | 'aliases' | 'savedDefault'> {
  return { baseRevision: saved.revision, aliases: saved.aliases, savedDefault: saved.draft }
}

/** A version as it is on disk; null for an alias the request no longer has or a Default never read. */
function savedVersion(current: TabDraft, aliasId: string | null): VersionState | null {
  if (aliasId === null) {
    return current.savedDefault === null ? null : { draft: current.savedDefault, savedName: current.savedDefault.name, dirty: false }
  }
  const alias = current.aliases.find((candidate) => candidate.id === aliasId)
  return alias ? { draft: alias.draft, savedName: alias.name, dirty: false } : null
}

/** The versions with the Default kept aside taking a rename made elsewhere, as the Default on screen would. */
function renamedDefault(versions: Record<string, VersionState>, previousName: string, name: string): Record<string, VersionState> {
  const stashed = versions[DEFAULT_VERSION]
  if (!stashed) {
    return versions
  }
  const draft = stashed.draft.name.trim() === previousName ? { ...stashed.draft, name } : stashed.draft
  return { ...versions, [DEFAULT_VERSION]: { ...stashed, draft, savedName: name } }
}

/** The versions kept aside that still exist on disk, without the one on screen. */
function keptVersions(versions: Record<string, VersionState>, saved: SavedRequest, shown: string | null): Record<string, VersionState> {
  return Object.fromEntries(
    Object.entries(versions).filter(
      ([key]) => key !== versionKey(shown) && (key === DEFAULT_VERSION || saved.aliases.some((alias) => alias.id === key)),
    ),
  )
}

export const useRequestBuilderStore = create<RequestBuilderState>()((set, get) => {
  function tab(tabId: string): TabDraft {
    return get().tabs[tabId] ?? newTab()
  }

  function patchTab(tabId: string, update: Partial<TabDraft>) {
    set((state) => ({ tabs: { ...state.tabs, [tabId]: { ...(state.tabs[tabId] ?? newTab()), ...update } } }))
    if (update.draft !== undefined) {
      debounce(`${tabId}:inspect`, () => get().inspect(tabId), INSPECT_DELAY_MS)
    }
  }

  function patchDraft(tabId: string, update: (draft: RequestDraft) => RequestDraft) {
    patchTab(tabId, { draft: update(tab(tabId).draft), dirty: true, curlPaste: null })
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

  /** Saves the alias on screen, its name included; the Default and the other aliases stay as they are on disk. */
  async function saveShownAlias(tabId: string, localId: string, aliasId: string, force: boolean): Promise<Result<boolean, AppError>> {
    try {
      const current = tab(tabId)
      const saved = await saveAlias(localId, aliasId, current.draft, current.baseRevision ?? '', force)
      const stored = saved.aliases.find((alias) => alias.id === aliasId)
      const latest = tab(tabId)
      patchTab(tabId, {
        ...fromDisk(saved),
        draft: stored ? { ...withSavedRows(latest.draft, stored.draft), name: latest.draft.name } : latest.draft,
        savedName: stored?.name ?? latest.savedName,
        conflict: false,
        error: null,
        dirty: false,
      })
      return { ok: true, value: true }
    } catch (error) {
      const appError = toAppError(error)
      patchTab(tabId, appError.code === SAVE_CONFLICT_CODE ? { conflict: true } : { error: appError })
      return { ok: false, error: appError }
    }
  }

  function setRows(tabId: string, table: BulkTable, rows: KeyValueRow[]) {
    patchDraft(tabId, (draft) => withRows(draft, table, rows))
    if (table === 'params') {
      syncParams(tabId)
    }
  }

  return {
    urlFocusTick: 0,
    tabs: {},

    ensureTab: (tabId) => {
      if (!get().tabs[tabId]) {
        patchTab(tabId, {})
      }
    },

    openSaved: async (tabId, requestId, aliasId) => {
      const localId = useWorkspacesStore.getState().activeLocalId
      if (!localId) {
        return false
      }
      try {
        const saved = await getRequest(localId, requestId)
        const current = tab(tabId)
        const same = current.draft.id === requestId
        const wanted = aliasId !== undefined ? aliasId : same ? current.aliasId : null
        const shown = saved.aliases.find((alias) => alias.id === wanted) ?? null
        patchTab(tabId, {
          ...newTab(),
          ...fromDisk(saved),
          subTab: current.subTab,
          draft: shown ? shown.draft : saved.draft,
          savedName: shown ? shown.name : saved.draft.name,
          aliasId: shown?.id ?? null,
          versions: same ? keptVersions(current.versions, saved, shown?.id ?? null) : {},
        })
        void get().refreshAutoHeaders(tabId)
        return true
      } catch (error) {
        fail(tabId, error)
        return false
      }
    },

    selectVersion: (tabId, aliasId) => {
      const current = tab(tabId)
      if (current.aliasId === aliasId) {
        return
      }
      const versions = { ...current.versions, [versionKey(current.aliasId)]: { draft: current.draft, savedName: current.savedName, dirty: current.dirty } }
      const target = versions[versionKey(aliasId)] ?? savedVersion(current, aliasId)
      if (!target) {
        return
      }
      delete versions[versionKey(aliasId)]
      patchTab(tabId, { aliasId, draft: target.draft, savedName: target.savedName, dirty: target.dirty, versions, curlPaste: null, selection: null, error: null, conflict: false })
      void get().refreshAutoHeaders(tabId)
    },

    createAlias: async (tabId, name) => {
      const localId = useWorkspacesStore.getState().activeLocalId
      const current = tab(tabId)
      if (!localId || current.draft.id === '') {
        return null
      }
      try {
        const saved = await createAlias(localId, current.draft.id, name, current.baseRevision ?? '')
        patchTab(tabId, fromDisk(saved))
        const created = saved.aliases.at(-1)
        if (created) {
          get().selectVersion(tabId, created.id)
        }
        return null
      } catch (error) {
        const appError = toAppError(error)
        if (appError.code === SAVE_CONFLICT_CODE) {
          patchTab(tabId, { conflict: true })
        }
        return appError
      }
    },

    deleteAlias: async (tabId, aliasId) => {
      const localId = useWorkspacesStore.getState().activeLocalId
      const current = tab(tabId)
      if (!localId || current.draft.id === '') {
        return null
      }
      try {
        const saved = await deleteAlias(localId, current.draft.id, aliasId, current.baseRevision ?? '')
        const latest = tab(tabId)
        const versions = { ...latest.versions }
        delete versions[aliasId]
        if (latest.aliasId !== aliasId) {
          patchTab(tabId, { ...fromDisk(saved), versions })
          return null
        }
        const fallback = versions[DEFAULT_VERSION] ?? { draft: saved.draft, savedName: saved.draft.name, dirty: false }
        delete versions[DEFAULT_VERSION]
        patchTab(tabId, { ...fromDisk(saved), aliasId: null, draft: fallback.draft, savedName: fallback.savedName, dirty: fallback.dirty, versions, error: null, conflict: false })
        void get().refreshAutoHeaders(tabId)
        return null
      } catch (error) {
        const appError = toAppError(error)
        if (appError.code === SAVE_CONFLICT_CODE) {
          patchTab(tabId, { conflict: true })
        }
        return appError
      }
    },

    restoreVersions: async (tabId, aliasId, versionDrafts) => {
      const restored = Object.fromEntries(
        Object.entries(versionDrafts).map(([key, draft]) => [key, { draft, savedName: null, dirty: true }]),
      )
      patchTab(tabId, { aliasId, versions: { ...tab(tabId).versions, ...restored } })
      const localId = useWorkspacesStore.getState().activeLocalId
      const requestId = tab(tabId).draft.id
      if (!localId || requestId === '') {
        return
      }
      try {
        const saved = await getRequest(localId, requestId)
        const current = tab(tabId)
        const known = current.aliasId === null || saved.aliases.some((alias) => alias.id === current.aliasId)
        patchTab(tabId, { aliases: saved.aliases, savedDefault: saved.draft, versions: keptVersions(current.versions, saved, known ? current.aliasId : null) })
        if (!known) {
          const fallback = current.versions[DEFAULT_VERSION] ?? { draft: saved.draft, savedName: saved.draft.name, dirty: false }
          patchTab(tabId, { aliasId: null, draft: fallback.draft, savedName: fallback.savedName, dirty: fallback.dirty })
        }
      } catch (error) {
        fail(tabId, error)
      }
    },

    copyVersions: (fromTabId, toTabId) => {
      const source = tab(fromTabId)
      patchTab(toTabId, { aliasId: source.aliasId, aliases: source.aliases, savedDefault: source.savedDefault, versions: structuredClone(source.versions) })
    },

    loadDraft: (tabId, draft, baseRevision, dirty, savedName = null) => {
      patchTab(tabId, { ...newTab(), draft, baseRevision, dirty, savedName })
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
      patchTab(tabId, {
        draft: { ...tab(tabId).draft, id: '' },
        baseRevision: null,
        savedName: null,
        conflict: false,
        dirty: true,
        aliasId: null,
        aliases: [],
        savedDefault: null,
        versions: {},
      }),

    save: async (tabId, force = false, options = {}) => {
      const localId = useWorkspacesStore.getState().activeLocalId
      const current = tab(tabId)
      if (!localId || current.draft.id === '' || current.baseRevision === null || current.saving) {
        return { ok: true, value: false }
      }
      patchTab(tabId, { saving: true })
      try {
        if (current.aliasId !== null) {
          return await saveShownAlias(tabId, localId, current.aliasId, force)
        }
        const sent = options.name === undefined ? current.draft : { ...current.draft, name: options.name }
        const saved = await saveRequest(localId, sent, current.baseRevision, force)
        const draftName = tab(tabId).draft.name
        patchTab(tabId, {
          ...fromDisk(saved),
          draft: { ...withSavedRows(tab(tabId).draft, saved.draft), name: options.name === undefined ? saved.draft.name : draftName },
          savedName: saved.draft.name,
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
          savedName: saved.draft.name,
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

    adoptRename: async (requestId, previousName) => {
      const localId = useWorkspacesStore.getState().activeLocalId
      const tabIds = Object.keys(get().tabs).filter((tabId) => tab(tabId).draft.id === requestId)
      if (!localId || tabIds.length === 0) {
        return
      }
      try {
        const saved = await getRequest(localId, requestId)
        for (const tabId of tabIds) {
          const current = tab(tabId)
          if (current.aliasId !== null) {
            patchTab(tabId, { ...fromDisk(saved), versions: renamedDefault(current.versions, previousName, saved.draft.name) })
            continue
          }
          const draft = current.draft
          const keepsOwnName = draft.name.trim() !== previousName
          patchTab(tabId, { ...fromDisk(saved), draft: keepsOwnName ? draft : { ...draft, name: saved.draft.name }, savedName: saved.draft.name })
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

    editName: (tabId, name) => {
      if (tabId !== '') {
        patchTab(tabId, { draft: { ...tab(tabId).draft, name } })
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

    updateScripts: (tabId, scripts) => patchDraft(tabId, (draft) => ({ ...draft, scripts })),

    setInheritScripts: (tabId, inheritScripts) => patchDraft(tabId, (draft) => ({ ...draft, inheritScripts })),

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

    pasteCurl: async (tabId, text) => {
      try {
        const parsed = await parseCurl(text)
        const before = tab(tabId)
        patchDraft(tabId, (draft) => withPastedRequest(draft, parsed.draft))
        patchTab(tabId, { curlPaste: { previous: before.draft, wasDirty: before.dirty, notices: parsed.notices }, error: null })
      } catch (error) {
        fail(tabId, error)
      }
    },

    undoCurlPaste: (tabId) => {
      const paste = tab(tabId).curlPaste
      if (!paste) {
        return
      }
      patchTab(tabId, { draft: paste.previous, dirty: paste.wasDirty, curlPaste: null })
      void get().refreshAutoHeaders(tabId)
    },

    dismissCurlPaste: (tabId) => patchTab(tabId, { curlPaste: null }),

    chooseBodyFile: async (tabId, rowId) => {
      try {
        const path = await pickFile('Choose a file to send')
        if (path === null) {
          return
        }
        get().updateBody(
          tabId,
          rowId === undefined
            ? { binary: { src: path, srcState: 'ok' } }
            : { formdata: tab(tabId).draft.body.formdata.map((row) => (row.id === rowId ? { ...row, src: path, srcState: 'ok' } : row)) },
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

    send: async (tabId, acknowledgement = {}) => {
      const localId = useWorkspacesStore.getState().activeLocalId
      if (!localId || tab(tabId).pending) {
        return
      }
      const acknowledged = { ...noAcknowledgement(), ...acknowledgement }
      const responses = useResponseViewerStore.getState()
      let pending: PendingSend
      try {
        pending = sendRequest(localId, tabId, tab(tabId).draft, acknowledged, tab(tabId).aliasId)
      } catch (error) {
        responses.setFailure(tabId, toAppError(error))
        return
      }
      patchTab(tabId, { pending, missingSecrets: null, confirmation: null, expiredTokenOwner: null, acknowledgement: acknowledged })
      responses.startSending(tabId)
      try {
        responses.setResult(tabId, await pending.result)
      } catch (error) {
        const appError = toAppError(error)
        const missingSecrets = missingSecretsOf(appError)
        const confirmation = confirmationOf(appError)
        const expiredTokenOwner = expiredTokenOwnerOf(appError)
        if (expiredTokenOwner !== null) {
          responses.stopSending(tabId)
          patchTab(tabId, { expiredTokenOwner })
        } else if (confirmation) {
          responses.stopSending(tabId)
          patchTab(tabId, { confirmation })
        } else if (missingSecrets) {
          responses.stopSending(tabId)
          patchTab(tabId, { missingSecrets })
        } else if (appError.cancelled) {
          responses.setResult(tabId, cancelledResult(tabId))
        } else {
          responses.setFailure(tabId, appError)
        }
      } finally {
        patchTab(tabId, { pending: null })
      }
    },

    dismissMissingSecrets: (tabId) => patchTab(tabId, { missingSecrets: null }),

    dismissConfirmation: (tabId) => patchTab(tabId, { confirmation: null }),

    dismissExpiredToken: (tabId) => patchTab(tabId, { expiredTokenOwner: null }),

    inspect: async (tabId) => {
      const localId = useWorkspacesStore.getState().activeLocalId
      if (!localId || !get().tabs[tabId]) {
        return
      }
      const draft = tab(tabId).draft
      try {
        const inspection = await inspectVariables(localId, draft)
        if (get().tabs[tabId] && tab(tabId).draft === draft && useWorkspacesStore.getState().activeLocalId === localId) {
          set((state) => ({ tabs: { ...state.tabs, [tabId]: { ...tab(tabId), inspection } } }))
        }
      } catch {
        // Highlighting is a hint: without an answer the editors keep the last decorations.
        return
      }
    },

    refreshInspections: () => {
      for (const tabId of Object.keys(get().tabs)) {
        void get().inspect(tabId)
      }
    },

    suggest: async (tabId, prefix) => {
      const localId = useWorkspacesStore.getState().activeLocalId
      if (!localId) {
        return []
      }
      try {
        return await suggestVariables(localId, tab(tabId).draft.id, prefix)
      } catch {
        return []
      }
    },

    setSelection: (tabId, selection) => {
      if (get().tabs[tabId]) {
        set((state) => ({ tabs: { ...state.tabs, [tabId]: { ...tab(tabId), selection } } }))
      }
    },

    replaceSelection: (tabId, text) => {
      const { draft, selection } = tab(tabId)
      const current = selection === null ? null : fieldText(draft, selection.field)
      if (selection === null || current === null || current.slice(selection.from, selection.to) !== selection.text) {
        return false
      }
      const replaced = current.slice(0, selection.from) + text + current.slice(selection.to)
      patchDraft(tabId, (previous) => withFieldText(previous, selection.field, replaced))
      patchTab(tabId, { selection: null })
      if (selection.field === 'url') {
        syncUrl(tabId)
      } else if (selection.field.startsWith('params.') || selection.field.startsWith('path_variables.')) {
        syncParams(tabId)
      }
      return true
    },

    cancel: (tabId) => tab(tabId).pending?.cancel(),

    focusUrl: () => set((state) => ({ urlFocusTick: state.urlFocusTick + 1 })),

    reset: () => {
      clearTimers('')
      set({ tabs: {} })
    },
  }
})

/** The versions a tab keeps aside with unsaved edits, Default first then aliases in file order (null is the Default). */
export function selectDirtyHiddenVersions(state: Pick<RequestBuilderState, 'tabs'>, tabId: string): (string | null)[] {
  const current = state.tabs[tabId]
  if (!current) {
    return []
  }
  return [null, ...current.aliases.map((alias) => alias.id)].filter((aliasId) => {
    const version = current.versions[versionKey(aliasId)]
    return version !== undefined && isVersionDirty(version)
  })
}
