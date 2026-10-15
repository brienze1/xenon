import { act, render, screen } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest'
import { AppError } from '@/platform/errors'
import { emptyDraft } from '@/platform/mappers/request'
import type { TreeNode } from '@/platform/mappers/tree'

vi.mock('@/platform/services/requests', () => ({
  applyUrl: vi.fn(),
  applyParams: vi.fn(),
  parseBulk: vi.fn(),
  formatBulk: vi.fn(),
  autoHeaders: vi.fn(),
  sendRequest: vi.fn(),
  getRequest: vi.fn(),
  saveRequest: vi.fn(),
}))
vi.mock('@/platform/services/tree', () => ({ getTree: vi.fn(), renameItem: vi.fn() }))
vi.mock('@/platform/services/environments', () => ({
  listEnvironments: vi.fn(async () => ({ environments: [], activeId: null })),
  setActiveEnvironment: vi.fn(async () => null),
}))
vi.mock('@/platform/services/workspaces', () => ({ listWorkspaces: vi.fn(), createWorkspace: vi.fn(), activateWorkspace: vi.fn() }))

const requests = await import('@/platform/services/requests')
const tree = await import('@/platform/services/tree')
const { useWorkspacesStore } = await import('@/features/workspaces/store')
const { useSidebarStore } = await import('@/features/sidebar/store')
const { useRequestBuilderStore } = await import('@/features/request-builder/store')
const { useTabsStore } = await import('@/features/tabs/store')
const { InfoPanel } = await import('./InfoPanel')
await import('@/shared/ui/markdown/MarkdownView')

function node(id: string, kind: TreeNode['kind'], name: string, children: TreeNode[] = []): TreeNode {
  return { id, kind, name, method: null, url: null, error: null, warning: null, children }
}

const sample = [node('c1', 'collection', 'Billing', [node('f1', 'folder', 'Invoices', [node('r1', 'request', 'Create invoice')])])]

function resetStores() {
  useRequestBuilderStore.getState().reset()
  useSidebarStore.getState().reset()
}

function tabId() {
  return 'tab-1'
}

function loadDraft(id: string, name: string, description = '') {
  useRequestBuilderStore.setState((state) => ({
    tabs: { ...state.tabs, [tabId()]: { ...state.tabs[tabId()]!, draft: { ...emptyDraft(), id, name, description }, baseRevision: id ? 'rev' : null } },
  }))
}

describe('InfoPanel', () => {
  beforeEach(() => {
    resetStores()
    vi.resetAllMocks()
    vi.mocked(requests.autoHeaders).mockResolvedValue([])
    vi.mocked(tree.getTree).mockResolvedValue(sample)
    useWorkspacesStore.setState({ activeLocalId: 'ws' })
    useRequestBuilderStore.getState().ensureTab('tab-1')
    useSidebarStore.setState({ tree: sample })
    useTabsStore.setState({ tabs: [{ id: tabId(), requestId: null, pinned: true }], activeTabId: tabId(), nameErrors: {} })
  })

  afterEach(() => resetStores())

  it('shows an unsaved request', () => {
    render(<InfoPanel tabId={tabId()} />)
    expect(screen.queryByRole('textbox', { name: /name/i })).not.toBeInTheDocument()
    expect(screen.getByRole('button', { name: /^Rename request/ })).toHaveTextContent('Untitled Request')
    expect(screen.getByText('Not saved yet')).toBeInTheDocument()
    expect(screen.getByText('No description')).toBeInTheDocument()
  })

  it('shows where a saved request lives and renders its description as markdown', async () => {
    loadDraft('r1', 'Create invoice', 'Creates a **draft** invoice')
    render(<InfoPanel tabId={tabId()} />)
    expect(screen.getByRole('navigation', { name: 'Location' })).toHaveTextContent('Billing / Invoices')
    expect(await screen.findByText('draft')).toContainHTML('strong')
  })

  it('follows the location when the tree changes', () => {
    loadDraft('r1', 'Create invoice')
    render(<InfoPanel tabId={tabId()} />)
    act(() => useSidebarStore.setState({ tree: [node('c1', 'collection', 'Billing', [node('f1', 'folder', 'Receipts', [node('r1', 'request', 'Create invoice')])])] }))
    expect(screen.getByRole('navigation', { name: 'Location' })).toHaveTextContent('Billing / Receipts')
    act(() => useSidebarStore.setState({ tree: [] }))
    expect(screen.getByText('Unknown')).toBeInTheDocument()
  })

  it('edits the description into the draft', async () => {
    render(<InfoPanel tabId={tabId()} />)
    await userEvent.click(screen.getByRole('button', { name: 'Edit Description' }))
    await userEvent.type(screen.getByRole('textbox', { name: 'Description' }), 'Hello')
    expect(useRequestBuilderStore.getState().tabs[tabId()]!.draft.description).toBe('Hello')
    await userEvent.click(screen.getByRole('button', { name: 'Done Editing' }))
    expect(await screen.findByText('Hello')).toBeInTheDocument()
  })

  it('focuses the description editor from its visible label', async () => {
    render(<InfoPanel tabId={tabId()} />)
    await userEvent.click(screen.getByRole('button', { name: 'Edit Description' }))
    await userEvent.click(screen.getByText('Description', { selector: 'label' }))
    expect(screen.getByRole('textbox', { name: 'Description' })).toHaveFocus()
  })

  it('renames an unsaved request in its draft', async () => {
    render(<InfoPanel tabId={tabId()} />)
    await userEvent.click(screen.getByRole('button', { name: /^Rename request/ }))
    const input = screen.getByRole('textbox', { name: 'Rename request' })
    await userEvent.clear(input)
    await userEvent.type(input, 'Ping{Enter}')
    expect(useRequestBuilderStore.getState().tabs[tabId()]!.draft.name).toBe('Ping')
    expect(screen.getByRole('button', { name: /^Rename request/ })).toHaveTextContent('Ping')
    expect(tree.renameItem).not.toHaveBeenCalled()
  })

  it('edits the name of a saved request as a draft and never renames it on blur', async () => {
    loadDraft('r1', 'Create invoice')
    render(<InfoPanel tabId={tabId()} />)
    await userEvent.click(screen.getByRole('button', { name: /^Rename request/ }))
    const input = screen.getByRole('textbox', { name: 'Rename request' })
    await userEvent.clear(input)
    await userEvent.type(input, 'Create draft')
    await userEvent.tab()
    expect(screen.queryByRole('textbox', { name: 'Rename request' })).not.toBeInTheDocument()
    expect(screen.getByRole('button', { name: /^Rename request/ })).toHaveTextContent('Create draft')
    expect(useRequestBuilderStore.getState().tabs[tabId()]!.draft.name).toBe('Create draft')
    expect(useRequestBuilderStore.getState().tabs[tabId()]!.dirty).toBe(false)
    expect(tree.renameItem).not.toHaveBeenCalled()
  })

  it('shows the same draft name the breadcrumb edits', () => {
    loadDraft('r1', 'Create invoice')
    render(<InfoPanel tabId={tabId()} />)
    act(() => useTabsStore.getState().editTabName(tabId(), 'From the breadcrumb'))
    expect(screen.getByRole('button', { name: /^Rename request/ })).toHaveTextContent('From the breadcrumb')
  })

  it('shows a failed rename under the name until the name is edited again', async () => {
    loadDraft('r1', 'Taken')
    useTabsStore.setState({ nameErrors: { [tabId()]: { kind: 'failed', error: new AppError('TRE-03409', 'A request with that name already exists') } } })
    render(<InfoPanel tabId={tabId()} />)
    const alert = screen.getByRole('alert')
    expect(alert).toHaveTextContent('A request with that name already exists. Choose another name, then save.')
    const name = screen.getByRole('button', { name: /^Rename request/ })
    expect(name).toHaveAttribute('aria-invalid', 'true')
    expect(name).toHaveAccessibleDescription(alert.textContent ?? '')
    await userEvent.click(name)
    await userEvent.type(screen.getByRole('textbox', { name: 'Rename request' }), '!')
    expect(screen.queryByRole('alert')).not.toBeInTheDocument()
  })
})
