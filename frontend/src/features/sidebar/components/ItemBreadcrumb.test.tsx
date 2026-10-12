import { fireEvent, render, screen, waitFor, within } from '@testing-library/react'
import { beforeEach, describe, expect, it, vi } from 'vitest'
import type { TreeNode } from '@/platform/mappers/tree'

vi.mock('@/platform/services/tree', () => ({ getTree: vi.fn(), renameItem: vi.fn() }))
vi.mock('@/platform/services/examples', () => ({ renameExample: vi.fn(), duplicateExample: vi.fn(), deleteExample: vi.fn() }))
vi.mock('@/platform/services/workspaces', () => ({ listWorkspaces: vi.fn(), createWorkspace: vi.fn(), activateWorkspace: vi.fn() }))

const treeService = await import('@/platform/services/tree')
const examplesService = await import('@/platform/services/examples')
const { useWorkspacesStore } = await import('@/features/workspaces/store')
const { useSidebarStore } = await import('@/features/sidebar/store')
const { ItemBreadcrumb } = await import('./ItemBreadcrumb')

function node(id: string, kind: TreeNode['kind'], name: string, children: TreeNode[] = [], examples: TreeNode['examples'] = []): TreeNode {
  return { id, kind, name, method: kind === 'request' ? 'POST' : null, url: null, error: null, warning: null, children, examples }
}

function sampleTree(requestName = 'Create ad'): TreeNode[] {
  return [
    node('c1', 'collection', 'Campaign Manager', [
      node('f1', 'folder', 'Ads', [node('f2', 'folder', 'CRUD', [node('r1', 'request', requestName, [], [{ id: 'x1', name: 'Created', code: 201 }])])]),
    ]),
  ]
}

const actions = { onOpenItem: vi.fn(), onOpenRequest: vi.fn(), onRenamed: vi.fn() }

function crumbs() {
  return within(screen.getByRole('navigation', { name: 'Breadcrumb' })).getAllByRole('listitem')
}

describe('ItemBreadcrumb', () => {
  beforeEach(() => {
    vi.resetAllMocks()
    useSidebarStore.getState().reset()
    useSidebarStore.setState({ tree: sampleTree(), status: 'ready' })
    useWorkspacesStore.setState({ activeLocalId: 'L1' })
  })

  it('shows the path from the collection down to the request', () => {
    render(<ItemBreadcrumb itemId="r1" name="Create ad" {...actions} />)
    const nav = screen.getByRole('navigation', { name: 'Breadcrumb' })
    expect(within(nav).getByRole('list').tagName).toBe('OL')
    expect(crumbs().map((item) => item.firstElementChild?.textContent)).toEqual(['Campaign Manager', 'Ads', 'CRUD', 'Create ad'])
    expect(nav).toHaveAttribute('title', 'Campaign Manager / Ads / CRUD / Create ad')
    expect(screen.getByRole('button', { name: 'Create ad' })).toHaveAttribute('aria-current', 'page')
    expect(screen.getByRole('button', { name: 'Ads' })).not.toHaveAttribute('aria-current')
  })

  it('opens a collection or folder like the sidebar does', () => {
    render(<ItemBreadcrumb itemId="r1" name="Create ad" {...actions} />)
    fireEvent.click(screen.getByRole('button', { name: 'Campaign Manager' }))
    expect(actions.onOpenItem).toHaveBeenCalledWith('c1', 'collection', 'preview')
    fireEvent.click(screen.getByRole('button', { name: 'CRUD' }))
    expect(actions.onOpenItem).toHaveBeenCalledWith('f2', 'folder', 'preview')
  })

  it('renames the request through the sidebar and tells the app', async () => {
    vi.mocked(treeService.renameItem).mockResolvedValue(undefined as never)
    vi.mocked(treeService.getTree).mockResolvedValue(sampleTree('Create campaign ad'))
    render(<ItemBreadcrumb itemId="r1" name="Create ad" {...actions} />)
    fireEvent.click(screen.getByRole('button', { name: 'Create ad' }))
    const input = screen.getByRole('textbox', { name: 'Name' })
    expect(input).toHaveValue('Create ad')
    fireEvent.change(input, { target: { value: 'Create campaign ad' } })
    fireEvent.keyDown(input, { key: 'Enter' })
    await waitFor(() => expect(actions.onRenamed).toHaveBeenCalledWith('r1'))
    expect(treeService.renameItem).toHaveBeenCalledWith('L1', 'r1', 'Create campaign ad')
    expect(screen.queryByRole('textbox', { name: 'Name' })).not.toBeInTheDocument()
    expect(screen.getByRole('button', { name: 'Create campaign ad' })).toHaveAttribute('aria-current', 'page')
  })

  it('keeps the old name when the edit is cancelled or left empty', async () => {
    render(<ItemBreadcrumb itemId="r1" name="Create ad" {...actions} />)
    fireEvent.click(screen.getByRole('button', { name: 'Create ad' }))
    fireEvent.change(screen.getByRole('textbox', { name: 'Name' }), { target: { value: 'Other' } })
    fireEvent.keyDown(screen.getByRole('textbox', { name: 'Name' }), { key: 'Escape' })
    expect(screen.getByRole('button', { name: 'Create ad' })).toBeInTheDocument()

    fireEvent.click(screen.getByRole('button', { name: 'Create ad' }))
    const input = screen.getByRole('textbox', { name: 'Name' })
    fireEvent.change(input, { target: { value: '   ' } })
    fireEvent.blur(input)
    await waitFor(() => expect(screen.queryByRole('textbox', { name: 'Name' })).not.toBeInTheDocument())
    expect(screen.getByRole('button', { name: 'Create ad' })).toBeInTheDocument()
    expect(treeService.renameItem).not.toHaveBeenCalled()
    expect(actions.onRenamed).not.toHaveBeenCalled()
  })

  it('leaves the sidebar row out of rename mode', () => {
    render(<ItemBreadcrumb itemId="r1" name="Create ad" {...actions} />)
    fireEvent.click(screen.getByRole('button', { name: 'Create ad' }))
    expect(useSidebarStore.getState().renamingId).toBeNull()
  })

  it('shows an example under its request and renames the example', async () => {
    vi.mocked(examplesService.renameExample).mockResolvedValue({} as never)
    vi.mocked(treeService.getTree).mockResolvedValue(sampleTree())
    render(<ItemBreadcrumb itemId="x1" name="Created" {...actions} />)
    expect(crumbs().map((item) => item.firstElementChild?.textContent)).toEqual(['Campaign Manager', 'Ads', 'CRUD', 'Create ad', 'Created'])
    fireEvent.click(screen.getByRole('button', { name: 'Create ad' }))
    expect(actions.onOpenRequest).toHaveBeenCalledWith('r1')
    fireEvent.click(screen.getByRole('button', { name: 'Created' }))
    const input = screen.getByRole('textbox', { name: 'Name' })
    fireEvent.change(input, { target: { value: 'Created OK' } })
    fireEvent.keyDown(input, { key: 'Enter' })
    await waitFor(() => expect(actions.onRenamed).toHaveBeenCalledWith('x1'))
    expect(examplesService.renameExample).toHaveBeenCalledWith('L1', 'r1', 'x1', 'Created OK')
  })

  it('shows only the name of an unsaved request, which the tree cannot rename', () => {
    render(<ItemBreadcrumb itemId={null} name="New Request" {...actions} />)
    expect(crumbs().map((item) => item.firstElementChild?.textContent)).toEqual(['New Request'])
    expect(screen.queryByRole('button')).not.toBeInTheDocument()
    expect(screen.getByText('New Request')).toHaveAttribute('aria-current', 'page')
  })
})
