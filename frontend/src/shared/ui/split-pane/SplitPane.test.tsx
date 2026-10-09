import { fireEvent, render, screen } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { useState } from 'react'
import { describe, expect, it, vi } from 'vitest'
import { clampRatio, SplitPane, type SplitDirection } from './SplitPane'

function Example({ direction, initial, onChange }: { direction: SplitDirection; initial: number; onChange?: (ratio: number) => void }) {
  const [ratio, setRatio] = useState(initial)
  return (
    <SplitPane.Root
      direction={direction}
      ratio={ratio}
      onRatioChange={(next) => {
        setRatio(next)
        onChange?.(next)
      }}
    >
      <SplitPane.Pane>request</SplitPane.Pane>
      <SplitPane.Divider label="Resize request and response" />
      <SplitPane.Pane>response</SplitPane.Pane>
    </SplitPane.Root>
  )
}

function mockBounds(element: HTMLElement) {
  vi.spyOn(element, 'getBoundingClientRect').mockReturnValue({
    left: 100,
    top: 50,
    width: 400,
    height: 200,
    right: 500,
    bottom: 250,
    x: 100,
    y: 50,
    toJSON: () => ({}),
  })
}

describe('SplitPane', () => {
  it('exposes the ratio on an accessible separator and the root style', () => {
    render(<Example direction="row" initial={0.5} />)
    const separator = screen.getByRole('separator', { name: 'Resize request and response' })
    expect(separator).toHaveAttribute('aria-orientation', 'vertical')
    expect(separator).toHaveAttribute('aria-valuenow', '50')
    expect(separator).toHaveAttribute('aria-valuemin', '20')
    expect(separator).toHaveAttribute('aria-valuemax', '80')
    expect(separator.parentElement).toHaveStyle({ '--split-ratio': '0.5' })
  })

  it('clamps an out-of-range ratio', () => {
    render(<Example direction="column" initial={0.95} />)
    const separator = screen.getByRole('separator')
    expect(separator).toHaveAttribute('aria-orientation', 'horizontal')
    expect(separator).toHaveAttribute('aria-valuenow', '80')
  })

  it('resizes with the keyboard within the limits', async () => {
    const onChange = vi.fn()
    render(<Example direction="row" initial={0.5} onChange={onChange} />)
    screen.getByRole('separator').focus()
    await userEvent.keyboard('{ArrowRight}')
    expect(onChange).toHaveBeenLastCalledWith(0.55)
    await userEvent.keyboard('{ArrowUp}{ArrowLeft}')
    expect(onChange).toHaveBeenLastCalledWith(0.45)
    await userEvent.keyboard('{End}')
    expect(onChange).toHaveBeenLastCalledWith(0.8)
    await userEvent.keyboard('{ArrowDown}')
    expect(onChange).toHaveBeenLastCalledWith(0.8)
    await userEvent.keyboard('{Home}')
    expect(onChange).toHaveBeenLastCalledWith(0.2)
    await userEvent.keyboard('{Enter}')
    expect(onChange).toHaveBeenCalledTimes(6)
  })

  it('follows the pointer while dragging, side by side and stacked', () => {
    const onChange = vi.fn()
    const { unmount } = render(<Example direction="row" initial={0.5} onChange={onChange} />)
    const separator = screen.getByRole('separator')
    mockBounds(separator.parentElement as HTMLElement)
    fireEvent.pointerDown(separator)
    fireEvent.pointerMove(window, { clientX: 220 })
    expect(onChange).toHaveBeenLastCalledWith(0.3)
    fireEvent.pointerMove(window, { clientX: 490 })
    expect(onChange).toHaveBeenLastCalledWith(0.8)
    fireEvent.pointerUp(window)
    fireEvent.pointerMove(window, { clientX: 300 })
    expect(onChange).toHaveBeenCalledTimes(2)
    unmount()

    render(<Example direction="column" initial={0.5} onChange={onChange} />)
    const stacked = screen.getByRole('separator')
    mockBounds(stacked.parentElement as HTMLElement)
    fireEvent.pointerDown(stacked)
    fireEvent.pointerMove(window, { clientY: 170 })
    expect(onChange).toHaveBeenLastCalledWith(0.6)
    fireEvent.pointerUp(window)
  })

  it('clampRatio keeps values in range', () => {
    expect(clampRatio(0.1, 0.2, 0.8)).toBe(0.2)
    expect(clampRatio(0.9, 0.2, 0.8)).toBe(0.8)
    expect(clampRatio(0.4, 0.2, 0.8)).toBe(0.4)
  })
})
