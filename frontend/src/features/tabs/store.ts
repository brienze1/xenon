import { create } from 'zustand'
import { useRequestBuilderStore } from '@/features/request-builder/store'
import { useResponseViewerStore } from '@/features/response-viewer/store'
import { useSidebarStore } from '@/features/sidebar/store'
import type { RequestChangedOnDisk } from '@/platform/events'
import { emptyDraft, toRequestDto, type RequestDraft } from '@/platform/mappers/request'
import { useWorkspacesStore } from '@/features/workspaces/store'
import type { TabState, UIState } from '@/platform/mappers/ui-state'
import { saveUIState } from '@/platform/services/ui-state'
import { moveTab } from './lib/order'

export const PERSIST_DELAY_MS = 300

const PRISTINE_DRAFT = JSON.stringify(toRequestDto(emptyDraft()))

function isPristine(draft: RequestDraft): boolean {
  return JSON.stringify(toRequestDto(draft)) === PRISTINE_DRAFT
}

export interface TabMeta {
  id: string
  requestId: string | null
  pinned: boolean
}

export type CloseChoice = 'save' | 'discard' | 'cancel'
export type TabNotice = 'changed' | 'deleted'

export interface SaveDialogState {
  tabId: string
  mode: 'save' | 'saveAs'
  closeAfter: boolean
}

interface TabsState {
  localId: string | null
  tabs: TabMeta[]
  activeTabId: string | null
  previewTabId: string | null
  closePrompt: string | null
  saveDialog: SaveDialogState | null
  notices: Record<string, TabNotice>
  hydrate: (localId: string, state: UIState) => void
  newTab: () => string
  openDraft: (draft: RequestDraft) => string
  activate: (tabId: string) => void
  openPreview: (requestId: string) => Promise<void>
  openPinned: (requestId: string) => Promise<void>
  pin: (tabId: string) => void
  pinIfEdited: () => void
  reorder: (fromId: string, toId: string) => void
  requestClose: (tabId?: string) => void
  resolveClose: (choice: CloseChoice) => Promise<void>
  closeTab: (tabId: string) => void
  save: (tabId?: string) => Promise<void>
  saveAs: (tabId?: string) => void
  confirmSave: (parentId: string, name: string) => Promise<void>
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

function isDirty(tabId: string): boolean {
  return builder().tabs[tabId]?.dirty ?? false
}

function withoutNotice(notices: Record<string, TabNotice>, tabId: string): Record<string, TabNotice> {
  const next = { ...notices }
  delete next[tabId]
  return next
}

function neighbourOf(tabs: TabMeta[], tabId: string): string | null {
  const index = tabs.findIndex((tab) => tab.id === tabId)
  return tabs[index + 1]?.id ?? tabs[index - 1]?.id ?? null
}

const initialState = {
  localId: null as string | null,
  tabs: [] as TabMeta[],
  activeTabId: null as string | null,
  previewTabId: null as string | null,
  closePrompt: null as string | null,
  saveDialog: null as SaveDialogState | null,
  notices: {} as Record<string, TabNotice>,
}

export const useTabsStore = create<TabsState>()((set, get) => {
  function patchTab(tabId: string, update: Partial<TabMeta>) {
    set((state) => ({ tabs: state.tabs.map((tab) => (tab.id === tabId ? { ...tab, ...update } : tab)) }))
  }

  function tabOf(requestId: string): TabMeta | undefined {
    return get().tabs.find((tab) => tab.requestId === requestId)
  }

  function addTab(meta: TabMeta) {
    builder().ensureTab(meta.id)
    set((state) => ({ tabs: [...state.tabs, meta], activeTabId: meta.id }))
  }

  function activeOr(tabId: string | undefined): string | null {
    return tabId ?? get().activeTabId
  }

  function sidebarSnapshot(): UIState['sidebar'] {
    const sidebar = useSidebarStore.getState()
    return {
      expanded: Object.entries(sidebar.expanded)
        .filter(([, open]) => open)
        .map(([id]) => id),
      selectedId: sidebar.selectedId,
      view: sidebar.view,
      search: sidebar.query,
    }
  }

  function tabSnapshot(tab: TabMeta): TabState {
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
        tabs: state.tabs.map((tab) => ({ id: tab.tabId, requestId: tab.requestId, pinned: tab.pinned })),
        activeTabId: state.activeTabId,
        previewTabId: state.previewTabId,
      })
      useSidebarStore.getState().applyState({
        expanded: Object.fromEntries(state.sidebar.expanded.map((id) => [id, true])),
        selectedId: state.sidebar.selectedId,
        query: state.sidebar.search,
        view: state.sidebar.view,
      })
      for (const tab of state.tabs) {
        if (tab.draft) {
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
      } else if (get().activeTabId === null || !get().tabs.some((tab) => tab.id === get().activeTabId)) {
        set({ activeTabId: state.tabs[0]!.tabId })
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

    activate: (tabId) => set({ activeTabId: tabId }),

    openPreview: async (requestId) => {
      const existing = tabOf(requestId)
      if (existing) {
        set({ activeTabId: existing.id })
        return
      }
      const preview = get().previewTabId
      if (preview !== null && !isDirty(preview)) {
        patchTab(preview, { requestId })
        set({ activeTabId: preview, notices: withoutNotice(get().notices, preview) })
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
        set({ activeTabId: existing.id })
        return
      }
      const id = newTabId()
      addTab({ id, requestId, pinned: true })
      await builder().openSaved(id, requestId)
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
        set({ closePrompt: target })
        return
      }
      get().closeTab(target)
    },

    resolveClose: async (choice) => {
      const tabId = get().closePrompt
      set({ closePrompt: null })
      if (tabId === null || choice === 'cancel') {
        return
      }
      if (choice === 'discard') {
        get().closeTab(tabId)
        return
      }
      const tab = get().tabs.find((candidate) => candidate.id === tabId)
      if (!tab?.requestId) {
        set({ saveDialog: { tabId, mode: 'save', closeAfter: true } })
        return
      }
      if (await builder().save(tabId)) {
        get().closeTab(tabId)
      }
    },

    closeTab: (tabId) => {
      set((state) => ({
        tabs: state.tabs.filter((tab) => tab.id !== tabId),
        activeTabId: state.activeTabId === tabId ? neighbourOf(state.tabs, tabId) : state.activeTabId,
        previewTabId: state.previewTabId === tabId ? null : state.previewTabId,
        notices: withoutNotice(state.notices, tabId),
      }))
      builder().dropTab(tabId)
      void useResponseViewerStore.getState().release(tabId)
    },

    save: async (tabId) => {
      const target = activeOr(tabId)
      const tab = get().tabs.find((candidate) => candidate.id === target)
      if (!tab) {
        return
      }
      if (tab.requestId === null) {
        set({ saveDialog: { tabId: tab.id, mode: 'save', closeAfter: false } })
        return
      }
      await builder().save(tab.id)
    },

    saveAs: (tabId) => {
      const target = activeOr(tabId)
      if (target !== null) {
        set({ saveDialog: { tabId: target, mode: 'saveAs', closeAfter: false } })
      }
    },

    confirmSave: async (parentId, name) => {
      const dialog = get().saveDialog
      if (dialog === null) {
        return
      }
      const requestId = await builder().saveAs(dialog.tabId, parentId, name)
      if (requestId === null) {
        return
      }
      set({ saveDialog: null, notices: withoutNotice(get().notices, dialog.tabId) })
      get().pin(dialog.tabId)
      patchTab(dialog.tabId, { requestId })
      if (dialog.closeAfter) {
        get().closeTab(dialog.tabId)
      }
    },

    cancelSave: () => set({ saveDialog: null }),

    handleDeleted: (requestIds) => {
      const deleted = new Set(requestIds)
      for (const tab of get().tabs) {
        if (tab.requestId === null || !deleted.has(tab.requestId)) {
          continue
        }
        if (isDirty(tab.id)) {
          set((state) => ({ notices: { ...state.notices, [tab.id]: 'deleted' } }))
        } else {
          get().closeTab(tab.id)
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
      for (const tab of get().tabs.filter((candidate) => candidate.requestId === event.requestId)) {
        const draft = builder().tabs[tab.id]
        if (draft?.baseRevision === event.revision) {
          continue
        }
        if (isDirty(tab.id)) {
          set((state) => ({ notices: { ...state.notices, [tab.id]: 'changed' } }))
        } else {
          await builder().openSaved(tab.id, event.requestId)
        }
      }
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
      if (await builder().save(tabId, true)) {
        set((state) => ({ notices: withoutNotice(state.notices, tabId) }))
      }
    },

    loadFromDisk: async (tabId) => {
      const requestId = get().tabs.find((tab) => tab.id === tabId)?.requestId
      if (requestId && (await builder().openSaved(tabId, requestId))) {
        set((state) => ({ notices: withoutNotice(state.notices, tabId) }))
      }
    },

    snapshot: () => {
      const state = get()
      return {
        tabs: state.tabs.map(tabSnapshot),
        activeTabId: state.activeTabId,
        previewTabId: state.previewTabId,
        sidebar: sidebarSnapshot(),
        layout: useWorkspacesStore.getState().layout,
      }
    },

    schedulePersist: () => {
      clearTimeout(persistTimer)
      persistTimer = setTimeout(() => {
        persistTimer = undefined
        void get().flushPersist()
      }, PERSIST_DELAY_MS)
    },

    flushPersist: async () => {
      clearTimeout(persistTimer)
      persistTimer = undefined
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
      set(initialState)
    },
  }
})
