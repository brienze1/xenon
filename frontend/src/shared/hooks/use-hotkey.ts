import { useEffect, useRef, type RefObject } from 'react'

export type HotkeyPlatform = 'mac' | 'other'

interface ParsedCombo {
  key: string
  mod: boolean
  ctrl: boolean
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
    ctrl: parts.includes('ctrl'),
    shift: parts.includes('shift'),
    alt: parts.includes('alt'),
  }
}

const KEY_CODES: Readonly<Record<string, string>> = {
  enter: 'Enter',
  tab: 'Tab',
  escape: 'Escape',
  '\\': 'Backslash',
  ',': 'Comma',
  '.': 'Period',
  '/': 'Slash',
  '[': 'BracketLeft',
  ']': 'BracketRight',
}

function codeOf(key: string): string | undefined {
  if (/^[a-z]$/.test(key)) {
    return `Key${key.toUpperCase()}`
  }
  if (/^[0-9]$/.test(key)) {
    return `Digit${key}`
  }
  return KEY_CODES[key]
}

/** Shortcuts match the physical key (`event.code`), so every layout presses the same keys; `key` is the fallback for events without a code. */
function matchesKey(event: KeyboardEvent, parsed: ParsedCombo): boolean {
  const code = codeOf(parsed.key)
  if (event.code && code) {
    return event.code === code || (code === 'Enter' && event.code === 'NumpadEnter')
  }
  return event.key.toLowerCase() === parsed.key
}

/** AltGr (Ctrl+Alt on Windows) types characters on many layouts, so no shortcut fires while it is held. */
export function matchesHotkey(event: KeyboardEvent, combo: string, platform: HotkeyPlatform): boolean {
  const parsed = parseCombo(combo)
  if (event.getModifierState?.('AltGraph')) {
    return false
  }
  const meta = platform === 'mac' && parsed.mod
  const ctrl = platform === 'mac' ? parsed.ctrl : parsed.mod || parsed.ctrl
  return (
    matchesKey(event, parsed) &&
    event.metaKey === meta &&
    event.ctrlKey === ctrl &&
    event.shiftKey === parsed.shift &&
    event.altKey === parsed.alt
  )
}

export function formatHotkey(combo: string, platform: HotkeyPlatform): string {
  const parsed = parseCombo(combo)
  const key = parsed.key.length === 1 ? parsed.key.toUpperCase() : parsed.key.charAt(0).toUpperCase() + parsed.key.slice(1)
  if (platform === 'mac') {
    return `${parsed.ctrl ? '⌃' : ''}${parsed.mod ? '⌘' : ''}${parsed.alt ? '⌥' : ''}${parsed.shift ? '⇧' : ''}${key}`
  }
  return [parsed.mod || parsed.ctrl ? 'Ctrl' : '', parsed.alt ? 'Alt' : '', parsed.shift ? 'Shift' : '', key].filter(Boolean).join('+')
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

export interface HotkeyBinding {
  combo: string
  handler: (event: KeyboardEvent) => void
}

interface HotkeysOptions {
  enabled?: boolean
  platform?: HotkeyPlatform
}

/** Registers a table of window shortcuts at once; the bindings are re-registered only when their combos change. */
export function useHotkeys(bindings: readonly HotkeyBinding[], options: HotkeysOptions = {}): void {
  const { enabled = true, platform } = options
  const bindingsRef = useRef(bindings)
  const combos = bindings.map((binding) => binding.combo).join('\n')

  useEffect(() => {
    bindingsRef.current = bindings
  }, [bindings])

  useEffect(() => {
    if (!enabled || combos === '') {
      return
    }
    const resolved = platform ?? detectPlatform()
    const unregister = combos.split('\n').map((combo, index) =>
      register(window, { combo, platform: resolved, handler: { current: (event: KeyboardEvent) => bindingsRef.current[index]?.handler(event) } }),
    )
    return () => {
      for (const undo of unregister) {
        undo()
      }
    }
  }, [combos, enabled, platform])
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
