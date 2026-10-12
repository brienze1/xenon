import { afterEach, describe, expect, it, vi } from 'vitest'

const listeners: { zoom?: (direction: 'in' | 'out' | 'reset') => void } = {}
vi.mock('@/platform/events', () => ({
  onViewZoomRequested: vi.fn((callback) => {
    listeners.zoom = callback
    return () => undefined
  }),
}))

const { applyZoom, installZoom, nextZoom } = await import('./zoom')

describe('zoom', () => {
  afterEach(() => {
    window.localStorage.clear()
    document.body.removeAttribute('style')
  })

  it('steps through browser-like levels and stops at both ends', () => {
    expect(nextZoom(1, 'in')).toBe(1.1)
    expect(nextZoom(1, 'out')).toBe(0.9)
    expect(nextZoom(2, 'in')).toBe(2)
    expect(nextZoom(0.67, 'out')).toBe(0.67)
    expect(nextZoom(1.5, 'reset')).toBe(1)
  })

  it('sizes the body to the window divided by the zoom so the scaled UI fits without scrollbars', () => {
    vi.stubGlobal('innerWidth', 1000)
    vi.stubGlobal('innerHeight', 500)
    applyZoom(1.25)
    expect(document.body.style.getPropertyValue('zoom')).toBe('1.25')
    expect(document.body.style.width).toBe('800px')
    expect(document.body.style.height).toBe('400px')
    vi.unstubAllGlobals()
  })

  it('follows the View menu and Ctrl shortcuts and remembers the level', () => {
    const uninstall = installZoom('other')
    listeners.zoom?.('in')
    expect(document.body.style.getPropertyValue('zoom')).toBe('1.1')
    window.dispatchEvent(new KeyboardEvent('keydown', { key: '=', ctrlKey: true }))
    expect(document.body.style.getPropertyValue('zoom')).toBe('1.25')
    window.dispatchEvent(new KeyboardEvent('keydown', { key: '0', ctrlKey: true }))
    expect(document.body.style.getPropertyValue('zoom')).toBe('1')
    listeners.zoom?.('out')
    uninstall()
    expect(window.localStorage.getItem('xenon.zoom')).toBe('0.9')
    installZoom('mac')()
    expect(document.body.style.getPropertyValue('zoom')).toBe('0.9')
  })
})
