import type {
  Layout as LayoutDto,
  Tab as TabDto,
  UIState as UIStateDto,
} from '../../../bindings/github.com/brienze1/xenon/internal/integration/entrypoint/dto/models'
import { toRequestDraft, toRequestDto, type RequestDraft } from './request'

export type SidebarView = 'collections' | 'history' | 'environments'
export type SplitOrientation = 'stacked' | 'side-by-side'

/** `environment` and `globals` tabs edit variables instead of a request. */
export type TabKind = 'request' | 'example' | 'environment' | 'globals'

export interface TabState {
  tabId: string
  /** Absent means `request`. */
  kind?: TabKind
  requestId: string | null
  /** Set for example tabs only. */
  exampleId?: string | null
  /** Set for environment tabs only. */
  environmentId?: string | null
  pinned: boolean
  baseRevision: string | null
  draft: RequestDraft | null
}

export interface SidebarUIState {
  expanded: string[]
  selectedId: string | null
  view: SidebarView
  search: string
}

export interface LayoutState {
  sidebarWidth: number
  sidebarVisible: boolean
  rightPanelWidth: number
  rightPanelVisible: boolean
  rightPanelView: string
  consoleHeight: number
  consoleVisible: boolean
  split: SplitOrientation
  splitRatio: number
}

/** A tab closed recently, for "Reopen closed tab" (newest last, at most 20). */
export interface ClosedTabState {
  closedAt: string
  tab: TabState
}

/** The options last used for one snippet language. */
export interface SnippetOptionsState {
  language: string
  options: Record<string, string>
}

export interface UIState {
  tabs: TabState[]
  closedTabs: ClosedTabState[]
  snippetOptions: SnippetOptionsState[]
  activeTabId: string | null
  previewTabId: string | null
  sidebar: SidebarUIState
  layout: LayoutState
}

export function defaultLayout(): LayoutState {
  return {
    sidebarWidth: 280,
    sidebarVisible: true,
    rightPanelWidth: 320,
    rightPanelVisible: true,
    rightPanelView: 'info',
    consoleHeight: 200,
    consoleVisible: false,
    split: 'stacked',
    splitRatio: 0.5,
  }
}

export function defaultUIState(): UIState {
  return {
    tabs: [],
    closedTabs: [],
    snippetOptions: [],
    activeTabId: null,
    previewTabId: null,
    sidebar: { expanded: [], selectedId: null, view: 'collections', search: '' },
    layout: defaultLayout(),
  }
}

function orNull(value: string): string | null {
  return value === '' ? null : value
}

const SIDEBAR_VIEWS: readonly SidebarView[] = ['collections', 'history', 'environments']

function toSidebarView(view: string): SidebarView {
  return SIDEBAR_VIEWS.find((known) => known === view) ?? 'collections'
}

function toTabKind(dto: TabDto): TabKind {
  if (dto.kind === 'globals') {
    return 'globals'
  }
  if (dto.kind === 'environment' && dto.environment_id) {
    return 'environment'
  }
  return dto.kind === 'example' && dto.example_id ? 'example' : 'request'
}

function toTabState(dto: TabDto): TabState {
  return {
    tabId: dto.tab_id,
    kind: toTabKind(dto),
    requestId: orNull(dto.request_id),
    exampleId: orNull(dto.example_id ?? ''),
    environmentId: orNull(dto.environment_id ?? ''),
    pinned: dto.pinned,
    baseRevision: orNull(dto.base_revision),
    draft: dto.draft ? toRequestDraft(dto.draft) : null,
  }
}

function toLayout(dto: LayoutDto): LayoutState {
  return {
    sidebarWidth: dto.sidebar_width,
    sidebarVisible: dto.sidebar_visible,
    rightPanelWidth: dto.right_panel_width,
    rightPanelVisible: dto.right_panel_visible,
    rightPanelView: dto.right_panel_view,
    consoleHeight: dto.console_height,
    consoleVisible: dto.console_visible,
    split: dto.split === 'side-by-side' ? 'side-by-side' : 'stacked',
    splitRatio: dto.split_ratio,
  }
}

export function toUIState(dto: UIStateDto | null): UIState {
  if (!dto) {
    return defaultUIState()
  }
  return {
    tabs: (dto.tabs ?? []).map(toTabState),
    closedTabs: (dto.closed_tabs ?? []).map((closed) => ({ closedAt: closed.closed_at, tab: toTabState(closed.tab) })),
    snippetOptions: (dto.snippet_options ?? []).map((saved) => ({ language: saved.language, options: Object.fromEntries(Object.entries(saved.options ?? {}).filter((entry): entry is [string, string] => typeof entry[1] === 'string')) })),
    activeTabId: orNull(dto.active_tab_id),
    previewTabId: orNull(dto.preview_tab_id),
    sidebar: {
      expanded: dto.sidebar.expanded ?? [],
      selectedId: orNull(dto.sidebar.selected_id),
      view: toSidebarView(dto.sidebar.view),
      search: dto.sidebar.search,
    },
    layout: toLayout(dto.layout),
  }
}

function toTabDto(tab: TabState): TabDto {
  return {
    tab_id: tab.tabId,
    kind: tab.kind ?? 'request',
    request_id: tab.requestId ?? '',
    example_id: tab.exampleId ?? '',
    environment_id: tab.environmentId ?? '',
    pinned: tab.pinned,
    base_revision: tab.baseRevision ?? '',
    ...(tab.draft ? { draft: toRequestDto(tab.draft) } : {}),
  }
}

export function toUIStateDto(localId: string, state: UIState): UIStateDto {
  return {
    local_id: localId,
    version: 1,
    active_tab_id: state.activeTabId ?? '',
    preview_tab_id: state.previewTabId ?? '',
    tabs: state.tabs.map(toTabDto),
    closed_tabs: state.closedTabs.map((closed) => ({ closed_at: closed.closedAt, tab: toTabDto(closed.tab) })),
    snippet_options: state.snippetOptions.map((saved) => ({ language: saved.language, options: saved.options })),
    sidebar: {
      expanded: state.sidebar.expanded,
      selected_id: state.sidebar.selectedId ?? '',
      view: state.sidebar.view,
      search: state.sidebar.search,
    },
    layout: {
      sidebar_width: state.layout.sidebarWidth,
      sidebar_visible: state.layout.sidebarVisible,
      right_panel_width: state.layout.rightPanelWidth,
      right_panel_visible: state.layout.rightPanelVisible,
      right_panel_view: state.layout.rightPanelView,
      console_height: state.layout.consoleHeight,
      console_visible: state.layout.consoleVisible,
      split: state.layout.split,
      split_ratio: state.layout.splitRatio,
    },
  }
}
