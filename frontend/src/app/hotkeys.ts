import { useHotkey } from '@/shared/hooks/use-hotkey'

export type HotkeyAction = 'send' | 'save' | 'saveAs' | 'newTab' | 'closeTab'

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
]

export type HotkeyHandlers = Partial<Record<HotkeyAction, () => void>>

function hotkey(action: HotkeyAction): HotkeyDefinition {
  return HOTKEYS.find((definition) => definition.action === action) as HotkeyDefinition
}

function useAppHotkey(action: HotkeyAction, handlers: HotkeyHandlers) {
  const handler = handlers[action]
  useHotkey(hotkey(action).combo, () => handler?.(), { enabled: handler !== undefined })
}

export function useAppHotkeys(handlers: HotkeyHandlers): void {
  useAppHotkey('send', handlers)
  useAppHotkey('save', handlers)
  useAppHotkey('saveAs', handlers)
  useAppHotkey('newTab', handlers)
  useAppHotkey('closeTab', handlers)
}
