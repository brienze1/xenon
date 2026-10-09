import { useEffect, useRef, type RefObject } from 'react'

export type HotkeyPlatform = 'mac' | 'other'

interface ParsedCombo {
  key: string
  mod: boolean
  shift: boolean
  alt: boolean
}

export function detectPlatform(platform: string = navigator.platform): HotkeyPlatform {
  return /mac|iphone|ipad/i.test(platform) ? 'mac' : 'other'
}

function parseCombo(combo: string): ParsedCombo {
  const parts = combo.split('+').map((part) => part.trim().toLowerCase())
  return {
    key: parts[parts.length - 1] ?? '',
    mod: parts.includes('mod'),
    shift: parts.includes('shift'),
    alt: parts.includes('alt'),
  }
}

export function matchesHotkey(event: KeyboardEvent, combo: string, platform: HotkeyPlatform): boolean {
  const parsed = parseCombo(combo)
  const modPressed = platform === 'mac' ? event.metaKey : event.ctrlKey
  const otherModPressed = platform === 'mac' ? event.ctrlKey : event.metaKey
  return (
    event.key.toLowerCase() === parsed.key &&
    modPressed === parsed.mod &&
    !otherModPressed &&
    event.shiftKey === parsed.shift &&
    event.altKey === parsed.alt
  )
}

export function formatHotkey(combo: string, platform: HotkeyPlatform): string {
  const parsed = parseCombo(combo)
  const key = parsed.key.length === 1 ? parsed.key.toUpperCase() : parsed.key.charAt(0).toUpperCase() + parsed.key.slice(1)
  if (platform === 'mac') {
    return `${parsed.mod ? '⌘' : ''}${parsed.alt ? '⌥' : ''}${parsed.shift ? '⇧' : ''}${key}`
  }
  return [parsed.mod ? 'Ctrl' : '', parsed.alt ? 'Alt' : '', parsed.shift ? 'Shift' : '', key].filter(Boolean).join('+')
}

interface HotkeyOptions {
  scope?: RefObject<HTMLElement | null>
  enabled?: boolean
  platform?: HotkeyPlatform
}

export function useHotkey(combo: string, handler: (event: KeyboardEvent) => void, options: HotkeyOptions = {}): void {
  const { scope, enabled = true, platform } = options
  const handlerRef = useRef(handler)

  useEffect(() => {
    handlerRef.current = handler
  }, [handler])

  useEffect(() => {
    if (!enabled) {
      return
    }
    const target: HTMLElement | Window | null = scope ? scope.current : window
    if (!target) {
      return
    }
    const resolvedPlatform = platform ?? detectPlatform()
    function onKeyDown(event: Event) {
      const keyboardEvent = event as KeyboardEvent
      if (matchesHotkey(keyboardEvent, combo, resolvedPlatform)) {
        keyboardEvent.preventDefault()
        handlerRef.current(keyboardEvent)
      }
    }
    target.addEventListener('keydown', onKeyDown)
    return () => target.removeEventListener('keydown', onKeyDown)
  }, [combo, enabled, platform, scope])
}
