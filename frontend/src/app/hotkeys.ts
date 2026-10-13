export type HotkeyAction =
  | 'send'
  | 'save'
  | 'saveAs'
  | 'newTab'
  | 'closeTab'
  | 'forceCloseTab'
  | 'reopenClosedTab'
  | 'nextTab'
  | 'previousTab'
  | 'goToTab1'
  | 'goToTab2'
  | 'goToTab3'
  | 'goToTab4'
  | 'goToTab5'
  | 'goToTab6'
  | 'goToTab7'
  | 'goToTab8'
  | 'lastTab'
  | 'focusUrl'
  | 'searchRequests'
  | 'findInResponse'
  | 'toggleSidebar'
  | 'console'
  | 'renameSelected'
  | 'duplicateSelected'
  | 'import'
  | 'settings'
  | 'environmentSwitcher'

export interface HotkeyDefinition {
  action: HotkeyAction
  /** `Mod` is Cmd on macOS and Ctrl elsewhere; `Ctrl` is Control on every OS. */
  combos: readonly string[]
  label: string
}

const GO_TO_TAB_ACTIONS = ['goToTab1', 'goToTab2', 'goToTab3', 'goToTab4', 'goToTab5', 'goToTab6', 'goToTab7', 'goToTab8'] as const

/** Every shortcut of the PRD (XEN-005 FR-012); the Settings shortcut list renders from this table. */
export const HOTKEYS: readonly HotkeyDefinition[] = [
  { action: 'send', combos: ['Mod+Enter'], label: 'Send request' },
  { action: 'save', combos: ['Mod+S'], label: 'Save' },
  { action: 'saveAs', combos: ['Mod+Shift+S'], label: 'Save as' },
  { action: 'newTab', combos: ['Mod+T'], label: 'New tab' },
  { action: 'closeTab', combos: ['Mod+W'], label: 'Close tab' },
  { action: 'forceCloseTab', combos: ['Mod+Alt+W'], label: 'Force close tab' },
  { action: 'reopenClosedTab', combos: ['Mod+Shift+T'], label: 'Reopen last closed tab' },
  { action: 'nextTab', combos: ['Ctrl+Tab', 'Mod+Shift+]'], label: 'Next tab' },
  { action: 'previousTab', combos: ['Ctrl+Shift+Tab', 'Mod+Shift+['], label: 'Previous tab' },
  ...GO_TO_TAB_ACTIONS.map((action, index) => ({ action, combos: [`Mod+${index + 1}`], label: `Go to tab ${index + 1}` })),
  { action: 'lastTab', combos: ['Mod+9'], label: 'Go to the last tab' },
  { action: 'focusUrl', combos: ['Mod+L'], label: 'Focus the URL' },
  { action: 'searchRequests', combos: ['Mod+K'], label: 'Search requests' },
  { action: 'findInResponse', combos: ['Mod+F'], label: 'Find in response' },
  { action: 'toggleSidebar', combos: ['Mod+\\'], label: 'Show or hide the sidebar' },
  { action: 'console', combos: ['Mod+Alt+C'], label: 'Show or hide the console' },
  { action: 'renameSelected', combos: ['Mod+E'], label: 'Rename the selected item' },
  { action: 'duplicateSelected', combos: ['Mod+D'], label: 'Duplicate the selected item' },
  { action: 'import', combos: ['Mod+O'], label: 'Import' },
  { action: 'settings', combos: ['Mod+,'], label: 'Settings' },
  { action: 'environmentSwitcher', combos: ['Mod+Shift+E'], label: 'Switch environment' },
]

export const HOTKEY_BY_ACTION: ReadonlyMap<HotkeyAction, HotkeyDefinition> = new Map(
  HOTKEYS.map((definition) => [definition.action, definition]),
)

export const GO_TO_TAB_POSITIONS: ReadonlyMap<HotkeyAction, number> = new Map(GO_TO_TAB_ACTIONS.map((action, index) => [action, index + 1]))

export type HotkeyHandlers = Partial<Record<HotkeyAction, () => void>>
