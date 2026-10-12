import { fireEvent, render, screen } from '@testing-library/react'
import { describe, expect, it, vi } from 'vitest'
import { defaultLayout, type LayoutState } from '@/platform/mappers/ui-state'

vi.mock('@/features/request-builder/components/RequestBuilder', () => ({ RequestBuilder: () => <div>builder</div> }))
vi.mock('@/features/response-viewer/components/ResponseViewer', () => ({ ResponseViewer: () => <div>viewer</div> }))

const { AppShell } = await import('./AppShell')
const { RequestWorkspace } = await import('./RequestWorkspace')

function renderWorkspace(overrides: Partial<LayoutState>) {
  const onLayoutChange = vi.fn()
  render(
    <AppShell layout={{ ...defaultLayout(), ...overrides }} onLayoutChange={onLayoutChange}>
      <RequestWorkspace tabId="t1" />
    </AppShell>,
  )
  return onLayoutChange
}

describe('RequestWorkspace', () => {
  it('stacks the request above the response and remembers the dragged ratio', () => {
    const onLayoutChange = renderWorkspace({ split: 'stacked', splitRatio: 0.4 })
    const divider = screen.getByRole('separator', { name: 'Resize request and response' })
    expect(divider).toHaveAttribute('aria-orientation', 'horizontal')
    expect(divider).toHaveAttribute('aria-valuenow', '40')
    fireEvent.keyDown(divider, { key: 'ArrowDown' })
    expect(onLayoutChange).toHaveBeenCalledWith({ splitRatio: 0.45 })
    expect(screen.getByText('builder')).toBeInTheDocument()
    expect(screen.getByText('viewer')).toBeInTheDocument()
  })

  it('places the request and response side by side', () => {
    renderWorkspace({ split: 'side-by-side', splitRatio: 0.7 })
    const divider = screen.getByRole('separator', { name: 'Resize request and response' })
    expect(divider).toHaveAttribute('aria-orientation', 'vertical')
    expect(divider).toHaveAttribute('aria-valuenow', '70')
  })
})
