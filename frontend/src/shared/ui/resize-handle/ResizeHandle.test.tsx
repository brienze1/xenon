import { fireEvent, render, screen } from '@testing-library/react'
import { describe, expect, it, vi } from 'vitest'
import { ResizeHandle } from './ResizeHandle'

describe('ResizeHandle', () => {
  it('ignores keys that do not resize and rounds dragged sizes', () => {
    const onChange = vi.fn()
    render(<ResizeHandle label="Resize" orientation="horizontal" edge="end" value={100} min={50} max={300} onChange={onChange} />)
    const handle = screen.getByRole('separator', { name: 'Resize' })
    fireEvent.keyDown(handle, { key: 'ArrowLeft' })
    fireEvent.keyDown(handle, { key: 'a' })
    expect(onChange).not.toHaveBeenCalled()
    fireEvent.keyDown(handle, { key: 'ArrowDown' })
    expect(onChange).toHaveBeenLastCalledWith(116)
    fireEvent.keyDown(handle, { key: 'ArrowUp' })
    expect(onChange).toHaveBeenLastCalledWith(84)
    fireEvent.pointerDown(handle, { clientX: 0, clientY: 10 })
    fireEvent.pointerMove(window, { clientX: 999, clientY: 20.4 })
    expect(onChange).toHaveBeenLastCalledWith(110)
    fireEvent.pointerUp(window)
  })

  it('follows the pointer one to one when the UI is zoomed', () => {
    document.body.style.setProperty('zoom', '1.25')
    const onChange = vi.fn()
    render(<ResizeHandle label="Resize" orientation="vertical" edge="end" value={100} min={50} max={300} onChange={onChange} />)
    fireEvent.pointerDown(screen.getByRole('separator', { name: 'Resize' }), { clientX: 0, clientY: 0 })
    fireEvent.pointerMove(window, { clientX: 50, clientY: 0 })
    expect(onChange).toHaveBeenLastCalledWith(140)
    fireEvent.pointerUp(window)
    document.body.style.removeProperty('zoom')
  })
})
