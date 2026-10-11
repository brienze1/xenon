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

const parsedCombos = new Map<string, ParsedCombo>()

function parseCombo(combo: string): ParsedCombo {
  const cached = parsedCombos.get(combo)
  if (cached !== undefined) {
    return cached
  }
  const parsed = readCombo(combo)
  parsedCombos.set(combo, parsed)
  return parsed
}

function readCombo(combo: string): ParsedCombo {
  const parts = combo.split('+').map((part) => part.trim().toLowerCase())
  return {
    key: parts[parts.length - 1] ?? '',
    mod: parts.includes('mod'),
    shift: parts.includes('shift'),
    alt: parts.includes('alt'),
  }
}

/** With Alt the key may type another character (Option+C is ç on a Mac), so Alt combos match the physical key. */
function matchesKey(event: KeyboardEvent, parsed: ParsedCombo): boolean {
  if (parsed.alt && /^[a-z0-9]$/.test(parsed.key)) {
    return event.code === (/[a-z]/.test(parsed.key) ? `Key${parsed.key.toUpperCase()}` : `Digit${parsed.key}`)
  }
  return event.key.toLowerCase() === parsed.key
}

export function matchesHotkey(event: KeyboardEvent, combo: string, platform: HotkeyPlatform): boolean {
  const parsed = parseCombo(combo)
  const modPressed = platform === 'mac' ? event.metaKey : event.ctrlKey
  const otherModPressed = platform === 'mac' ? event.ctrlKey : event.metaKey
  if (event.getModifierState?.('AltGraph')) {
    return false
  }
  return (
    matchesKey(event, parsed) &&
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

interface HotkeyRegistration {
  combo: string
  platform: HotkeyPlatform
  handler: RefObject<(event: KeyboardEvent) => void>
}

interface TargetListener {
  registrations: Set<HotkeyRegistration>
  onKeyDown: (event: Event) => void
}

const registry = new Map<EventTarget, TargetListener>()

function dispatchHotkey(registrations: Set<HotkeyRegistration>, event: KeyboardEvent) {
  for (const registration of [...registrations]) {
    if (matchesHotkey(event, registration.combo, registration.platform)) {
      event.preventDefault()
      registration.handler.current(event)
    }
  }
}

function register(target: EventTarget, registration: HotkeyRegistration): () => void {
  let entry = registry.get(target)
  if (!entry) {
    const registrations = new Set<HotkeyRegistration>()
    entry = { registrations, onKeyDown: (event) => dispatchHotkey(registrations, event as KeyboardEvent) }
    registry.set(target, entry)
    target.addEventListener('keydown', entry.onKeyDown)
  }
  const { registrations, onKeyDown } = entry
  registrations.add(registration)
  return () => {
    registrations.delete(registration)
    if (registrations.size === 0) {
      registry.delete(target)
      target.removeEventListener('keydown', onKeyDown)
    }
  }
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
    return register(target, { combo, platform: platform ?? detectPlatform(), handler: handlerRef })
  }, [combo, enabled, platform, scope])
}
