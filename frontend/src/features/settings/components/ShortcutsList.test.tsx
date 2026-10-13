import { render, screen, within } from '@testing-library/react'
import { describe, expect, it } from 'vitest'
import { ShortcutsList } from './ShortcutsList'

const SHORTCUTS = [
  { label: 'Save', combos: ['Mod+S'] },
  { label: 'Next tab', combos: ['Ctrl+Tab', 'Mod+Shift+]'] },
  { label: 'Force close tab', combos: ['Mod+Alt+W'] },
  { label: 'Show or hide the sidebar', combos: ['Mod+\\'] },
]

function rows() {
  return within(screen.getByRole('table')).getAllByRole('row').slice(1).map((row) => row.textContent)
}

describe('ShortcutsList', () => {
  it('lists every shortcut with macOS key symbols', () => {
    render(<ShortcutsList shortcuts={SHORTCUTS} platform="mac" />)
    expect(screen.getByRole('heading', { name: 'Shortcuts' })).toBeInTheDocument()
    expect(rows()).toEqual(['Save⌘S', 'Next tab⌃Tab or ⌘⇧]', 'Force close tab⌘⌥W', 'Show or hide the sidebar⌘\\'])
  })

  it('uses Ctrl on Windows and Linux and documents OS conflicts', () => {
    render(<ShortcutsList shortcuts={SHORTCUTS} platform="other" />)
    expect(rows()).toEqual(['SaveCtrl+S', 'Next tabCtrl+Tab or Ctrl+Shift+]', 'Force close tabCtrl+Alt+W', 'Show or hide the sidebarCtrl+\\'])
    expect(screen.getByText(/Ctrl\+Alt shortcuts on Linux desktops/)).toBeInTheDocument()
  })

  it('follows the current OS by default', () => {
    render(<ShortcutsList shortcuts={SHORTCUTS} />)
    expect(rows()).toHaveLength(4)
  })
})
