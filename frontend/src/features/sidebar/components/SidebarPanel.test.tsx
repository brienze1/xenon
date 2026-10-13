import { act, render, screen, waitFor, within } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { afterAll, beforeAll, beforeEach, describe, expect, it, vi } from 'vitest'
import { AppError } from '@/platform/errors'
import type { TreeNode } from '@/platform/mappers/tree'

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
vi.mock('@/platform/services/examples', () => ({ renameExample: vi.fn(), duplicateExample: vi.fn(), deleteExample: vi.fn() }))
vi.mock('@/platform/services/environments', () => ({
  listEnvironments: vi.fn(async () => ({ environments: [], activeId: null })),
  setActiveEnvironment: vi.fn(async () => null),
}))
vi.mock('@/platform/services/workspaces', () => ({ listWorkspaces: vi.fn(), createWorkspace: vi.fn(), activateWorkspace: vi.fn() }))

const tree = await import('@/platform/services/tree')
const examples = await import('@/platform/services/examples')
const { useWorkspacesStore } = await import('@/features/workspaces/store')
const { SEARCH_DELAY_MS, useSidebarStore } = await import('@/features/sidebar/store')
const { SidebarPanel } = await import('./SidebarPanel')
const { useImportStore } = await import('@/features/import-export/store')

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
  const actions = {
    onOpenRequest: vi.fn(),
    onPinRequest: vi.fn(),
    onOpenExample: vi.fn(),
    onPinExample: vi.fn(),
    onExampleDeleted: vi.fn(),
    onRequestsDeleted: vi.fn(),
    onRenamed: vi.fn(),
    onOpenHistoryEntry: vi.fn(),
    onOpenVariables: vi.fn(),
    selectedVariables: null,
    onOpenItem: vi.fn(),
  }
  render(<SidebarPanel {...actions} />)
  await act(() => useSidebarStore.getState().load())
  return actions
}

function item(name: string) {
  return screen.getByRole('treeitem', { name: new RegExp(name) })
}

const withExamples = [
  node('c1', 'collection', 'Billing', [
    node('r1', 'request', 'Create invoice', [], {
      examples: [
        { id: 'x1', name: 'Created', code: 201 },
        { id: 'x2', name: 'Not found', code: 404 },
      ],
    }),
  ]),
]

describe('SidebarPanel examples', () => {
  beforeEach(() => {
    vi.resetAllMocks()
    useWorkspacesStore.setState({ activeLocalId: 'ws' })
    useSidebarStore.getState().reset()
    useSidebarStore.setState({ expanded: { c1: true } })
    vi.mocked(tree.getTree).mockResolvedValue(withExamples)
  })

  it('nests examples under their request with their status code and opens them', async () => {
    const actions = await renderPanel()
    expect(item('Create invoice')).toHaveAttribute('aria-expanded', 'false')
    expect(screen.queryByRole('treeitem', { name: /Created/ })).toBeNull()
    item('Create invoice').focus()
    await userEvent.keyboard('{ArrowRight}')
    expect(item('Create invoice')).toHaveAttribute('aria-expanded', 'true')
    expect(item('Created')).toHaveTextContent('201')
    expect(item('Not found')).toHaveTextContent('404')
    await userEvent.click(item('Created'))
    expect(actions.onOpenExample).toHaveBeenCalledWith('r1', 'x1')
    await userEvent.dblClick(item('Not found'))
    expect(actions.onPinExample).toHaveBeenCalledWith('r1', 'x2')
    expect(actions.onOpenRequest).not.toHaveBeenCalled()
  })

  it('renames, duplicates and deletes an example from its context menu', async () => {
    useSidebarStore.setState({ expanded: { c1: true, r1: true } })
    vi.mocked(examples.renameExample).mockResolvedValue({} as never)
    vi.mocked(examples.duplicateExample).mockResolvedValue({ id: 'x3' } as never)
    vi.mocked(examples.deleteExample).mockResolvedValue(undefined)
    const actions = await renderPanel()
    await userEvent.pointer({ keys: '[MouseRight]', target: item('Created') })
    await userEvent.click(screen.getByRole('menuitem', { name: 'Rename' }))
    const input = screen.getByRole('textbox', { name: 'Name' })
    await userEvent.clear(input)
    await userEvent.type(input, 'Invoice created{Enter}')
    await waitFor(() => expect(examples.renameExample).toHaveBeenCalledWith('ws', 'r1', 'x1', 'Invoice created'))
    await userEvent.pointer({ keys: '[MouseRight]', target: item('Created') })
    await userEvent.click(screen.getByRole('menuitem', { name: 'Duplicate' }))
    await waitFor(() => expect(examples.duplicateExample).toHaveBeenCalledWith('ws', 'r1', 'x1'))
    await userEvent.pointer({ keys: '[MouseRight]', target: item('Not found') })
    await userEvent.click(screen.getByRole('menuitem', { name: 'Delete' }))
    await userEvent.click(within(screen.getByRole('dialog')).getByRole('button', { name: 'Delete Example' }))
    await waitFor(() => expect(actions.onExampleDeleted).toHaveBeenCalledWith('x2'))
    expect(examples.deleteExample).toHaveBeenCalledWith('ws', 'r1', 'x2')
  })
})

describe('SidebarPanel', () => {
  beforeEach(() => {
    vi.resetAllMocks()
    useWorkspacesStore.setState({ activeLocalId: 'ws' })
    useSidebarStore.getState().reset()
    useImportStore.getState().reset()
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
    expect(screen.getAllByRole('treeitem').map((row) => row.textContent)).toEqual(['Billing', 'Invoices', 'POSTCreate invoice'])
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

  it('opens a collection in a tab from a click and from its context menu', async () => {
    const actions = await renderPanel()
    await userEvent.click(item('Billing'))
    expect(actions.onOpenItem).toHaveBeenLastCalledWith('c1', 'collection', 'preview')
    await userEvent.pointer({ keys: '[MouseRight]', target: item('Invoices') })
    await userEvent.click(screen.getByRole('menuitem', { name: 'Open' }))
    expect(actions.onOpenItem).toHaveBeenLastCalledWith('f1', 'folder', 'pinned')
    expect(screen.queryByRole('dialog')).not.toBeInTheDocument()
  })

  it('opens the Import dialog from the toolbar', async () => {
    await renderPanel()
    await userEvent.click(screen.getByRole('button', { name: 'Import' }))
    expect(useImportStore.getState().step).toBe('choose')
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
})
