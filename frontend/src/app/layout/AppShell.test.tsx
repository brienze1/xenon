import { render, screen } from '@testing-library/react'
import { describe, expect, it } from 'vitest'
import { AppShell } from './AppShell'
import { ConsoleDrawer } from './regions/ConsoleDrawer'
import { Footer } from './regions/Footer'
import { MainArea } from './regions/MainArea'
import { RightPanel } from './regions/RightPanel'
import { Sidebar } from './regions/Sidebar'
import { TopBar } from './regions/TopBar'

describe('AppShell', () => {
  it('places every window region as a landmark, with the console hidden by default', () => {
    render(
      <AppShell>
        <TopBar>switcher</TopBar>
        <Sidebar>tree</Sidebar>
        <MainArea>tabs</MainArea>
        <RightPanel>info</RightPanel>
        <ConsoleDrawer open={false}>logs</ConsoleDrawer>
        <Footer>status</Footer>
      </AppShell>,
    )
    expect(screen.getByRole('banner')).toHaveTextContent('switcher')
    expect(screen.getByRole('navigation', { name: 'Sidebar' })).toHaveTextContent('tree')
    expect(screen.getByRole('main')).toHaveTextContent('tabs')
    expect(screen.getByRole('complementary', { name: 'Side panel' })).toHaveTextContent('info')
    expect(screen.getByRole('contentinfo')).toHaveTextContent('status')
    expect(screen.queryByRole('region', { name: 'Console' })).not.toBeInTheDocument()
  })

  it('shows the console drawer when open', () => {
    render(<ConsoleDrawer open>logs</ConsoleDrawer>)
    expect(screen.getByRole('region', { name: 'Console' })).toBeVisible()
  })
})
