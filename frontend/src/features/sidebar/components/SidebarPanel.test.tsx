import { act, render, screen, waitFor, within } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { afterAll, beforeAll, beforeEach, describe, expect, it, vi } from 'vitest'
import { AppError } from '@/platform/errors'
import type { ItemSettings, TreeNode } from '@/platform/mappers/tree'

vi.mock('@/platform/services/tree', () => ({
  getTree: vi.fn(),
  searchTree: vi.fn(),
  createCollection: vi.fn(),
  createFolder: vi.fn(),
  createRequest: vi.fn(),
  renameItem: vi.fn(),
  duplicateItem: vi.fn(),
  deleteItem: vi.fn(),
  getItemSettings: vi.fn(),
  saveItemSettings: vi.fn(),
}))
vi.mock('@/platform/services/requests', () => ({ parseBulk: vi.fn(), formatBulk: vi.fn() }))
vi.mock('@/platform/services/workspaces', () => ({ listWorkspaces: vi.fn(), createWorkspace: vi.fn(), activateWorkspace: vi.fn() }))

const tree = await import('@/platform/services/tree')
const requests = await import('@/platform/services/requests')
const { useWorkspacesStore } = await import('@/features/workspaces/store')
const { SEARCH_DELAY_MS, useSidebarStore } = await import('@/features/sidebar/store')
const { SidebarPanel } = await import('./SidebarPanel')

function node(id: string, kind: TreeNode['kind'], name: string, children: TreeNode[] = [], extra: Partial<TreeNode> = {}): TreeNode {
  return { id, kind, name, method: kind === 'request' ? 'POST' : null, url: null, error: null, warning: null, children, ...extra }
}

const sample = [
  node('c1', 'collection', 'Billing', [
    node('f1', 'folder', 'Invoices', [node('r1', 'request', 'Create invoice')]),
    node('r2', 'request', 'Ping', [], { method: 'GET', warning: 'same id' }),
    node('error:x', 'error', 'broken.yaml', [], { error: 'broken.yaml cannot be read' }),
  ]),
]

const settings: ItemSettings = {
  id: 'c1',
  kind: 'collection',
  name: 'Billing',
  description: '',
  auth: { type: 'inherit', bearer: { token: '' }, basic: { username: '', password: '' } },
  headers: [],
}

const offsetHeight = Object.getOwnPropertyDescriptor(HTMLElement.prototype, 'offsetHeight')

beforeAll(() => {
  Object.defineProperty(HTMLElement.prototype, 'offsetHeight', { configurable: true, get: () => 600 })
})

afterAll(() => {
  if (offsetHeight) {
    Object.defineProperty(HTMLElement.prototype, 'offsetHeight', offsetHeight)
  }
})

async function renderPanel() {
  const actions = { onOpenRequest: vi.fn(), onPinRequest: vi.fn(), onRequestsDeleted: vi.fn(), onRenamed: vi.fn(), onOpenHistoryEntry: vi.fn() }
  render(<SidebarPanel {...actions} />)
  await act(() => useSidebarStore.getState().load())
  return actions
}

function item(name: string) {
  return screen.getByRole('treeitem', { name: new RegExp(name) })
}

describe('SidebarPanel', () => {
  beforeEach(() => {
    vi.resetAllMocks()
    useWorkspacesStore.setState({ activeLocalId: 'ws' })
    useSidebarStore.getState().reset()
    vi.mocked(tree.getTree).mockResolvedValue(sample)
  })

  it('shows an empty state that creates a collection', async () => {
    vi.mocked(tree.getTree).mockResolvedValue([])
    vi.mocked(tree.createCollection).mockResolvedValue(node('c9', 'collection', 'New Collection'))
    await renderPanel()
    await userEvent.click(screen.getByRole('button', { name: 'Create Collection' }))
    expect(tree.createCollection).toHaveBeenCalledWith('ws', 'New Collection')
  })

  it('expands containers and opens requests', async () => {
    const actions = await renderPanel()
    expect(screen.getAllByRole('treeitem')).toHaveLength(1)
    expect(item('Billing')).toHaveAttribute('tabindex', '0')
    await userEvent.click(item('Billing'))
    expect(item('Billing')).toHaveAttribute('aria-expanded', 'true')
    expect(item('Ping')).toHaveTextContent('GET')
    expect(item('Ping')).toHaveTextContent('Duplicate ID')
    expect(item('broken.yaml')).toHaveAttribute('title', 'broken.yaml cannot be read')
    await userEvent.click(item('Invoices'))
    await userEvent.click(item('Create invoice'))
    expect(actions.onOpenRequest).toHaveBeenCalledWith('r1')
    expect(item('Create invoice')).toHaveAttribute('aria-selected', 'true')
    await userEvent.keyboard('{Enter}')
    expect(actions.onOpenRequest).toHaveBeenCalledTimes(2)
    await userEvent.dblClick(item('Create invoice'))
    expect(actions.onPinRequest).toHaveBeenCalledWith('r1')
  })

  it('moves through rows and expands or collapses them with the keyboard', async () => {
    await renderPanel()
    item('Billing').focus()
    await userEvent.keyboard('{ArrowRight}')
    expect(item('Billing')).toHaveAttribute('aria-expanded', 'true')
    expect(item('Billing')).toHaveFocus()
    await userEvent.keyboard('{ArrowRight}')
    expect(item('Invoices')).toHaveFocus()
    expect(item('Invoices')).toHaveAttribute('tabindex', '0')
    expect(item('Billing')).toHaveAttribute('tabindex', '-1')
    await userEvent.keyboard('{ArrowDown}')
    expect(item('Ping')).toHaveFocus()
    await userEvent.keyboard('{End}')
    expect(item('broken.yaml')).toHaveFocus()
    await userEvent.keyboard('{ArrowDown}')
    expect(item('broken.yaml')).toHaveFocus()
    await userEvent.keyboard('{ArrowUp}')
    expect(item('Ping')).toHaveFocus()
    await userEvent.keyboard('{ArrowLeft}')
    expect(item('Billing')).toHaveFocus()
    await userEvent.keyboard('{End}{Home}')
    expect(item('Billing')).toHaveFocus()
    await userEvent.keyboard('{ArrowLeft}')
    expect(item('Billing')).toHaveAttribute('aria-expanded', 'false')
    expect(screen.getAllByRole('treeitem')).toHaveLength(1)
  })

  it('filters the tree through the debounced search', async () => {
    vi.mocked(tree.searchTree).mockResolvedValue({ matchIds: ['r1'], ancestorIds: ['c1', 'f1'] })
    await renderPanel()
    vi.useFakeTimers({ shouldAdvanceTime: true })
    await userEvent.type(screen.getByRole('searchbox', { name: 'Search requests' }), 'inv')
    await act(() => vi.advanceTimersByTimeAsync(SEARCH_DELAY_MS))
    vi.useRealTimers()
    expect(tree.searchTree).toHaveBeenCalledWith('ws', 'inv')
    expect(screen.getAllByRole('treeitem').map((row) => row.textContent)).toEqual(['▾Billing', '▾Invoices', 'POSTCreate invoice'])
    vi.mocked(tree.searchTree).mockResolvedValue({ matchIds: [], ancestorIds: [] })
    await userEvent.type(screen.getByRole('searchbox', { name: 'Search requests' }), 'zzz')
    await waitFor(() => expect(screen.getByText('No requests match your search.')).toBeInTheDocument())
  })

  it('renames inline from the context menu', async () => {
    vi.mocked(tree.renameItem).mockResolvedValue(node('c1', 'collection', 'Payments'))
    const actions = await renderPanel()
    await userEvent.pointer({ keys: '[MouseRight]', target: item('Billing') })
    await userEvent.click(screen.getByRole('menuitem', { name: 'Rename' }))
    const input = screen.getByRole('textbox', { name: 'Name' })
    await userEvent.clear(input)
    await userEvent.type(input, 'Payments{Enter}')
    await waitFor(() => expect(actions.onRenamed).toHaveBeenCalledWith('c1'))
    expect(tree.renameItem).toHaveBeenCalledWith('ws', 'c1', 'Payments')
  })

  it('cancels an inline rename with Escape', async () => {
    await renderPanel()
    act(() => useSidebarStore.getState().startRename('c1'))
    await userEvent.type(screen.getByRole('textbox', { name: 'Name' }), 'x{Escape}')
    expect(screen.queryByRole('textbox', { name: 'Name' })).not.toBeInTheDocument()
    expect(tree.renameItem).not.toHaveBeenCalled()
  })

  it('adds folders and requests and duplicates from the context menu', async () => {
    vi.mocked(tree.createFolder).mockResolvedValue(node('f2', 'folder', 'New Folder'))
    vi.mocked(tree.createRequest).mockResolvedValue(node('r9', 'request', 'New Request'))
    vi.mocked(tree.duplicateItem).mockResolvedValue(node('c2', 'collection', 'Billing Copy'))
    await renderPanel()
    await userEvent.pointer({ keys: '[MouseRight]', target: item('Billing') })
    await userEvent.click(screen.getByRole('menuitem', { name: 'Add Folder' }))
    expect(tree.createFolder).toHaveBeenCalledWith('ws', 'c1', 'New Folder')
    await userEvent.pointer({ keys: '[MouseRight]', target: item('Billing') })
    await userEvent.click(screen.getByRole('menuitem', { name: 'Add Request' }))
    expect(tree.createRequest).toHaveBeenCalledWith('ws', 'c1', 'New Request')
    await userEvent.pointer({ keys: '[MouseRight]', target: item('Billing') })
    await userEvent.click(screen.getByRole('menuitem', { name: 'Duplicate' }))
    expect(tree.duplicateItem).toHaveBeenCalledWith('ws', 'c1')
  })

  it('deletes only after confirmation and reports the deleted requests', async () => {
    vi.mocked(tree.deleteItem).mockResolvedValue(['r1', 'r2'])
    const actions = await renderPanel()
    await userEvent.pointer({ keys: '[MouseRight]', target: item('Billing') })
    await userEvent.click(screen.getByRole('menuitem', { name: 'Delete' }))
    const dialog = screen.getByRole('dialog', { name: 'Delete “Billing”?' })
    expect(dialog).toHaveTextContent('everything inside it')
    await userEvent.click(within(dialog).getByRole('button', { name: 'Cancel' }))
    expect(tree.deleteItem).not.toHaveBeenCalled()
    await userEvent.pointer({ keys: '[MouseRight]', target: item('Billing') })
    await userEvent.click(screen.getByRole('menuitem', { name: 'Delete' }))
    await userEvent.click(within(screen.getByRole('dialog')).getByRole('button', { name: 'Delete Collection' }))
    await waitFor(() => expect(actions.onRequestsDeleted).toHaveBeenCalledWith(['r1', 'r2']))
    expect(tree.deleteItem).toHaveBeenCalledWith('ws', 'c1')
  })

  it('edits collection auth and headers in the settings dialog', async () => {
    vi.mocked(tree.getItemSettings).mockResolvedValue(settings)
    vi.mocked(tree.saveItemSettings).mockResolvedValue(settings)
    vi.mocked(requests.formatBulk).mockResolvedValue('X-Tenant:acme')
    vi.mocked(requests.parseBulk).mockResolvedValue([{ id: 'h1', key: 'X-Tenant', value: 'acme', description: '', enabled: true }])
    await renderPanel()
    await userEvent.pointer({ keys: '[MouseRight]', target: item('Billing') })
    await userEvent.click(screen.getByRole('menuitem', { name: 'Settings…' }))
    const dialog = await screen.findByRole('dialog', { name: 'Billing Settings' })
    expect(dialog).toHaveTextContent('A collection has no parent')
    await userEvent.selectOptions(within(dialog).getByRole('combobox', { name: 'Auth type' }), 'bearer')
    await userEvent.type(within(dialog).getByRole('textbox', { name: 'Token' }), 'secret')
    await userEvent.click(within(dialog).getByRole('button', { name: 'Bulk Edit' }))
    await userEvent.click(within(dialog).getByRole('button', { name: 'Save Settings' }))
    await waitFor(() => expect(screen.queryByRole('dialog')).not.toBeInTheDocument())
    expect(tree.saveItemSettings).toHaveBeenCalledWith('ws', {
      ...settings,
      auth: { ...settings.auth, type: 'bearer', bearer: { token: 'secret' } },
      headers: [{ id: 'h1', key: 'X-Tenant', value: 'acme', description: '', enabled: true }],
    })
  })

  it('shows the settings save in progress', async () => {
    let finish: () => void = () => undefined
    vi.mocked(tree.getItemSettings).mockResolvedValue(settings)
    vi.mocked(tree.saveItemSettings).mockImplementation(() => new Promise((resolve) => (finish = () => resolve(settings))))
    await renderPanel()
    await userEvent.pointer({ keys: '[MouseRight]', target: item('Billing') })
    await userEvent.click(screen.getByRole('menuitem', { name: 'Settings…' }))
    const dialog = await screen.findByRole('dialog', { name: 'Billing Settings' })
    await userEvent.click(within(dialog).getByRole('button', { name: 'Save Settings' }))
    expect(within(dialog).getByRole('button', { name: 'Saving…' })).toBeDisabled()
    await act(async () => finish())
    await waitFor(() => expect(screen.queryByRole('dialog')).not.toBeInTheDocument())
  })

  it('shows and dismisses errors', async () => {
    vi.mocked(tree.duplicateItem).mockRejectedValue(Object.assign(new Error('x'), { cause: { error: { code: 'TRE-02404', description: 'The item does not exist' } } }))
    await renderPanel()
    await userEvent.pointer({ keys: '[MouseRight]', target: item('Billing') })
    await userEvent.click(screen.getByRole('menuitem', { name: 'Duplicate' }))
    expect(await screen.findByRole('alert')).toHaveTextContent('The item does not exist')
    await userEvent.click(screen.getByRole('button', { name: 'Dismiss Error' }))
    expect(screen.queryByRole('alert')).not.toBeInTheDocument()
  })

  it('shows a rename failure under the name input and keeps the rejected text', async () => {
    vi.mocked(tree.renameItem).mockRejectedValue(new AppError('TRE-03409', 'An item with that name already exists'))
    const actions = await renderPanel()
    act(() => useSidebarStore.getState().startRename('c1'))
    const input = screen.getByRole('textbox', { name: 'Name' })
    await userEvent.clear(input)
    await userEvent.type(input, 'Taken{Enter}')
    const alert = await screen.findByRole('alert')
    expect(alert).toHaveTextContent('An item with that name already exists. Choose another name')
    expect(input).toHaveValue('Taken')
    expect(input).toHaveAttribute('aria-invalid', 'true')
    expect(input).toHaveAccessibleDescription(alert.textContent ?? '')
    expect(useSidebarStore.getState().error).toBeNull()
    expect(actions.onRenamed).not.toHaveBeenCalled()
    await userEvent.type(input, '2')
    expect(screen.queryByRole('alert')).not.toBeInTheDocument()
  })

  it('keeps the delete dialog open and shows why deleting failed', async () => {
    vi.mocked(tree.deleteItem).mockRejectedValue(new AppError('TRE-02500', 'The folder is locked'))
    const actions = await renderPanel()
    await userEvent.pointer({ keys: '[MouseRight]', target: item('Billing') })
    await userEvent.click(screen.getByRole('menuitem', { name: 'Delete' }))
    const dialog = screen.getByRole('dialog', { name: 'Delete “Billing”?' })
    await userEvent.click(within(dialog).getByRole('button', { name: 'Delete Collection' }))
    expect(await within(dialog).findByRole('alert')).toHaveTextContent('The folder is locked. Check that the file')
    expect(useSidebarStore.getState().error).toBeNull()
    expect(actions.onRequestsDeleted).not.toHaveBeenCalled()
  })

  it('shows a failed settings load with a way to try again', async () => {
    vi.mocked(tree.getItemSettings).mockRejectedValueOnce(new AppError('TRE-02404', 'The item does not exist'))
    vi.mocked(tree.getItemSettings).mockResolvedValueOnce(settings)
    await renderPanel()
    await userEvent.pointer({ keys: '[MouseRight]', target: item('Billing') })
    await userEvent.click(screen.getByRole('menuitem', { name: 'Settings…' }))
    const failed = await screen.findByRole('dialog', { name: 'Settings' })
    expect(within(failed).getByRole('alert')).toHaveTextContent('The item does not exist. Check that the workspace folder exists')
    await userEvent.click(within(failed).getByRole('button', { name: 'Try Again' }))
    expect(await screen.findByRole('dialog', { name: 'Billing Settings' })).toBeInTheDocument()
    expect(tree.getItemSettings).toHaveBeenCalledTimes(2)
  })

  it('shows settings save and bulk failures inside the dialog', async () => {
    vi.mocked(tree.getItemSettings).mockResolvedValue(settings)
    vi.mocked(tree.saveItemSettings).mockRejectedValue(new AppError('TRE-05500', 'The folder is read-only'))
    vi.mocked(requests.formatBulk).mockResolvedValue('X-Tenant:acme')
    vi.mocked(requests.parseBulk).mockRejectedValue(new AppError('REQ-01400', 'Line 1 has no colon'))
    await renderPanel()
    await userEvent.pointer({ keys: '[MouseRight]', target: item('Billing') })
    await userEvent.click(screen.getByRole('menuitem', { name: 'Settings…' }))
    const dialog = await screen.findByRole('dialog', { name: 'Billing Settings' })
    await userEvent.click(within(dialog).getByRole('button', { name: 'Save Settings' }))
    expect(await within(dialog).findByRole('alert')).toHaveTextContent('The folder is read-only. Check that the workspace folder exists and is writable')
    await userEvent.click(within(dialog).getByRole('button', { name: 'Bulk Edit' }))
    await userEvent.click(within(dialog).getByRole('button', { name: 'Key-Value Edit' }))
    expect(await within(dialog).findByRole('alert')).toHaveTextContent('Line 1 has no colon. Check that each line reads key:value')
    expect(useSidebarStore.getState().error).toBeNull()
  })
})
