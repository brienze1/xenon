import { create } from 'zustand'
import { useExamplesStore } from '@/features/examples/store'
import { useRequestBuilderStore } from '@/features/request-builder/store'
import { useResponseViewerStore } from '@/features/response-viewer/store'
import { useSidebarStore } from '@/features/sidebar/store'
import { useWorkspacesStore } from '@/features/workspaces/store'
import type { AppError } from '@/platform/errors'
import type { RequestChangedOnDisk } from '@/platform/events'
import { emptyDraft, type RequestDraft } from '@/platform/mappers/request'
import type { ClosedTabState, SnippetOptionsState, TabState, UIState } from '@/platform/mappers/ui-state'
import { saveUIState } from '@/platform/services/ui-state'
import { sameDraftContent } from './lib/draft'
import { withoutNotice } from './lib/notices'
import { latestClosedTabs } from './lib/closed'
import { moveTab, neighbourOf, positionAt, tabsAfter, tabsBefore, wrappedNeighbour } from './lib/order'
import type { CloseChoice, SaveDialogState, TabMeta, TabNotice } from './model'

export const PERSIST_DELAY_MS = 300

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
  hydrate: (localId: string, state: UIState) => void
  newTab: () => string
  openDraft: (draft: RequestDraft) => string
  activate: (tabId: string) => void
  openPreview: (requestId: string) => Promise<void>
  openPinned: (requestId: string) => Promise<void>
  openExamplePreview: (requestId: string, exampleId: string) => Promise<void>
  openExamplePinned: (requestId: string, exampleId: string) => Promise<void>
  handleExampleDeleted: (exampleId: string) => void
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

function isExampleTab(tab: TabMeta | undefined): boolean {
  return tab?.kind === 'example'
}

function isDirty(tabId: string): boolean {
  return (builder().tabs[tabId]?.dirty ?? false) || examples().isDirty(tabId)
}

function isPristine(draft: RequestDraft): boolean {
  return sameDraftContent(draft, emptyDraft())
}

function cancelPersist(): void {
  clearTimeout(persistTimer)
  persistTimer = undefined
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
    builder().ensureTab(meta.id)
    set((state) => {
      const index = after === undefined ? -1 : state.tabs.findIndex((tab) => tab.id === after)
      const tabs = index < 0 ? [...state.tabs, meta] : [...state.tabs.slice(0, index + 1), meta, ...state.tabs.slice(index + 1)]
      return { tabs }
    })
    focus(meta.id)
  }

  function remember(tab: TabMeta) {
    const snapshot = tabSnapshot(tab)
    if (snapshot.requestId === null && (snapshot.draft === null || isPristine(snapshot.draft))) {
      return
    }
    set((state) => ({ closedTabs: latestClosedTabs([...state.closedTabs, { closedAt: new Date().toISOString(), tab: snapshot }]) }))
  }

  function closeMany(tabIds: string[], force: boolean) {
    const dirtyIds: string[] = []
    for (const tabId of tabIds) {
      if (!force && isDirty(tabId)) {
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
      if (isDirty(sibling.id)) {
        set((state) => ({ notices: { ...state.notices, [sibling.id]: 'changed' } }))
      } else {
        reloads.push(builder().openSaved(sibling.id, savedTab.requestId))
      }
    }
    await Promise.all(reloads)
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
    }
  }

  function tabSnapshot(tab: TabMeta): TabState {
    if (isExampleTab(tab)) {
      return { tabId: tab.id, kind: 'example', requestId: tab.requestId, exampleId: tab.exampleId ?? null, pinned: tab.pinned, baseRevision: null, draft: null }
    }
    const draft = builder().tabs[tab.id]
    const keepDraft = tab.requestId === null || (draft?.dirty ?? false)
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
        tabs: state.tabs.map((tab) =>
          tab.kind === 'example'
            ? { id: tab.tabId, requestId: tab.requestId, pinned: tab.pinned, kind: 'example' as const, exampleId: tab.exampleId ?? null }
            : { id: tab.tabId, requestId: tab.requestId, pinned: tab.pinned },
        ),
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
      })
      for (const tab of state.tabs) {
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
      if (preview !== null && !isDirty(preview)) {
        patchTab(preview, { requestId, kind: undefined, exampleId: undefined })
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
      if (preview !== null && !isDirty(preview)) {
        patchTab(preview, { requestId, kind: 'example', exampleId })
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
      if (preview !== null && isDirty(preview)) {
        get().pin(preview)
      }
    },

    reorder: (fromId, toId) => set((state) => ({ tabs: moveTab(state.tabs, fromId, toId) })),

    requestClose: (tabId) => {
      const target = activeOr(tabId)
      if (target === null) {
        return
      }
      if (isDirty(target)) {
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
      if (!tab?.requestId) {
        set({ closePrompt: null, closeError: null, saveDialog: { tabId, mode: 'save', closeAfter: true } })
        return
      }
      set({ closeError: null })
      const saved = isExampleTab(tab) ? await examples().save(tabId) : await builder().save(tabId)
      if (!saved.ok) {
        set((state) => (state.closePrompt === tabId ? { closeError: saved.error } : state))
        return
      }
      set({ closePrompt: null })
      if (saved.value) {
        await syncSiblings(tabId)
        get().closeTab(tabId)
      }
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
      builder().loadDraft(id, structuredClone(draft.draft), draft.baseRevision, draft.dirty)
      return id
    },

    revealInSidebar: (tabId) => {
      const requestId = tabById(tabId)?.requestId
      if (!requestId) {
        return
      }
      const sidebar = useSidebarStore.getState()
      if (sidebar.view !== 'collections') {
        sidebar.setView('collections')
      }
      sidebar.reveal(requestId)
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
      const target = activeOr(tabId)
      const tab = tabById(target)
      if (!tab) {
        return
      }
      if (isExampleTab(tab)) {
        await examples().save(tab.id)
        return
      }
      if (tab.requestId === null) {
        set({ saveDialog: { tabId: tab.id, mode: 'save', closeAfter: false } })
        return
      }
      const saved = await builder().save(tab.id)
      if (saved.ok && saved.value) {
        await syncSiblings(tab.id)
      }
    },

    saveAs: (tabId) => {
      const target = activeOr(tabId)
      if (target !== null && !isExampleTab(tabById(target))) {
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
      const drafts = builder().tabs
      for (const tab of get().tabs) {
        if (tab.requestId === null || !deleted.has(tab.requestId)) {
          continue
        }
        if (isExampleTab(tab)) {
          get().closeTab(tab.id, false)
        } else if (drafts[tab.id]?.dirty ?? false) {
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
          if (loaded?.example?.revision !== event.revision && !isDirty(tab.id) && tab.exampleId) {
            reloads.push(examples().open(tab.id, event.requestId, tab.exampleId))
          }
          continue
        }
        if (draft?.baseRevision === event.revision) {
          continue
        }
        if (draft?.dirty ?? false) {
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
