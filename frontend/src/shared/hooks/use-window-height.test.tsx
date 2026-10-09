import { act, renderHook } from '@testing-library/react'
import { afterEach, describe, expect, it } from 'vitest'
import { useWindowHeight } from './use-window-height'

const initialHeight = window.innerHeight

function resizeTo(height: number) {
  Object.defineProperty(window, 'innerHeight', { configurable: true, value: height })
  window.dispatchEvent(new Event('resize'))
}

describe('useWindowHeight', () => {
  afterEach(() => resizeTo(initialHeight))

  it('follows the window height as it resizes', () => {
    const { result } = renderHook(() => useWindowHeight())
    expect(result.current).toBe(window.innerHeight)
    act(() => resizeTo(640))
    expect(result.current).toBe(640)
  })
})
