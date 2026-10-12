import { fireEvent, render, screen, within } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { useState } from 'react'
import { describe, expect, it, vi } from 'vitest'
import { defaultLayout, type LayoutState } from '@/platform/mappers/ui-state'
import { AppShell } from './AppShell'
import { ConsoleDrawer } from './regions/ConsoleDrawer'
import { Footer } from './regions/Footer'
import { MainArea } from './regions/MainArea'
import { RightPanel } from './regions/RightPanel'
import { Sidebar } from './regions/Sidebar'
import { TopBar } from './regions/TopBar'

type LayoutPatch = Partial<LayoutState>

interface ShellProps {
  layout: LayoutState
  onLayoutChange: (patch: LayoutPatch) => void
  sidePanelApplies?: boolean
}

function Shell({ layout, onLayoutChange, sidePanelApplies = true }: ShellProps) {
  return (
    <AppShell layout={layout} onLayoutChange={onLayoutChange}>
      <TopBar>switcher</TopBar>
      <Sidebar>tree</Sidebar>
      <MainArea>tabs</MainArea>
      <RightPanel applies={sidePanelApplies}>info</RightPanel>
      <ConsoleDrawer>logs</ConsoleDrawer>
      <Footer>status</Footer>
    </AppShell>
  )
}

function renderShell(overrides: LayoutPatch = {}) {
  const onLayoutChange = vi.fn()
  render(<Shell layout={{ ...defaultLayout(), ...overrides }} onLayoutChange={onLayoutChange} />)
  return onLayoutChange
}

function StatefulShell({ sidePanelApplies = true }: { sidePanelApplies?: boolean }) {
  const [layout, setLayout] = useState(defaultLayout())
  return <Shell layout={layout} sidePanelApplies={sidePanelApplies} onLayoutChange={(patch) => setLayout((current) => ({ ...current, ...patch }))} />
}

describe('AppShell', () => {
  it('places every window region as a landmark, with the console hidden by default', () => {
    renderShell()
    expect(screen.getByRole('banner')).toHaveTextContent('switcher')
    expect(screen.getByRole('navigation', { name: 'Sidebar' })).toHaveTextContent('tree')
    expect(screen.getByRole('main')).toHaveTextContent('tabs')
    expect(screen.getByRole('complementary', { name: 'Side panel' })).toHaveTextContent('info')
    expect(screen.getByRole('contentinfo')).toHaveTextContent('status')
    expect(screen.queryByRole('region', { name: 'Console' })).not.toBeInTheDocument()
  })

  it('sizes the regions from the layout', () => {
    renderShell({ sidebarWidth: 320, rightPanelWidth: 300, consoleHeight: 240, consoleVisible: true })
    expect(screen.getByRole('separator', { name: 'Resize sidebar' })).toHaveAttribute('aria-valuenow', '320')
    expect(screen.getByRole('separator', { name: 'Resize side panel' })).toHaveAttribute('aria-valuenow', '300')
    expect(screen.getByRole('separator', { name: 'Resize console' })).toHaveAttribute('aria-valuenow', '240')
    expect(screen.getByRole('region', { name: 'Console' })).toHaveTextContent('logs')
  })

  it('hides the sidebar, side panel and console when they are switched off', () => {
    renderShell({ sidebarVisible: false, rightPanelVisible: false, consoleVisible: false })
    expect(screen.queryByRole('navigation', { name: 'Sidebar' })).not.toBeInTheDocument()
    expect(screen.queryByRole('complementary', { name: 'Side panel' })).not.toBeInTheDocument()
    expect(screen.queryByRole('separator')).not.toBeInTheDocument()
  })

  it('toggles each region and the request/response orientation from the footer', async () => {
    render(<StatefulShell />)
    const sidebar = screen.getByRole('button', { name: 'Sidebar' })
    expect(sidebar).toHaveAttribute('aria-pressed', 'true')
    await userEvent.click(sidebar)
    expect(screen.queryByRole('navigation', { name: 'Sidebar' })).not.toBeInTheDocument()
    expect(sidebar).toHaveAttribute('aria-pressed', 'false')
    await userEvent.click(screen.getByRole('button', { name: 'Side Panel' }))
    expect(screen.queryByRole('complementary', { name: 'Side panel' })).not.toBeInTheDocument()
    const console = screen.getByRole('button', { name: 'Console' })
    expect(console).toHaveAttribute('aria-pressed', 'false')
    await userEvent.click(console)
    expect(screen.getByRole('region', { name: 'Console' })).toBeVisible()
    const sideBySide = screen.getByRole('button', { name: 'Side by Side' })
    expect(sideBySide).toHaveAttribute('aria-pressed', 'false')
    await userEvent.click(sideBySide)
    expect(sideBySide).toHaveAttribute('aria-pressed', 'true')
  })

  it('hides the side panel from its own button and reopens it from the same place or the footer', async () => {
    render(<StatefulShell />)
    const panel = screen.getByRole('complementary', { name: 'Side panel' })
    await userEvent.click(within(panel).getByRole('button', { name: 'Hide side panel' }))
    expect(screen.queryByRole('complementary', { name: 'Side panel' })).not.toBeInTheDocument()
    await userEvent.click(screen.getByRole('button', { name: 'Show side panel' }))
    expect(screen.getByRole('complementary', { name: 'Side panel' })).toBeInTheDocument()
    expect(screen.queryByRole('button', { name: 'Show side panel' })).not.toBeInTheDocument()
    await userEvent.click(screen.getByRole('button', { name: 'Side Panel' }))
    expect(screen.queryByRole('complementary', { name: 'Side panel' })).not.toBeInTheDocument()
    await userEvent.click(screen.getByRole('button', { name: 'Side Panel' }))
    expect(screen.getByRole('complementary', { name: 'Side panel' })).toBeInTheDocument()
  })

  it('steps aside for tabs without a side panel and comes back for requests, unless it was hidden by hand', async () => {
    const { rerender } = render(<StatefulShell sidePanelApplies={false} />)
    expect(screen.queryByRole('complementary', { name: 'Side panel' })).not.toBeInTheDocument()
    expect(screen.queryByRole('button', { name: 'Show side panel' })).not.toBeInTheDocument()
    rerender(<StatefulShell sidePanelApplies />)
    const panel = screen.getByRole('complementary', { name: 'Side panel' })
    await userEvent.click(within(panel).getByRole('button', { name: 'Hide side panel' }))
    rerender(<StatefulShell sidePanelApplies={false} />)
    expect(screen.queryByRole('button', { name: 'Show side panel' })).not.toBeInTheDocument()
    rerender(<StatefulShell sidePanelApplies />)
    expect(screen.queryByRole('complementary', { name: 'Side panel' })).not.toBeInTheDocument()
    expect(screen.getByRole('button', { name: 'Show side panel' })).toBeInTheDocument()
  })

  it('reports the side-by-side switch as a layout change', async () => {
    const onLayoutChange = renderShell({ split: 'side-by-side' })
    await userEvent.click(screen.getByRole('button', { name: 'Side by Side' }))
    expect(onLayoutChange).toHaveBeenCalledWith({ split: 'stacked' })
    await userEvent.click(screen.getByRole('button', { name: 'Console' }))
    expect(onLayoutChange).toHaveBeenLastCalledWith({ consoleVisible: true })
  })

  it('resizes the sidebar by dragging or with the keyboard, within 200-600 px', () => {
    const onLayoutChange = renderShell({ sidebarWidth: 280 })
    const handle = screen.getByRole('separator', { name: 'Resize sidebar' })
    expect(handle).toHaveAttribute('aria-orientation', 'vertical')
    expect(handle).toHaveAttribute('aria-valuemin', '200')
    expect(handle).toHaveAttribute('aria-valuemax', '600')
    fireEvent.pointerDown(handle, { clientX: 280, clientY: 10 })
    fireEvent.pointerMove(window, { clientX: 330, clientY: 10 })
    expect(onLayoutChange).toHaveBeenLastCalledWith({ sidebarWidth: 330 })
    fireEvent.pointerMove(window, { clientX: 2000, clientY: 10 })
    expect(onLayoutChange).toHaveBeenLastCalledWith({ sidebarWidth: 600 })
    fireEvent.pointerUp(window)
    fireEvent.pointerMove(window, { clientX: 100, clientY: 10 })
    expect(onLayoutChange).toHaveBeenLastCalledWith({ sidebarWidth: 600 })
    fireEvent.keyDown(handle, { key: 'ArrowRight' })
    expect(onLayoutChange).toHaveBeenLastCalledWith({ sidebarWidth: 296 })
    fireEvent.keyDown(handle, { key: 'ArrowLeft' })
    expect(onLayoutChange).toHaveBeenLastCalledWith({ sidebarWidth: 264 })
    fireEvent.keyDown(handle, { key: 'Home' })
    expect(onLayoutChange).toHaveBeenLastCalledWith({ sidebarWidth: 200 })
    fireEvent.keyDown(handle, { key: 'End' })
    expect(onLayoutChange).toHaveBeenLastCalledWith({ sidebarWidth: 600 })
  })

  it('resizes the side panel from its left edge, within 240-600 px', () => {
    const onLayoutChange = renderShell({ rightPanelWidth: 320 })
    const handle = screen.getByRole('separator', { name: 'Resize side panel' })
    fireEvent.pointerDown(handle, { clientX: 900, clientY: 10 })
    fireEvent.pointerMove(window, { clientX: 850, clientY: 10 })
    expect(onLayoutChange).toHaveBeenLastCalledWith({ rightPanelWidth: 370 })
    fireEvent.pointerMove(window, { clientX: 1200, clientY: 10 })
    expect(onLayoutChange).toHaveBeenLastCalledWith({ rightPanelWidth: 240 })
    fireEvent.pointerUp(window)
    fireEvent.keyDown(handle, { key: 'ArrowLeft' })
    expect(onLayoutChange).toHaveBeenLastCalledWith({ rightPanelWidth: 336 })
  })

  it('resizes the console from its top edge, between 120 px and 60% of the window', () => {
    const onLayoutChange = renderShell({ consoleVisible: true, consoleHeight: 200 })
    const handle = screen.getByRole('separator', { name: 'Resize console' })
    expect(handle).toHaveAttribute('aria-orientation', 'horizontal')
    expect(handle).toHaveAttribute('aria-valuemax', String(Math.round(window.innerHeight * 0.6)))
    fireEvent.pointerDown(handle, { clientX: 10, clientY: 500 })
    fireEvent.pointerMove(window, { clientX: 10, clientY: 440 })
    expect(onLayoutChange).toHaveBeenLastCalledWith({ consoleHeight: 260 })
    fireEvent.pointerMove(window, { clientX: 10, clientY: -5000 })
    expect(onLayoutChange).toHaveBeenLastCalledWith({ consoleHeight: Math.round(window.innerHeight * 0.6) })
    fireEvent.pointerMove(window, { clientX: 10, clientY: 5000 })
    expect(onLayoutChange).toHaveBeenLastCalledWith({ consoleHeight: 120 })
    fireEvent.pointerUp(window)
  })
})
