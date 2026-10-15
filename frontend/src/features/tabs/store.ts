import { create } from 'zustand'
import { isDocumentationDirty, useDocumentationStore } from '@/features/documentation/store'
import { findNameClash, isScopeDirty, useEnvironmentsStore } from '@/features/environments/store'
import { useExamplesStore } from '@/features/examples/store'
import { isItemDirty, useItemSettingsStore } from '@/features/item-settings/store'
import { useRequestBuilderStore } from '@/features/request-builder/store'
import { isSettingsDirty, useSettingsStore } from '@/features/settings/store'
import { useResponseViewerStore } from '@/features/response-viewer/store'
import { useSidebarStore } from '@/features/sidebar/store'
import { useWorkspacesStore } from '@/features/workspaces/store'
import { AppError } from '@/platform/errors'
import type { RequestChangedOnDisk } from '@/platform/events'
import { emptyDraft, type RequestDraft } from '@/platform/mappers/request'
import type { TreeNode } from '@/platform/mappers/tree'
import type { ClosedTabState, SnippetOptionsState, TabState, UIState } from '@/platform/mappers/ui-state'
import { collectionScope, environmentScope, GLOBALS_SCOPE } from '@/platform/mappers/variable'
import type { Result } from '@/shared/types/result'
import { containerOptions } from './lib/containers'
import { saveUIState } from '@/platform/services/ui-state'
import { sameDraftContent } from './lib/draft'
import { withoutNotice } from './lib/notices'
import { latestClosedTabs } from './lib/closed'
import { isNamePending, nameErrorMessage, type NameError } from './lib/name-draft'
import { namedTreeId, tabNameOf, treeExampleName, treeNodeName, type TabName } from './lib/tab-name'
import { moveTab, neighbourOf, positionAt, tabsAfter, tabsBefore, wrappedNeighbour } from './lib/order'
import type { CloseChoice, EnvironmentTabTarget, ItemTabTarget, OpenMode, SaveDialogState, TabMeta, TabNotice } from './model'

export const PERSIST_DELAY_MS = 300

/** The code of the error a save returns when the tab's name is empty or taken (shown in the close prompt). */
export const INVALID_NAME = 'TAB-01400'

interface TabsState {
  localId: string | null
  tabs: TabMeta[]
  activeTabId: string | null
  previewTabId: string | null
  closePrompt: string | null
  closeError: AppError | null
  saveDialog: SaveDialogState | null
  notices: Record<string, TabNotice>
  /** Dirty tabs still to ask about after the current close prompt (Close others/all/left/right). */
  closeQueue: string[]
  /** Tab ids from the least to the most recently active, so a sidebar click finds the latest tab of a request. */
  recent: string[]
  /** The last closed tabs with their drafts, newest last (at most 20), for "Reopen closed tab". */
  closedTabs: ClosedTabState[]
  /** The options last used per snippet language, the last entry being the language last used. */
  snippetOptions: SnippetOptionsState[]
  /** Why a tab's pending name was not saved, by tab id; cleared when the name is edited again. */
  nameErrors: Record<string, NameError>
  hydrate: (localId: string, state: UIState) => void
  newTab: () => string
  openDraft: (draft: RequestDraft) => string
  activate: (tabId: string) => void
  openPreview: (requestId: string) => Promise<void>
  openPinned: (requestId: string) => Promise<void>
  openExamplePreview: (requestId: string, exampleId: string) => Promise<void>
  openExamplePinned: (requestId: string, exampleId: string) => Promise<void>
  handleExampleDeleted: (exampleId: string) => void
  /** Opens (or focuses) the tab editing an environment's variables or the globals. */
  openEnvironmentTab: (target: EnvironmentTabTarget, mode: OpenMode) => void
  /** Opens (or focuses) the single, pinned Settings tab; with no workspace, the full-page Settings. */
  openSettingsTab: () => void
  /** Closes the tabs of environments that no longer exist. */
  handleEnvironmentsChanged: (environmentIds: string[]) => void
  /** Opens (or focuses) the tab editing a collection's or folder's settings. */
  openItemTab: (target: ItemTabTarget, mode: OpenMode) => void
  /** Follows the workspace tree: closes the tabs of collections and folders that no longer exist. */
  handleTreeChanged: (tree: TreeNode[]) => void
  pin: (tabId: string) => void
  pinIfEdited: () => void
  reorder: (fromId: string, toId: string) => void
  requestClose: (tabId?: string) => void
  resolveClose: (choice: CloseChoice) => Promise<void>
  /** Closes the tab, keeping it in the closed-tab stack unless `remember` is false. */
  closeTab: (tabId: string, remember?: boolean) => void
  forceClose: (tabId?: string) => void
  closeOthers: (tabId: string) => void
  closeAll: () => void
  closeLeft: (tabId: string) => void
  closeRight: (tabId: string) => void
  forceCloseAll: () => void
  /** Opens a second tab on the same request (or a copy of an unsaved one) with a copy of the draft; null for example tabs. */
  duplicateTab: (tabId: string) => string | null
  revealInSidebar: (tabId: string) => void
  reopenClosed: () => Promise<void>
  /** Activates the tab at a 1-based position; 9 is always the last tab. */
  activateIndex: (position: number) => void
  cycle: (step: 1 | -1) => void
  setSnippetOptions: (language: string, options: Record<string, string>) => void
  /** Edits the name of the item the tab shows, as a draft the tab saves with its other changes. */
  editTabName: (tabId: string, name: string) => void
  /** Drops the tab's name draft (Escape): the tab shows the saved name again. */
  cancelTabName: (tabId: string) => void
  /** Saves the tab: the name first checked, then the other changes, then the rename; any failure stops it. */
  save: (tabId?: string) => Promise<void>
  saveAs: (tabId?: string) => void
  confirmSave: (parentId: string, name: string) => Promise<AppError | null>
  cancelSave: () => void
  handleDeleted: (requestIds: string[]) => void
  handleChangedOnDisk: (event: RequestChangedOnDisk) => Promise<void>
  keepAsNew: (tabId: string) => void
  keepMine: (tabId: string) => Promise<void>
  loadFromDisk: (tabId: string) => Promise<void>
  snapshot: () => UIState
  schedulePersist: () => void
  flushPersist: () => Promise<void>
  reset: () => void
}

let persistTimer: ReturnType<typeof setTimeout> | undefined

function newTabId(): string {
  return `tab-${globalThis.crypto.randomUUID()}`
}

function builder() {
  return useRequestBuilderStore.getState()
}

function examples() {
  return useExamplesStore.getState()
}

function environments() {
  return useEnvironmentsStore.getState()
}

function isExampleTab(tab: TabMeta | undefined): boolean {
  return tab?.kind === 'example'
}

/** Whether the tab edits variables (an environment or the globals) rather than a request. */
export function isEnvironmentTab(tab: TabMeta | undefined): boolean {
  return tab?.kind === 'environment' || tab?.kind === 'globals'
}

/** The variable table scope an environment tab edits; null for request and example tabs. */
export function environmentTabScope(tab: TabMeta | undefined): string | null {
  if (tab?.kind === 'globals') {
    return GLOBALS_SCOPE
  }
  return tab?.kind === 'environment' && tab.environmentId ? environmentScope(tab.environmentId) : null
}

/** Whether the tab edits a collection's or folder's settings. */
export function isItemTab(tab: TabMeta | undefined): boolean {
  return (tab?.kind === 'collection' || tab?.kind === 'folder') && Boolean(tab.itemId)
}

/** Whether the tab edits the app settings (the single Settings tab). */
export function isAppSettingsTab(tab: TabMeta | undefined): boolean {
  return tab?.kind === 'settings'
}

/** Whether the tab has no request draft (environment, globals, collection, folder and Settings tabs). */
function hasNoRequestDraft(tab: TabMeta | undefined): boolean {
  return isEnvironmentTab(tab) || isItemTab(tab) || isAppSettingsTab(tab)
}

/** A request tab never saved: saving it asks where to put it (the Save dialog). */
function needsSaveDialog(tab: TabMeta): boolean {
  return !hasNoRequestDraft(tab) && !isExampleTab(tab) && tab.requestId === null
}

function appSettings() {
  return useSettingsStore.getState()
}

function settingsTabMeta(id: string): TabMeta {
  return { id, requestId: null, pinned: true, kind: 'settings', exampleId: null, environmentId: null, itemId: null }
}

function items() {
  return useItemSettingsStore.getState()
}

function documentation() {
  return useDocumentationStore.getState()
}

/** Whether the tab is a collection or folder tab whose documentation YAML has unsaved edits. */
function isDocumentationTabDirty(tab: TabMeta | undefined): boolean {
  return isItemTab(tab) && Boolean(tab!.itemId) && isDocumentationDirty(documentation(), tab!.itemId!)
}

function itemTabMeta(id: string, target: ItemTabTarget, pinned: boolean): TabMeta {
  return { id, requestId: null, pinned, kind: target.kind, exampleId: null, environmentId: null, itemId: target.itemId }
}

/** The collection variables shown in a collection tab; null for every other tab. */
function collectionVariablesScope(tab: TabMeta | undefined): string | null {
  return tab?.kind === 'collection' && tab.itemId ? collectionScope(tab.itemId) : null
}

function targetOf(tab: TabMeta): EnvironmentTabTarget | null {
  if (tab.kind === 'globals') {
    return { kind: 'globals' }
  }
  return tab.kind === 'environment' && tab.environmentId ? { kind: 'environment', environmentId: tab.environmentId } : null
}

function isTargetOf(tab: TabMeta, target: EnvironmentTabTarget): boolean {
  return target.kind === 'globals' ? tab.kind === 'globals' : tab.kind === 'environment' && tab.environmentId === target.environmentId
}

function environmentTabMeta(id: string, target: EnvironmentTabTarget, pinned: boolean): TabMeta {
  return { id, requestId: null, pinned, kind: target.kind, exampleId: null, environmentId: target.kind === 'environment' ? target.environmentId : null }
}

/** The name the tab edits, read from the stores that own its draft and its saved name. */
function tabName(tab: TabMeta | undefined): TabName | null {
  if (!tab) {
    return null
  }
  const tree = useSidebarStore.getState().tree
  const treeId = namedTreeId(tab)
  const entry = tab.itemId ? items().items[tab.itemId] : undefined
  const environmentId = tab.kind === 'environment' ? tab.environmentId : null
  return tabNameOf(tab, {
    builderName: builder().tabs[tab.id]?.draft.name,
    builderSavedName: builder().tabs[tab.id]?.savedName ?? null,
    treeName: treeId === null ? null : isExampleTab(tab) ? treeExampleName(tree, treeId) : treeNodeName(tree, treeId),
    itemDraft: entry?.draft?.name,
    itemBase: entry?.base?.name ?? null,
    exampleName: examples().tabs[tab.id]?.example?.name ?? null,
    environmentDraft: environmentId ? environments().nameDrafts[environmentId] : undefined,
    environmentName: environments().environments.find((environment) => environment.id === environmentId)?.name ?? null,
  })
}

function isTabNamePending(tab: TabMeta | undefined): boolean {
  const name = tabName(tab)
  return name !== null && isNamePending(name.draft, name.saved)
}

/** The message of the tab's name error (empty, taken, or what the rename returned); null when there is none. */
export function selectNameErrorText(state: Pick<TabsState, 'nameErrors'>, tabId: string | null): string | null {
  const error = tabId === null ? undefined : state.nameErrors[tabId]
  return error ? nameErrorMessage(error) : null
}

/** The tab showing a collection or folder (one tab per item). */
export function selectItemTabId(state: Pick<TabsState, 'tabs'>, itemId: string): string | null {
  return state.tabs.find((tab) => isItemTab(tab) && tab.itemId === itemId)?.id ?? null
}

/** The tab showing an environment (one tab per environment). */
export function selectEnvironmentTabId(state: Pick<TabsState, 'tabs'>, environmentId: string): string | null {
  return state.tabs.find((tab) => tab.kind === 'environment' && tab.environmentId === environmentId)?.id ?? null
}

/** Whether the tab has unsaved changes, its name included: closing it asks first. */
export function isTabDirty(tabId: string): boolean {
  const tab = useTabsStore.getState().tabs.find((candidate) => candidate.id === tabId)
  return isContentDirty(tab, tabId) || isTabNamePending(tab)
}

function isContentDirty(tab: TabMeta | undefined, tabId: string): boolean {
  const scope = environmentTabScope(tab)
  if (scope !== null) {
    return isScopeDirty(environments(), scope)
  }
  if (isAppSettingsTab(tab)) {
    return isSettingsDirty(appSettings())
  }
  if (isItemTab(tab)) {
    const variables = collectionVariablesScope(tab)
    return isItemDirty(items(), tab!.itemId!) || (variables !== null && isScopeDirty(environments(), variables)) || isDocumentationTabDirty(tab)
  }
  return (builder().tabs[tabId]?.dirty ?? false) || examples().isDirty(tabId)
}

/** What stops a pending name from being saved, checked before anything of the tab is written. */
function nameProblem(tab: TabMeta, name: TabName, wanted: string): NameError | null {
  if (wanted === '') {
    return { kind: 'required', subject: name.subject }
  }
  if (tab.kind === 'environment' && tab.environmentId && findNameClash(environments(), tab.environmentId, wanted) !== null) {
    return { kind: 'clash', name: name.draft ?? wanted }
  }
  return null
}

function isPristine(draft: RequestDraft): boolean {
  return sameDraftContent(draft, emptyDraft())
}

function cancelPersist(): void {
  clearTimeout(persistTimer)
  persistTimer = undefined
}

function toTabMeta(tab: TabState): TabMeta {
  if (tab.kind === 'settings') {
    return settingsTabMeta(tab.tabId)
  }
  if ((tab.kind === 'collection' || tab.kind === 'folder') && tab.itemId) {
    return itemTabMeta(tab.tabId, { kind: tab.kind, itemId: tab.itemId }, tab.pinned)
  }
  if (tab.kind === 'environment' || tab.kind === 'globals') {
    return { id: tab.tabId, requestId: null, pinned: tab.pinned, kind: tab.kind, exampleId: null, environmentId: tab.environmentId ?? null }
  }
  if (tab.kind === 'example') {
    return { id: tab.tabId, requestId: tab.requestId, pinned: tab.pinned, kind: 'example', exampleId: tab.exampleId ?? null }
  }
  return { id: tab.tabId, requestId: tab.requestId, pinned: tab.pinned }
}

const initialState = {
  localId: null as string | null,
  tabs: [] as TabMeta[],
  activeTabId: null as string | null,
  previewTabId: null as string | null,
  closePrompt: null as string | null,
  closeError: null as AppError | null,
  saveDialog: null as SaveDialogState | null,
  notices: {} as Record<string, TabNotice>,
  closeQueue: [] as string[],
  recent: [] as string[],
  closedTabs: [] as ClosedTabState[],
  snippetOptions: [] as SnippetOptionsState[],
  nameErrors: {} as Record<string, NameError>,
}

export const useTabsStore = create<TabsState>()((set, get) => {
  function patchTab(tabId: string, update: Partial<TabMeta>) {
    set((state) => ({ tabs: state.tabs.map((tab) => (tab.id === tabId ? { ...tab, ...update } : tab)) }))
  }

  function tabOf(requestId: string): TabMeta | undefined {
    const matching = get().tabs.filter((tab) => tab.requestId === requestId && !isExampleTab(tab))
    const recent = get().recent
    return matching.reduce<TabMeta | undefined>((latest, tab) => (latest === undefined || recent.indexOf(tab.id) > recent.indexOf(latest.id) ? tab : latest), undefined)
  }

  function focus(tabId: string | null) {
    set((state) => ({ activeTabId: tabId, recent: tabId === null ? state.recent : [...state.recent.filter((id) => id !== tabId), tabId] }))
  }

  function exampleTabOf(exampleId: string): TabMeta | undefined {
    return get().tabs.find((tab) => isExampleTab(tab) && tab.exampleId === exampleId)
  }

  function tabById(tabId: string | null): TabMeta | undefined {
    return get().tabs.find((tab) => tab.id === tabId)
  }

  function addTab(meta: TabMeta, after?: string) {
    if (!hasNoRequestDraft(meta)) {
      builder().ensureTab(meta.id)
    }
    set((state) => {
      const index = after === undefined ? -1 : state.tabs.findIndex((tab) => tab.id === after)
      const tabs = index < 0 ? [...state.tabs, meta] : [...state.tabs.slice(0, index + 1), meta, ...state.tabs.slice(index + 1)]
      return { tabs }
    })
    focus(meta.id)
  }

  function remember(tab: TabMeta) {
    const snapshot = tabSnapshot(tab)
    if (!hasNoRequestDraft(tab) && snapshot.requestId === null && (snapshot.draft === null || isPristine(snapshot.draft))) {
      return
    }
    set((state) => ({ closedTabs: latestClosedTabs([...state.closedTabs, { closedAt: new Date().toISOString(), tab: snapshot }]) }))
  }

  function closeMany(tabIds: string[], force: boolean) {
    const dirtyIds: string[] = []
    for (const tabId of tabIds) {
      if (!force && isTabDirty(tabId)) {
        dirtyIds.push(tabId)
        continue
      }
      get().closeTab(tabId)
    }
    const [first, ...rest] = dirtyIds
    set({ closePrompt: first ?? null, closeQueue: rest, closeError: null })
  }

  function promptNext() {
    const [next, ...rest] = get().closeQueue
    set({ closePrompt: next ?? null, closeQueue: rest, closeError: null })
  }

  async function syncSiblings(tabId: string) {
    const savedTab = tabById(tabId)
    const revision = builder().tabs[tabId]?.baseRevision
    if (!savedTab?.requestId || isExampleTab(savedTab) || !revision) {
      return
    }
    const reloads: Promise<unknown>[] = []
    for (const sibling of get().tabs) {
      if (sibling.id === tabId || sibling.requestId !== savedTab.requestId || isExampleTab(sibling) || builder().tabs[sibling.id]?.baseRevision === revision) {
        continue
      }
      if (isTabDirty(sibling.id)) {
        set((state) => ({ notices: { ...state.notices, [sibling.id]: 'changed' } }))
      } else {
        reloads.push(builder().openSaved(sibling.id, savedTab.requestId))
      }
    }
    await Promise.all(reloads)
  }

  /**
   * Saves a collection or folder tab: its settings, its documentation and, for a collection, its variables.
   * Edited documentation first shows what it changes; the save is `false` until that is confirmed.
   */
  async function saveItemTab(tab: TabMeta): Promise<Result<boolean, AppError>> {
    const itemId = tab.itemId!
    const settings = await items().save(itemId)
    if (!settings.ok) {
      return settings
    }
    const variables = collectionVariablesScope(tab)
    if (variables !== null && isScopeDirty(environments(), variables)) {
      const saved = await environments().saveDraft(variables)
      if (!saved.ok || !saved.value) {
        items().setSection(itemId, 'variables')
        return saved
      }
    }
    if (!isDocumentationTabDirty(tab)) {
      return settings
    }
    const documented = await documentation().save(itemId)
    if (!documented.ok || !documented.value) {
      items().setSection(itemId, 'documentation')
    }
    return documented
  }

  /** Tabs whose save is running; a second save of the same tab waits for nothing and writes nothing. */
  const saving = new Set<string>()

  function setNameError(tabId: string, error: NameError | null) {
    set((state) => {
      if (error === null && !(tabId in state.nameErrors)) {
        return state
      }
      const nameErrors = { ...state.nameErrors }
      if (error === null) {
        delete nameErrors[tabId]
      } else {
        nameErrors[tabId] = error
      }
      return { nameErrors }
    })
  }

  /** Saves everything the tab edits except its name; `savedName` is what a request must keep writing as its name. */
  async function saveContent(tab: TabMeta, namePending: boolean, savedName: string | null): Promise<Result<boolean, AppError>> {
    const scope = environmentTabScope(tab)
    if (scope !== null) {
      return environments().saveDraft(scope)
    }
    if (isItemTab(tab)) {
      return saveItemTab(tab)
    }
    if (isAppSettingsTab(tab)) {
      return appSettings().save()
    }
    if (namePending && !isContentDirty(tab, tab.id)) {
      return { ok: true, value: true }
    }
    if (isExampleTab(tab)) {
      return examples().save(tab.id)
    }
    const saved = await builder().save(tab.id, false, savedName === null ? {} : { name: savedName })
    if (saved.ok && saved.value) {
      await syncSiblings(tab.id)
    }
    return saved
  }

  /** Writes the tab's new name through the item's rename operation and lets every view take it. */
  async function rename(tab: TabMeta, wanted: string, previous: string): Promise<Result<boolean, AppError>> {
    if (tab.kind === 'environment' && tab.environmentId) {
      const failure = await environments().renameEnvironment(tab.environmentId, wanted)
      if (failure !== null) {
        return { ok: false, error: failure }
      }
      if (environments().nameDrafts[tab.environmentId]?.trim() === wanted) {
        environments().dropName(tab.environmentId)
      }
      return { ok: true, value: true }
    }
    const sidebar = useSidebarStore.getState()
    const treeId = namedTreeId(tab)!
    const renamed = isExampleTab(tab) ? await sidebar.renameExample(treeId, wanted) : await sidebar.rename(treeId, wanted)
    if (!renamed.ok) {
      return renamed
    }
    if (isItemTab(tab)) {
      if (items().items[treeId]?.draft?.name.trim() === wanted) {
        items().editName(treeId, wanted)
      }
      await items().load(treeId)
      return { ok: true, value: true }
    }
    if (builder().tabs[tab.id]?.draft.name.trim() === wanted) {
      builder().editName(tab.id, wanted)
    }
    if (isExampleTab(tab)) {
      await examples().adoptRename(treeId, previous)
    } else {
      await builder().adoptRename(treeId, previous)
    }
    return { ok: true, value: true }
  }

  /** The save order of every tab: check the name, save the other changes, then rename; any failure stops it. */
  async function saveTabChanges(tab: TabMeta): Promise<Result<boolean, AppError>> {
    if (saving.has(tab.id)) {
      return { ok: true, value: false }
    }
    saving.add(tab.id)
    try {
      const name = tabName(tab)
      const pending = name !== null && isNamePending(name.draft, name.saved)
      const wanted = pending ? name!.draft!.trim() : ''
      const problem = pending ? nameProblem(tab, name!, wanted) : null
      if (problem !== null) {
        setNameError(tab.id, problem)
        return { ok: false, error: new AppError(INVALID_NAME, nameErrorMessage(problem)) }
      }
      const content = await saveContent(tab, pending, name?.saved ?? null)
      if (!content.ok || !content.value || !pending) {
        return content
      }
      const renamed = await rename(tab, wanted, name!.saved!)
      if (!renamed.ok) {
        setNameError(tab.id, { kind: 'failed', error: renamed.error })
        return renamed
      }
      setNameError(tab.id, null)
      return { ok: true, value: true }
    } finally {
      saving.delete(tab.id)
    }
  }

  /** Clears what a preview tab showed before it is reused for something else. */
  function releasePreview(preview: string) {
    const replaced = tabById(preview)
    if (isItemTab(replaced)) {
      dropItem(replaced!)
    } else if (!isEnvironmentTab(replaced)) {
      builder().dropTab(preview)
      examples().drop(preview)
      void useResponseViewerStore.getState().release(preview)
    }
  }

  /** Forgets the edits of a collection or folder tab (one tab per item). */
  function dropItem(tab: TabMeta) {
    const variables = collectionVariablesScope(tab)
    if (variables !== null) {
      environments().discardDraft(variables)
    }
    items().drop(tab.itemId!)
    documentation().drop(tab.itemId!)
  }

  function activeOr(tabId: string | undefined): string | null {
    return tabId ?? get().activeTabId
  }

  function sidebarSnapshot(): UIState['sidebar'] {
    const sidebar = useSidebarStore.getState()
    return {
      expanded: Object.entries(sidebar.expanded).flatMap(([id, open]) => (open ? [id] : [])),
      selectedId: sidebar.selectedId,
      view: sidebar.view,
      search: sidebar.query,
      historySource: sidebar.historySource,
    }
  }

  function tabSnapshot(tab: TabMeta): TabState {
    if (isAppSettingsTab(tab)) {
      return { tabId: tab.id, kind: 'settings', requestId: null, pinned: true, baseRevision: null, draft: null }
    }
    if (isItemTab(tab)) {
      return { tabId: tab.id, kind: tab.kind, requestId: null, itemId: tab.itemId ?? null, pinned: tab.pinned, baseRevision: null, draft: null }
    }
    if (isEnvironmentTab(tab)) {
      return { tabId: tab.id, kind: tab.kind, requestId: null, environmentId: tab.environmentId ?? null, pinned: tab.pinned, baseRevision: null, draft: null }
    }
    if (isExampleTab(tab)) {
      return { tabId: tab.id, kind: 'example', requestId: tab.requestId, exampleId: tab.exampleId ?? null, pinned: tab.pinned, baseRevision: null, draft: null }
    }
    const draft = builder().tabs[tab.id]
    const keepDraft = tab.requestId === null || (draft?.dirty ?? false) || isTabNamePending(tab)
    return {
      tabId: tab.id,
      requestId: tab.requestId,
      pinned: tab.pinned,
      baseRevision: draft?.baseRevision ?? null,
      draft: keepDraft && draft ? draft.draft : null,
    }
  }

  return {
    ...initialState,

    hydrate: (localId, state) => {
      set({
        ...initialState,
        localId,
        tabs: state.tabs.map(toTabMeta),
        activeTabId: state.activeTabId,
        previewTabId: state.previewTabId,
        recent: state.activeTabId ? [state.activeTabId] : [],
        closedTabs: latestClosedTabs(state.closedTabs),
        snippetOptions: state.snippetOptions,
      })
      useSidebarStore.getState().applyState({
        expanded: Object.fromEntries(state.sidebar.expanded.map((id) => [id, true])),
        selectedId: state.sidebar.selectedId,
        query: state.sidebar.search,
        view: state.sidebar.view,
        historySource: state.sidebar.historySource,
      })
      for (const tab of state.tabs) {
        if (tab.kind === 'environment' || tab.kind === 'globals' || tab.kind === 'collection' || tab.kind === 'folder' || tab.kind === 'settings') {
          continue
        }
        if (tab.kind === 'example' && tab.requestId && tab.exampleId) {
          void examples().open(tab.tabId, tab.requestId, tab.exampleId)
        } else if (tab.draft) {
          builder().loadDraft(tab.tabId, tab.draft, tab.baseRevision, tab.requestId !== null || !isPristine(tab.draft))
        } else if (tab.requestId) {
          builder().ensureTab(tab.tabId)
          void builder().openSaved(tab.tabId, tab.requestId)
        } else {
          builder().ensureTab(tab.tabId)
        }
      }
      if (state.tabs.length === 0) {
        get().newTab()
      } else {
        const activeTabId = get().activeTabId
        if (activeTabId === null || !get().tabs.some((tab) => tab.id === activeTabId)) {
          set({ activeTabId: state.tabs[0]!.tabId })
        }
      }
    },

    newTab: () => {
      const id = newTabId()
      addTab({ id, requestId: null, pinned: true })
      return id
    },

    openDraft: (draft) => {
      const id = newTabId()
      addTab({ id, requestId: null, pinned: true })
      builder().loadDraft(id, draft, null, true)
      return id
    },

    activate: (tabId) => focus(tabId),

    openPreview: async (requestId) => {
      const existing = tabOf(requestId)
      if (existing) {
        focus(existing.id)
        return
      }
      const preview = get().previewTabId
      if (preview !== null && !isTabDirty(preview)) {
        const replaced = tabById(preview)
        if (isItemTab(replaced)) {
          dropItem(replaced!)
        }
        patchTab(preview, { requestId, kind: undefined, exampleId: undefined, environmentId: undefined, itemId: undefined })
        examples().drop(preview)
        set({ notices: withoutNotice(get().notices, preview) })
        focus(preview)
        void useResponseViewerStore.getState().release(preview)
        await builder().openSaved(preview, requestId)
        return
      }
      const id = newTabId()
      addTab({ id, requestId, pinned: false })
      set({ previewTabId: id })
      await builder().openSaved(id, requestId)
    },

    openPinned: async (requestId) => {
      const existing = tabOf(requestId)
      if (existing) {
        get().pin(existing.id)
        focus(existing.id)
        return
      }
      const id = newTabId()
      addTab({ id, requestId, pinned: true })
      await builder().openSaved(id, requestId)
    },

    openExamplePreview: async (requestId, exampleId) => {
      const existing = exampleTabOf(exampleId)
      if (existing) {
        focus(existing.id)
        return
      }
      const preview = get().previewTabId
      if (preview !== null && !isTabDirty(preview)) {
        const replaced = tabById(preview)
        if (isItemTab(replaced)) {
          dropItem(replaced!)
        }
        patchTab(preview, { requestId, kind: 'example', exampleId, environmentId: undefined, itemId: undefined })
        set({ notices: withoutNotice(get().notices, preview) })
        focus(preview)
        void useResponseViewerStore.getState().release(preview)
        await examples().open(preview, requestId, exampleId)
        return
      }
      const id = newTabId()
      addTab({ id, requestId, pinned: false, kind: 'example', exampleId })
      set({ previewTabId: id })
      await examples().open(id, requestId, exampleId)
    },

    openExamplePinned: async (requestId, exampleId) => {
      const existing = exampleTabOf(exampleId)
      if (existing) {
        get().pin(existing.id)
        focus(existing.id)
        return
      }
      const id = newTabId()
      addTab({ id, requestId, pinned: true, kind: 'example', exampleId })
      await examples().open(id, requestId, exampleId)
    },

    openEnvironmentTab: (target, mode) => {
      const existing = get().tabs.find((tab) => isTargetOf(tab, target))
      if (existing) {
        if (mode === 'pinned') {
          get().pin(existing.id)
        }
        focus(existing.id)
        return
      }
      const preview = get().previewTabId
      if (mode === 'preview' && preview !== null && !isTabDirty(preview)) {
        releasePreview(preview)
        patchTab(preview, environmentTabMeta(preview, target, false))
        set({ notices: withoutNotice(get().notices, preview) })
        focus(preview)
        return
      }
      const id = newTabId()
      addTab(environmentTabMeta(id, target, mode === 'pinned'))
      if (mode === 'preview') {
        set({ previewTabId: id })
      }
    },

    openSettingsTab: () => {
      if (get().localId === null) {
        appSettings().openPage()
        return
      }
      const existing = get().tabs.find(isAppSettingsTab)
      if (existing) {
        focus(existing.id)
        return
      }
      addTab(settingsTabMeta(newTabId()))
    },

    handleEnvironmentsChanged: (environmentIds) => {
      const existing = new Set(environmentIds)
      for (const tab of get().tabs) {
        if (tab.kind === 'environment' && (!tab.environmentId || !existing.has(tab.environmentId))) {
          get().closeTab(tab.id, false)
        }
      }
    },

    openItemTab: (target, mode) => {
      const existing = get().tabs.find((tab) => isItemTab(tab) && tab.itemId === target.itemId)
      if (existing) {
        if (existing.kind !== target.kind) {
          patchTab(existing.id, { kind: target.kind })
        }
        if (mode === 'pinned') {
          get().pin(existing.id)
        }
        focus(existing.id)
        return
      }
      const preview = get().previewTabId
      if (mode === 'preview' && preview !== null && !isTabDirty(preview)) {
        releasePreview(preview)
        patchTab(preview, itemTabMeta(preview, target, false))
        set({ notices: withoutNotice(get().notices, preview) })
        focus(preview)
        return
      }
      const id = newTabId()
      addTab(itemTabMeta(id, target, mode === 'pinned'))
      if (mode === 'preview') {
        set({ previewTabId: id })
      }
    },

    handleTreeChanged: (tree) => {
      const containers = new Map(containerOptions(tree).map((container) => [container.id, container.kind]))
      for (const tab of get().tabs) {
        if (!isItemTab(tab)) {
          continue
        }
        const kind = containers.get(tab.itemId!)
        if (kind === undefined) {
          get().closeTab(tab.id, false)
        } else if (kind !== tab.kind) {
          patchTab(tab.id, { kind })
        }
      }
    },

    handleExampleDeleted: (exampleId) => {
      const tab = exampleTabOf(exampleId)
      if (tab) {
        get().closeTab(tab.id, false)
      }
    },

    pin: (tabId) => {
      patchTab(tabId, { pinned: true })
      if (get().previewTabId === tabId) {
        set({ previewTabId: null })
      }
    },

    pinIfEdited: () => {
      const preview = get().previewTabId
      if (preview !== null && isTabDirty(preview)) {
        get().pin(preview)
      }
    },

    reorder: (fromId, toId) => set((state) => ({ tabs: moveTab(state.tabs, fromId, toId) })),

    requestClose: (tabId) => {
      const target = activeOr(tabId)
      if (target === null) {
        return
      }
      if (isTabDirty(target)) {
        set({ closePrompt: target, closeError: null })
        return
      }
      get().closeTab(target)
    },

    resolveClose: async (choice) => {
      const tabId = get().closePrompt
      if (choice === 'cancel') {
        set({ closePrompt: null, closeQueue: [], closeError: null })
        return
      }
      if (tabId === null || choice === 'discard') {
        if (tabId !== null) {
          get().closeTab(tabId)
        }
        promptNext()
        return
      }
      const tab = tabById(tabId)
      if (!tab) {
        promptNext()
        return
      }
      if (needsSaveDialog(tab)) {
        set({ closePrompt: null, closeError: null, saveDialog: { tabId, mode: 'save', closeAfter: true } })
        return
      }
      set({ closeError: null })
      const saved = await saveTabChanges(tab)
      if (!saved.ok) {
        set((state) => (state.closePrompt === tabId ? { closeError: saved.error } : state))
        return
      }
      if (!saved.value) {
        set({ closePrompt: null, closeQueue: [] })
        focus(tabId)
        return
      }
      set({ closePrompt: null })
      get().closeTab(tabId)
      promptNext()
    },

    closeTab: (tabId, rememberTab = true) => {
      const closing = tabById(tabId)
      if (closing && rememberTab) {
        remember(closing)
      }
      set((state) => ({
        tabs: state.tabs.filter((tab) => tab.id !== tabId),
        activeTabId: state.activeTabId === tabId ? neighbourOf(state.tabs, tabId) : state.activeTabId,
        previewTabId: state.previewTabId === tabId ? null : state.previewTabId,
        notices: withoutNotice(state.notices, tabId),
        recent: state.recent.filter((id) => id !== tabId),
      }))
      setNameError(tabId, null)
      const scope = environmentTabScope(closing)
      if (scope !== null) {
        environments().discardDraft(scope)
        if (closing?.environmentId) {
          environments().dropName(closing.environmentId)
        }
        return
      }
      if (isItemTab(closing)) {
        dropItem(closing!)
        return
      }
      if (isAppSettingsTab(closing)) {
        appSettings().discard()
        return
      }
      builder().dropTab(tabId)
      examples().drop(tabId)
      void useResponseViewerStore.getState().release(tabId)
    },

    forceClose: (tabId) => {
      const target = activeOr(tabId)
      if (target !== null) {
        get().closeTab(target)
      }
    },

    closeOthers: (tabId) => closeMany(get().tabs.flatMap((tab) => (tab.id === tabId ? [] : [tab.id])), false),

    closeAll: () => closeMany(get().tabs.map((tab) => tab.id), false),

    closeLeft: (tabId) => closeMany(tabsBefore(get().tabs, tabId), false),

    closeRight: (tabId) => closeMany(tabsAfter(get().tabs, tabId), false),

    forceCloseAll: () => closeMany(get().tabs.map((tab) => tab.id), true),

    duplicateTab: (tabId) => {
      const source = tabById(tabId)
      const draft = builder().tabs[tabId]
      if (!source || isExampleTab(source) || !draft) {
        return null
      }
      const id = newTabId()
      addTab({ id, requestId: source.requestId, pinned: true }, tabId)
      builder().loadDraft(id, structuredClone(draft.draft), draft.baseRevision, draft.dirty, draft.savedName)
      return id
    },

    revealInSidebar: (tabId) => {
      const tab = tabById(tabId)
      const sidebar = useSidebarStore.getState()
      const treeId = isItemTab(tab) ? tab!.itemId! : tab?.requestId
      if (isEnvironmentTab(tab)) {
        sidebar.setView('environments')
      } else if (treeId) {
        if (sidebar.view !== 'collections') {
          sidebar.setView('collections')
        }
        sidebar.reveal(treeId)
      } else {
        return
      }
      if (!useWorkspacesStore.getState().layout.sidebarVisible) {
        useWorkspacesStore.getState().setLayout({ sidebarVisible: true })
      }
    },

    reopenClosed: async () => {
      const closed = get().closedTabs.at(-1)
      if (!closed) {
        return
      }
      set((state) => ({ closedTabs: state.closedTabs.slice(0, -1) }))
      const { tab } = closed
      const meta = toTabMeta(tab)
      if (isAppSettingsTab(meta)) {
        get().openSettingsTab()
        return
      }
      if (isItemTab(meta)) {
        const container = containerOptions(useSidebarStore.getState().tree).find((candidate) => candidate.id === meta.itemId)
        if (container) {
          get().openItemTab({ kind: container.kind, itemId: container.id }, 'pinned')
        }
        return
      }
      const target = targetOf(meta)
      if (target !== null) {
        if (target.kind === 'globals' || environments().environments.some((environment) => environment.id === target.environmentId)) {
          get().openEnvironmentTab(target, 'pinned')
        }
        return
      }
      if (tab.kind === 'example' && tab.requestId && tab.exampleId) {
        await get().openExamplePinned(tab.requestId, tab.exampleId)
        return
      }
      const id = newTabId()
      addTab({ id, requestId: tab.requestId, pinned: true })
      if (tab.draft) {
        builder().loadDraft(id, tab.draft, tab.baseRevision, true)
      } else if (tab.requestId) {
        await builder().openSaved(id, tab.requestId)
      }
    },

    activateIndex: (position) => {
      const target = positionAt(get().tabs, position)
      if (target) {
        focus(target.id)
      }
    },

    cycle: (step) => {
      const target = wrappedNeighbour(get().tabs, get().activeTabId, step)
      if (target) {
        focus(target.id)
      }
    },

    setSnippetOptions: (language, options) =>
      set((state) => ({ snippetOptions: [...state.snippetOptions.filter((saved) => saved.language !== language), { language, options }] })),

    save: async (tabId) => {
      const tab = tabById(activeOr(tabId))
      if (!tab) {
        return
      }
      if (needsSaveDialog(tab)) {
        set({ saveDialog: { tabId: tab.id, mode: 'save', closeAfter: false } })
        return
      }
      await saveTabChanges(tab)
    },

    editTabName: (tabId, name) => {
      const tab = tabById(tabId)
      if (!tab) {
        return
      }
      if (isItemTab(tab)) {
        items().editName(tab.itemId!, name)
      } else if (tab.kind === 'environment' && tab.environmentId) {
        environments().editName(tab.environmentId, name)
      } else if (isExampleTab(tab) || tab.requestId !== null) {
        builder().editName(tabId, name)
      } else {
        builder().setName(tabId, name)
      }
      setNameError(tabId, null)
    },

    cancelTabName: (tabId) => {
      const tab = tabById(tabId)
      const saved = tabName(tab)?.saved ?? null
      setNameError(tabId, null)
      if (!tab || saved === null) {
        return
      }
      if (isItemTab(tab)) {
        items().editName(tab.itemId!, saved)
      } else if (tab.kind === 'environment' && tab.environmentId) {
        environments().dropName(tab.environmentId)
      } else {
        builder().editName(tabId, saved)
      }
    },

    saveAs: (tabId) => {
      const target = activeOr(tabId)
      const tab = tabById(target)
      if (target !== null && !isExampleTab(tab) && !hasNoRequestDraft(tab)) {
        set({ saveDialog: { tabId: target, mode: 'saveAs', closeAfter: false } })
      }
    },

    confirmSave: async (parentId, name) => {
      const dialog = get().saveDialog
      if (dialog === null) {
        return null
      }
      const saved = await builder().saveAs(dialog.tabId, parentId, name)
      if (!saved.ok) {
        return saved.error
      }
      const requestId = saved.value
      if (requestId === null) {
        return null
      }
      set({ saveDialog: null, notices: withoutNotice(get().notices, dialog.tabId) })
      get().pin(dialog.tabId)
      patchTab(dialog.tabId, { requestId })
      if (dialog.closeAfter) {
        get().closeTab(dialog.tabId)
      }
      return null
    },

    cancelSave: () => set({ saveDialog: null }),

    handleDeleted: (requestIds) => {
      const deleted = new Set(requestIds)
      for (const tab of get().tabs) {
        if (tab.requestId === null || !deleted.has(tab.requestId)) {
          continue
        }
        if (isExampleTab(tab)) {
          get().closeTab(tab.id, false)
        } else if (isTabDirty(tab.id)) {
          set((state) => ({ notices: { ...state.notices, [tab.id]: 'deleted' } }))
        } else {
          get().closeTab(tab.id, false)
        }
      }
    },

    handleChangedOnDisk: async (event) => {
      if (event.localId !== get().localId) {
        return
      }
      if (event.deleted) {
        get().handleDeleted([event.requestId])
        return
      }
      const reloads: Promise<unknown>[] = []
      const drafts = builder().tabs
      for (const tab of get().tabs) {
        const draft = drafts[tab.id]
        if (tab.requestId !== event.requestId) {
          continue
        }
        if (isExampleTab(tab)) {
          const loaded = examples().tabs[tab.id]
          if (loaded?.example?.revision !== event.revision && !isTabDirty(tab.id) && tab.exampleId) {
            reloads.push(examples().open(tab.id, event.requestId, tab.exampleId))
          }
          continue
        }
        if (draft?.baseRevision === event.revision) {
          continue
        }
        if (isTabDirty(tab.id)) {
          set((state) => ({ notices: { ...state.notices, [tab.id]: 'changed' } }))
        } else {
          reloads.push(builder().openSaved(tab.id, event.requestId))
        }
      }
      await Promise.all(reloads)
    },

    keepAsNew: (tabId) => {
      builder().detach(tabId)
      patchTab(tabId, { requestId: null, pinned: true })
      set((state) => ({
        notices: withoutNotice(state.notices, tabId),
        previewTabId: state.previewTabId === tabId ? null : state.previewTabId,
      }))
    },

    keepMine: async (tabId) => {
      const saved = await builder().save(tabId, true)
      if (saved.ok && saved.value) {
        set((state) => ({ notices: withoutNotice(state.notices, tabId) }))
        await syncSiblings(tabId)
      }
    },

    loadFromDisk: async (tabId) => {
      const requestId = tabById(tabId)?.requestId
      if (requestId && (await builder().openSaved(tabId, requestId))) {
        set((state) => ({ notices: withoutNotice(state.notices, tabId) }))
      }
    },

    snapshot: () => {
      const state = get()
      return {
        tabs: state.tabs.map(tabSnapshot),
        closedTabs: state.closedTabs,
        snippetOptions: state.snippetOptions,
        activeTabId: state.activeTabId,
        previewTabId: state.previewTabId,
        sidebar: sidebarSnapshot(),
        layout: useWorkspacesStore.getState().layout,
      }
    },

    schedulePersist: () => {
      cancelPersist()
      persistTimer = setTimeout(() => {
        persistTimer = undefined
        void get().flushPersist()
      }, PERSIST_DELAY_MS)
    },

    flushPersist: async () => {
      cancelPersist()
      const localId = get().localId
      if (localId === null) {
        return
      }
      try {
        await saveUIState(localId, get().snapshot())
      } catch {
        return
      }
    },

    reset: () => {
      if (persistTimer !== undefined) {
        void get().flushPersist()
      }
      cancelPersist()
      examples().reset()
      set(initialState)
    },
  }
})
