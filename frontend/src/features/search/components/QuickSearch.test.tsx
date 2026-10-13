import { act, render, screen, within } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest'

vi.mock('@/platform/services/tree', () => ({ quickSearch: vi.fn(), getTree: vi.fn(async () => []), searchTree: vi.fn() }))
vi.mock('@/platform/services/workspaces', () => ({ listWorkspaces: vi.fn(), createWorkspace: vi.fn(), activateWorkspace: vi.fn() }))

const tree = await import('@/platform/services/tree')
const { useWorkspacesStore } = await import('@/features/workspaces/store')
const { useTabsStore } = await import('@/features/tabs/store')
const { useSearchStore } = await import('@/features/search/store')
const { QuickSearch } = await import('./QuickSearch')

const hits = [
  { id: 'r1', name: 'Create invoice', method: 'POST', url: 'https://x/invoices', path: 'Billing / Invoices' },
  { id: 'r2', name: 'Invoice totals', method: 'GET', url: 'https://x/totals', path: 'Billing' },
]

describe('QuickSearch', () => {
  beforeEach(() => {
    vi.resetAllMocks()
    useSearchStore.getState().reset()
    useWorkspacesStore.setState({ activeLocalId: 'ws' })
  })

  afterEach(() => useSearchStore.getState().reset())

  it('stays closed until opened', () => {
    render(<QuickSearch />)
    expect(screen.queryByRole('dialog')).not.toBeInTheDocument()
  })

  it('lists the matches as you type and opens the chosen one with the keyboard', async () => {
    vi.mocked(tree.quickSearch).mockResolvedValue(hits)
    const openPinned = vi.spyOn(useTabsStore.getState(), 'openPinned').mockResolvedValue()
    render(<QuickSearch />)
    act(() => useSearchStore.getState().openSearch())
    const input = screen.getByRole('combobox', { name: 'Search requests' })
    expect(input).toHaveFocus()
    expect(screen.getByRole('status')).toHaveTextContent('Type to search the requests of this workspace.')
    await userEvent.type(input, 'inv')
    const options = await within(screen.getByRole('listbox', { name: 'Matching requests' })).findAllByRole('option')
    expect(options.map((option) => option.textContent)).toEqual(['POSTCreate invoiceBilling / Invoiceshttps://x/invoices', 'GETInvoice totalsBillinghttps://x/totals'])
    expect(options[0]).toHaveAttribute('aria-selected', 'true')
    expect(input).toHaveAttribute('aria-activedescendant', options[0]!.id)
    expect(screen.getByRole('status')).toHaveTextContent('2 requests found.')
    await userEvent.keyboard('{ArrowDown}')
    expect(options[1]).toHaveAttribute('aria-selected', 'true')
    await userEvent.keyboard('{Enter}')
    expect(openPinned).toHaveBeenCalledWith('r2')
    expect(screen.queryByRole('dialog')).not.toBeInTheDocument()
    openPinned.mockRestore()
  })

  it('opens a result on click, says when nothing matches and closes with Escape', async () => {
    vi.mocked(tree.quickSearch).mockResolvedValueOnce([])
    const openPinned = vi.spyOn(useTabsStore.getState(), 'openPinned').mockResolvedValue()
    render(<QuickSearch />)
    act(() => useSearchStore.getState().openSearch())
    await userEvent.type(screen.getByRole('combobox'), 'zz')
    expect(await screen.findByText('No requests match “zz”.')).toBeInTheDocument()
    vi.mocked(tree.quickSearch).mockResolvedValue(hits)
    await userEvent.type(screen.getByRole('combobox'), 'x')
    await userEvent.click(await screen.findByRole('option', { name: /Invoice totals/ }))
    expect(openPinned).toHaveBeenCalledWith('r2')
    act(() => useSearchStore.getState().openSearch())
    await userEvent.keyboard('{Escape}')
    expect(screen.queryByRole('dialog')).not.toBeInTheDocument()
    openPinned.mockRestore()
  })

  it('shows a failed search', async () => {
    vi.mocked(tree.quickSearch).mockRejectedValue({ cause: { error: { code: 'TRE-40400', description: 'The search could not run' } } })
    render(<QuickSearch />)
    act(() => useSearchStore.getState().openSearch())
    await userEvent.type(screen.getByRole('combobox'), 'a')
    expect(await screen.findByRole('alert')).toHaveTextContent('The search could not run')
  })
})
