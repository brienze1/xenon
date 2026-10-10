export type HotkeyAction = 'send' | 'save' | 'saveAs' | 'newTab' | 'closeTab' | 'environmentSwitcher' | 'import'

export interface HotkeyDefinition {
  action: HotkeyAction
  combo: string
  label: string
}

export const HOTKEYS: readonly HotkeyDefinition[] = [
  { action: 'send', combo: 'Mod+Enter', label: 'Send request' },
  { action: 'save', combo: 'Mod+S', label: 'Save' },
  { action: 'saveAs', combo: 'Mod+Shift+S', label: 'Save as' },
  { action: 'newTab', combo: 'Mod+T', label: 'New tab' },
  { action: 'closeTab', combo: 'Mod+W', label: 'Close tab' },
  { action: 'environmentSwitcher', combo: 'Mod+Shift+E', label: 'Switch environment' },
  { action: 'import', combo: 'Mod+O', label: 'Import' },
]

export const HOTKEY_BY_ACTION: ReadonlyMap<HotkeyAction, HotkeyDefinition> = new Map(
  HOTKEYS.map((definition) => [definition.action, definition]),
)

export type HotkeyHandlers = Partial<Record<HotkeyAction, () => void>>
