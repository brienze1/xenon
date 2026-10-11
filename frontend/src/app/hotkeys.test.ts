import { renderHook } from '@testing-library/react'
import { describe, expect, it, vi } from 'vitest'
import { formatHotkey, type HotkeyPlatform } from '@/shared/hooks/use-hotkey'
import { HOTKEY_BY_ACTION, HOTKEYS, type HotkeyAction, type HotkeyHandlers } from './hotkeys'
import { useAppHotkeys } from './use-app-hotkeys'

const FR_012: Array<[HotkeyAction, string[]]> = [
  ['send', ['Mod+Enter']],
  ['save', ['Mod+S']],
  ['saveAs', ['Mod+Shift+S']],
  ['newTab', ['Mod+T']],
  ['closeTab', ['Mod+W']],
  ['forceCloseTab', ['Mod+Alt+W']],
  ['reopenClosedTab', ['Mod+Shift+T']],
  ['nextTab', ['Ctrl+Tab', 'Mod+Shift+]']],
  ['previousTab', ['Ctrl+Shift+Tab', 'Mod+Shift+[']],
  ['goToTab1', ['Mod+1']],
  ['goToTab2', ['Mod+2']],
  ['goToTab3', ['Mod+3']],
  ['goToTab4', ['Mod+4']],
  ['goToTab5', ['Mod+5']],
  ['goToTab6', ['Mod+6']],
  ['goToTab7', ['Mod+7']],
  ['goToTab8', ['Mod+8']],
  ['lastTab', ['Mod+9']],
  ['focusUrl', ['Mod+L']],
  ['searchRequests', ['Mod+K']],
  ['findInResponse', ['Mod+F']],
  ['toggleSidebar', ['Mod+\\']],
  ['console', ['Mod+Alt+C']],
  ['renameSelected', ['Mod+E']],
  ['duplicateSelected', ['Mod+D']],
  ['import', ['Mod+O']],
  ['settings', ['Mod+,']],
  ['environmentSwitcher', ['Mod+Shift+E']],
]

const PROFILES: Array<{ name: string; platform: HotkeyPlatform }> = [
  { name: 'macOS', platform: 'mac' },
  { name: 'Windows', platform: 'other' },
  { name: 'Linux', platform: 'other' },
]

const CODES: Record<string, string> = { enter: 'Enter', tab: 'Tab', '\\': 'Backslash', ',': 'Comma', '[': 'BracketLeft', ']': 'BracketRight' }

function codeOf(key: string): string {
  if (/^[a-z]$/.test(key)) {
    return `Key${key.toUpperCase()}`
  }
  if (/^[0-9]$/.test(key)) {
    return `Digit${key}`
  }
  return CODES[key]!
}

/** A key press for the combo whose `key` is what an unusual layout would type, so only `code` can match. */
function pressFor(combo: string, platform: HotkeyPlatform, extra: KeyboardEventInit = {}): KeyboardEvent {
  const parts = combo.toLowerCase().split('+')
  const key = parts.at(-1)!
  const mod = parts.includes('mod')
  const ctrl = parts.includes('ctrl')
  return new KeyboardEvent('keydown', {
    key: 'Unidentified',
    code: codeOf(key),
    metaKey: platform === 'mac' && mod,
    ctrlKey: platform === 'mac' ? ctrl : mod || ctrl,
    shiftKey: parts.includes('shift'),
    altKey: parts.includes('alt'),
    cancelable: true,
    ...extra,
  })
}

function allHandlers(): Record<HotkeyAction, ReturnType<typeof vi.fn>> {
  return Object.fromEntries(HOTKEYS.map((hotkey) => [hotkey.action, vi.fn()])) as Record<HotkeyAction, ReturnType<typeof vi.fn>>
}

describe('hotkey table', () => {
  it('lists every FR-012 shortcut once, in order', () => {
    expect(HOTKEYS.map((hotkey) => [hotkey.action, [...hotkey.combos]])).toEqual(FR_012)
    expect(HOTKEY_BY_ACTION.size).toBe(HOTKEYS.length)
    expect(HOTKEY_BY_ACTION.get('nextTab')?.label).toBe('Next tab')
    for (const hotkey of HOTKEYS) {
      expect(hotkey.label).not.toBe('')
    }
  })

  it.each(PROFILES)('gives every shortcut its own key combination on $name', ({ platform }) => {
    const shown = HOTKEYS.flatMap((hotkey) => hotkey.combos.map((combo) => formatHotkey(combo, platform)))
    expect(new Set(shown).size).toBe(shown.length)
  })
})

describe.each(PROFILES)('app hotkeys on $name', ({ platform }) => {
  it.each(HOTKEYS.flatMap((hotkey) => hotkey.combos.map((combo) => [hotkey.action, combo] as const)))('%s runs on %s, matched by the physical key', (action, combo) => {
    const handlers = allHandlers()
    renderHook(() => useAppHotkeys(handlers, platform))
    const event = pressFor(combo, platform)
    window.dispatchEvent(event)
    expect(handlers[action]).toHaveBeenCalledOnce()
    expect(event.defaultPrevented).toBe(true)
    for (const [other, handler] of Object.entries(handlers)) {
      if (other !== action) {
        expect(handler).not.toHaveBeenCalled()
      }
    }
  })

  it('ignores the character a layout types and uses the key position', () => {
    const handlers = allHandlers()
    renderHook(() => useAppHotkeys(handlers, platform))
    const modifier = platform === 'mac' ? { metaKey: true } : { ctrlKey: true }
    window.dispatchEvent(new KeyboardEvent('keydown', { key: 'k', code: 'KeyV', cancelable: true, ...modifier }))
    expect(handlers.searchRequests).not.toHaveBeenCalled()
    window.dispatchEvent(new KeyboardEvent('keydown', { key: 'л', code: 'KeyK', cancelable: true, ...modifier }))
    expect(handlers.searchRequests).toHaveBeenCalledOnce()
  })

  it('needs the exact modifiers', () => {
    const handlers = allHandlers()
    renderHook(() => useAppHotkeys(handlers, platform))
    window.dispatchEvent(new KeyboardEvent('keydown', { key: 'k', code: 'KeyK', cancelable: true }))
    window.dispatchEvent(pressFor('Mod+K', platform, { shiftKey: true }))
    window.dispatchEvent(pressFor('Mod+K', platform, { altKey: true }))
    expect(handlers.searchRequests).not.toHaveBeenCalled()
  })
})

describe('tab cycling', () => {
  it.each(PROFILES)('Ctrl+Tab and Cmd/Ctrl+Shift+] both go to the next tab, Ctrl+Shift+Tab and Cmd/Ctrl+Shift+[ to the previous on $name', ({ platform }) => {
    const nextTab = vi.fn()
    const previousTab = vi.fn()
    renderHook(() => useAppHotkeys({ nextTab, previousTab }, platform))
    window.dispatchEvent(new KeyboardEvent('keydown', { key: 'Tab', code: 'Tab', ctrlKey: true, cancelable: true }))
    window.dispatchEvent(pressFor('Mod+Shift+]', platform, { key: '}' }))
    window.dispatchEvent(new KeyboardEvent('keydown', { key: 'Tab', code: 'Tab', ctrlKey: true, shiftKey: true, cancelable: true }))
    window.dispatchEvent(pressFor('Mod+Shift+[', platform, { key: '{' }))
    expect(nextTab).toHaveBeenCalledTimes(2)
    expect(previousTab).toHaveBeenCalledTimes(2)
  })

  it('does not take Cmd+Tab on macOS, which belongs to the system', () => {
    const nextTab = vi.fn()
    renderHook(() => useAppHotkeys({ nextTab }, 'mac'))
    window.dispatchEvent(new KeyboardEvent('keydown', { key: 'Tab', code: 'Tab', metaKey: true, cancelable: true }))
    expect(nextTab).not.toHaveBeenCalled()
  })
})

describe('AltGr typing', () => {
  it.each([
    ['@ on a German layout', 'KeyQ', '@'],
    ['ć on a Polish layout', 'KeyC', 'ć'],
    ['∑ typed with AltGr+W', 'KeyW', '∑'],
    ['{ on a German layout', 'Digit7', '{'],
  ])('never triggers a Ctrl+Alt shortcut while typing %s', (_case, code, key) => {
    const handlers = allHandlers()
    renderHook(() => useAppHotkeys(handlers, 'other'))
    const event = new KeyboardEvent('keydown', { key, code, ctrlKey: true, altKey: true, modifierAltGraph: true, cancelable: true })
    window.dispatchEvent(event)
    for (const handler of Object.values(handlers)) {
      expect(handler).not.toHaveBeenCalled()
    }
    expect(event.defaultPrevented).toBe(false)
  })
})

describe('missing handlers', () => {
  it('leaves shortcuts without a handler to the webview', () => {
    const handlers: HotkeyHandlers = { newTab: vi.fn() }
    renderHook(() => useAppHotkeys(handlers, 'other'))
    const event = pressFor('Mod+S', 'other')
    window.dispatchEvent(event)
    expect(event.defaultPrevented).toBe(false)
    expect(handlers.newTab).not.toHaveBeenCalled()
  })
})
