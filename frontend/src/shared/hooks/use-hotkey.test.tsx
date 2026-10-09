import { fireEvent, render, renderHook, screen } from '@testing-library/react'
import { useRef } from 'react'
import { describe, expect, it, vi } from 'vitest'
import { detectPlatform, formatHotkey, matchesHotkey, useHotkey } from './use-hotkey'

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

describe('formatHotkey', () => {
  it('renders platform labels', () => {
    expect(formatHotkey('Mod+Shift+S', 'mac')).toBe('⌘⇧S')
    expect(formatHotkey('Mod+Alt+Enter', 'mac')).toBe('⌘⌥Enter')
    expect(formatHotkey('Mod+Shift+S', 'other')).toBe('Ctrl+Shift+S')
    expect(formatHotkey('Mod+Alt+enter', 'other')).toBe('Ctrl+Alt+Enter')
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
