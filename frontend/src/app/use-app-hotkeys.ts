import { HOTKEY_BY_ACTION, type HotkeyAction, type HotkeyDefinition, type HotkeyHandlers } from '@/app/hotkeys'
import { useHotkey } from '@/shared/hooks/use-hotkey'

function useAppHotkey(action: HotkeyAction, handlers: HotkeyHandlers) {
  const handler = handlers[action]
  const { combo } = HOTKEY_BY_ACTION.get(action) as HotkeyDefinition
  useHotkey(combo, () => handler?.(), { enabled: handler !== undefined })
}

export function useAppHotkeys(handlers: HotkeyHandlers): void {
  useAppHotkey('send', handlers)
  useAppHotkey('save', handlers)
  useAppHotkey('saveAs', handlers)
  useAppHotkey('newTab', handlers)
  useAppHotkey('closeTab', handlers)
  useAppHotkey('environmentSwitcher', handlers)
  useAppHotkey('import', handlers)
}
