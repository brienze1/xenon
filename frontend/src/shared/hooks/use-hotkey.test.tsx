import { fireEvent, render, renderHook, screen } from '@testing-library/react'
import { useRef } from 'react'
import { describe, expect, it, vi } from 'vitest'
import { detectPlatform, formatHotkey, matchesHotkey, useHotkey, useHotkeys } from './use-hotkey'

function keydown(init: KeyboardEventInit): KeyboardEvent {
  return new KeyboardEvent('keydown', init)
}

describe('matchesHotkey', () => {
  it('maps Mod to Cmd on macOS and Ctrl elsewhere', () => {
    expect(matchesHotkey(keydown({ key: 's', metaKey: true }), 'Mod+S', 'mac')).toBe(true)
    expect(matchesHotkey(keydown({ key: 's', ctrlKey: true }), 'Mod+S', 'mac')).toBe(false)
    expect(matchesHotkey(keydown({ key: 's', ctrlKey: true }), 'Mod+S', 'other')).toBe(true)
    expect(matchesHotkey(keydown({ key: 's', metaKey: true }), 'Mod+S', 'other')).toBe(false)
  })

  it('requires the exact modifiers', () => {
    expect(matchesHotkey(keydown({ key: 'S', metaKey: true, shiftKey: true }), 'Mod+Shift+S', 'mac')).toBe(true)
    expect(matchesHotkey(keydown({ key: 'S', metaKey: true, shiftKey: true }), 'Mod+S', 'mac')).toBe(false)
    expect(matchesHotkey(keydown({ key: 's', metaKey: true }), 'Mod+Shift+S', 'mac')).toBe(false)
    expect(matchesHotkey(keydown({ key: 'Enter', ctrlKey: true, altKey: true }), 'Mod+Enter', 'other')).toBe(false)
    expect(matchesHotkey(keydown({ key: 'Enter', ctrlKey: true, altKey: true }), 'Mod+Alt+Enter', 'other')).toBe(true)
    expect(matchesHotkey(keydown({ key: 't' }), 'Mod+T', 'other')).toBe(false)
    expect(matchesHotkey(keydown({ key: 'Escape' }), 'Escape', 'other')).toBe(true)
  })
})

describe('matchesHotkey with physical keys', () => {
  it('matches event.code and falls back to key without one', () => {
    expect(matchesHotkey(keydown({ key: 'z', code: 'KeyY', ctrlKey: true }), 'Mod+Y', 'other')).toBe(true)
    expect(matchesHotkey(keydown({ key: 'y', code: 'KeyZ', ctrlKey: true }), 'Mod+Y', 'other')).toBe(false)
    expect(matchesHotkey(keydown({ key: 'Enter', code: 'NumpadEnter', metaKey: true }), 'Mod+Enter', 'mac')).toBe(true)
    expect(matchesHotkey(keydown({ key: '}', code: 'BracketRight', metaKey: true, shiftKey: true }), 'Mod+Shift+]', 'mac')).toBe(true)
    expect(matchesHotkey(keydown({ key: ',', code: 'Comma', ctrlKey: true }), 'Mod+,', 'other')).toBe(true)
    expect(matchesHotkey(keydown({ key: 'F2', code: 'F2' }), 'F2', 'other')).toBe(true)
  })

  it('treats Ctrl as Control on every OS', () => {
    expect(matchesHotkey(keydown({ key: 'Tab', code: 'Tab', ctrlKey: true }), 'Ctrl+Tab', 'mac')).toBe(true)
    expect(matchesHotkey(keydown({ key: 'Tab', code: 'Tab', metaKey: true }), 'Ctrl+Tab', 'mac')).toBe(false)
    expect(matchesHotkey(keydown({ key: 'Tab', code: 'Tab', ctrlKey: true }), 'Ctrl+Tab', 'other')).toBe(true)
    expect(matchesHotkey(keydown({ key: 'Tab', code: 'Tab', ctrlKey: true, metaKey: true }), 'Ctrl+Tab', 'other')).toBe(false)
  })

  it('never matches while AltGr is held', () => {
    expect(matchesHotkey(keydown({ key: '@', code: 'KeyQ', ctrlKey: true, altKey: true, modifierAltGraph: true }), 'Mod+Alt+Q', 'other')).toBe(false)
  })
})

describe('formatHotkey', () => {
  it('renders platform labels', () => {
    expect(formatHotkey('Mod+Shift+S', 'mac')).toBe('⌘⇧S')
    expect(formatHotkey('Mod+Alt+Enter', 'mac')).toBe('⌘⌥Enter')
    expect(formatHotkey('Mod+Shift+S', 'other')).toBe('Ctrl+Shift+S')
    expect(formatHotkey('Mod+Alt+enter', 'other')).toBe('Ctrl+Alt+Enter')
    expect(formatHotkey('Ctrl+Shift+Tab', 'mac')).toBe('⌃⇧Tab')
    expect(formatHotkey('Ctrl+Shift+Tab', 'other')).toBe('Ctrl+Shift+Tab')
  })
})

describe('detectPlatform', () => {
  it('recognises Apple platforms', () => {
    expect(detectPlatform('MacIntel')).toBe('mac')
    expect(detectPlatform('Win32')).toBe('other')
    expect(detectPlatform('Linux x86_64')).toBe('other')
    expect(['mac', 'other']).toContain(detectPlatform())
  })
})

describe('useHotkey', () => {
  it('runs the handler for matching window key presses and prevents the default', () => {
    const handler = vi.fn()
    const { unmount } = renderHook(() => useHotkey('Mod+T', handler, { platform: 'other' }))
    const event = new KeyboardEvent('keydown', { key: 't', ctrlKey: true, cancelable: true })
    window.dispatchEvent(event)
    expect(handler).toHaveBeenCalledOnce()
    expect(event.defaultPrevented).toBe(true)
    fireEvent.keyDown(window, { key: 'w', ctrlKey: true })
    expect(handler).toHaveBeenCalledOnce()
    unmount()
    fireEvent.keyDown(window, { key: 't', ctrlKey: true })
    expect(handler).toHaveBeenCalledOnce()
  })

  it('shares one keydown listener per target across hook instances', () => {
    const addListener = vi.spyOn(window, 'addEventListener')
    const removeListener = vi.spyOn(window, 'removeEventListener')
    const first = vi.fn()
    const second = vi.fn()
    const { unmount } = renderHook(() => {
      useHotkey('Mod+T', first, { platform: 'other' })
      useHotkey('Mod+W', second, { platform: 'other' })
    })
    expect(addListener.mock.calls.filter(([type]) => type === 'keydown')).toHaveLength(1)
    fireEvent.keyDown(window, { key: 't', ctrlKey: true })
    fireEvent.keyDown(window, { key: 'w', ctrlKey: true })
    expect(first).toHaveBeenCalledOnce()
    expect(second).toHaveBeenCalledOnce()
    unmount()
    expect(removeListener.mock.calls.filter(([type]) => type === 'keydown')).toHaveLength(1)
    addListener.mockRestore()
    removeListener.mockRestore()
  })

  it('does nothing while disabled', () => {
    const handler = vi.fn()
    renderHook(() => useHotkey('Mod+T', handler, { platform: 'other', enabled: false }))
    fireEvent.keyDown(window, { key: 't', ctrlKey: true })
    expect(handler).not.toHaveBeenCalled()
  })

  it('listens only inside its scope element', () => {
    const handler = vi.fn()
    function Scoped() {
      const ref = useRef<HTMLDivElement>(null)
      useHotkey('Mod+F', handler, { scope: ref, platform: 'mac' })
      return (
        <div>
          <div ref={ref} tabIndex={-1}>
            body
          </div>
          <p>outside</p>
        </div>
      )
    }
    render(<Scoped />)
    fireEvent.keyDown(screen.getByText('outside'), { key: 'f', metaKey: true })
    expect(handler).not.toHaveBeenCalled()
    fireEvent.keyDown(screen.getByText('body'), { key: 'f', metaKey: true })
    expect(handler).toHaveBeenCalledOnce()
  })

  it('skips a scope that is not mounted', () => {
    const handler = vi.fn()
    renderHook(() => useHotkey('Mod+F', handler, { scope: { current: null }, platform: 'mac' }))
    fireEvent.keyDown(window, { key: 'f', metaKey: true })
    expect(handler).not.toHaveBeenCalled()
  })
})

describe('useHotkeys', () => {
  it('registers a table of combos and follows handler changes without re-registering', () => {
    const first = vi.fn()
    const second = vi.fn()
    const addListener = vi.spyOn(window, 'addEventListener')
    const { rerender, unmount } = renderHook(({ handler }) => useHotkeys([{ combo: 'Mod+K', handler }, { combo: 'Ctrl+Tab', handler }], { platform: 'other' }), { initialProps: { handler: first } })
    fireEvent.keyDown(window, { key: 'k', code: 'KeyK', ctrlKey: true })
    expect(first).toHaveBeenCalledOnce()
    rerender({ handler: second })
    fireEvent.keyDown(window, { key: 'Tab', code: 'Tab', ctrlKey: true })
    expect(second).toHaveBeenCalledOnce()
    expect(addListener.mock.calls.filter(([type]) => type === 'keydown')).toHaveLength(1)
    unmount()
    fireEvent.keyDown(window, { key: 'k', code: 'KeyK', ctrlKey: true })
    expect(first).toHaveBeenCalledOnce()
    addListener.mockRestore()
  })

  it('registers nothing while disabled or empty', () => {
    const handler = vi.fn()
    renderHook(() => useHotkeys([{ combo: 'Mod+K', handler }], { platform: 'other', enabled: false }))
    renderHook(() => useHotkeys([], { platform: 'other' }))
    fireEvent.keyDown(window, { key: 'k', code: 'KeyK', ctrlKey: true })
    expect(handler).not.toHaveBeenCalled()
  })
})
