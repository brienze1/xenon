import { fireEvent, renderHook } from '@testing-library/react'
import { describe, expect, it, vi } from 'vitest'
import { detectPlatform } from '@/shared/hooks/use-hotkey'
import { HOTKEY_BY_ACTION, HOTKEYS } from './hotkeys'
import { useAppHotkeys } from './use-app-hotkeys'

const modifier = detectPlatform() === 'mac' ? { metaKey: true } : { ctrlKey: true }

describe('hotkey table', () => {
  it('lists the shortcuts (XEN-001 FR-042, XEN-002 FR-028) once each', () => {
    expect(HOTKEYS.map((hotkey) => [hotkey.action, hotkey.combo])).toEqual([
      ['send', 'Mod+Enter'],
      ['save', 'Mod+S'],
      ['saveAs', 'Mod+Shift+S'],
      ['newTab', 'Mod+T'],
      ['closeTab', 'Mod+W'],
      ['environmentSwitcher', 'Mod+Shift+E'],
    ])
    expect(new Set(HOTKEYS.map((hotkey) => hotkey.combo)).size).toBe(HOTKEYS.length)
    expect(HOTKEY_BY_ACTION.get('saveAs')?.combo).toBe('Mod+Shift+S')
    expect(HOTKEY_BY_ACTION.size).toBe(HOTKEYS.length)
  })

  it('routes each shortcut to its action handler', () => {
    const handlers = { send: vi.fn(), save: vi.fn(), saveAs: vi.fn(), newTab: vi.fn(), closeTab: vi.fn(), environmentSwitcher: vi.fn() }
    renderHook(() => useAppHotkeys(handlers))
    fireEvent.keyDown(window, { key: 'Enter', ...modifier })
    fireEvent.keyDown(window, { key: 's', ...modifier })
    fireEvent.keyDown(window, { key: 'S', shiftKey: true, ...modifier })
    fireEvent.keyDown(window, { key: 't', ...modifier })
    fireEvent.keyDown(window, { key: 'w', ...modifier })
    fireEvent.keyDown(window, { key: 'E', shiftKey: true, ...modifier })
    for (const handler of Object.values(handlers)) {
      expect(handler).toHaveBeenCalledOnce()
    }
  })

  it('leaves shortcuts without a handler alone', () => {
    const newTab = vi.fn()
    renderHook(() => useAppHotkeys({ newTab }))
    const event = new KeyboardEvent('keydown', { key: 's', cancelable: true, ...modifier })
    window.dispatchEvent(event)
    expect(event.defaultPrevented).toBe(false)
    expect(newTab).not.toHaveBeenCalled()
  })
})
