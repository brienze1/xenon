import { useMemo } from 'react'
import { HOTKEYS, type HotkeyHandlers } from '@/app/hotkeys'
import { useHotkeys, type HotkeyBinding, type HotkeyPlatform } from '@/shared/hooks/use-hotkey'

/** Binds the hotkey table to the handlers given; shortcuts without a handler stay with the webview. */
export function useAppHotkeys(handlers: HotkeyHandlers, platform?: HotkeyPlatform): void {
  const bindings = useMemo<HotkeyBinding[]>(
    () =>
      HOTKEYS.flatMap((hotkey) => {
        const handler = handlers[hotkey.action]
        return handler ? hotkey.combos.map((combo) => ({ combo, handler: () => handler() })) : []
      }),
    [handlers],
  )
  useHotkeys(bindings, { platform })
}
