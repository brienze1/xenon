import { fireEvent, render, screen, within } from '@testing-library/react'
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

const actions = { onOpenItem: vi.fn(), onOpenRequest: vi.fn() }

const EDITOR = (
  <button type="button" aria-label="Rename request">
    Draft name
  </button>
)

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

  it('shows the path from the collection down to the request, with the name editor last', () => {
    render(<ItemBreadcrumb itemId="r1" name="Create ad" nameEditor={EDITOR} {...actions} />)
    const nav = screen.getByRole('navigation', { name: 'Breadcrumb' })
    expect(within(nav).getByRole('list').tagName).toBe('OL')
    expect(crumbs().map((item) => item.textContent)).toEqual(['Campaign Manager/', 'Ads/', 'CRUD/', 'Draft name'])
    expect(nav).toHaveAttribute('title', 'Campaign Manager / Ads / CRUD / Create ad')
    expect(crumbs().at(-1)).toHaveAttribute('aria-current', 'page')
    expect(screen.getByRole('button', { name: 'Ads' })).not.toHaveAttribute('aria-current')
  })

  it('with an alias shown, ends with the alias name editor after the request name as plain text', () => {
    const aliasEditor = (
      <button type="button" aria-label="Rename alias">
        mx-test-user
      </button>
    )
    render(<ItemBreadcrumb itemId="r1" name="Create ad · mx-test-user" nameEditor={null} aliasEditor={aliasEditor} {...actions} />)
    expect(crumbs().map((item) => item.textContent)).toEqual(['Campaign Manager/', 'Ads/', 'CRUD/', 'Create ad/', 'mx-test-user'])
    expect(crumbs().at(-1)).toHaveAttribute('aria-current', 'page')
    expect(screen.getByRole('button', { name: 'Rename alias' })).toBeInTheDocument()
    expect(screen.queryByRole('button', { name: 'Create ad' })).not.toBeInTheDocument()
  })

  it('opens a collection or folder like the sidebar does', () => {
    render(<ItemBreadcrumb itemId="r1" name="Create ad" nameEditor={EDITOR} {...actions} />)
    fireEvent.click(screen.getByRole('button', { name: 'Campaign Manager' }))
    expect(actions.onOpenItem).toHaveBeenCalledWith('c1', 'collection', 'preview')
    fireEvent.click(screen.getByRole('button', { name: 'CRUD' }))
    expect(actions.onOpenItem).toHaveBeenCalledWith('f2', 'folder', 'preview')
  })

  it('never renames by itself: the name editor it is given does', () => {
    render(<ItemBreadcrumb itemId="r1" name="Create ad" nameEditor={EDITOR} {...actions} />)
    fireEvent.click(screen.getByRole('button', { name: /^Rename request/ }))
    expect(screen.queryByRole('textbox')).not.toBeInTheDocument()
    expect(treeService.renameItem).not.toHaveBeenCalled()
    expect(examplesService.renameExample).not.toHaveBeenCalled()
    expect(useSidebarStore.getState().renamingId).toBeNull()
  })

  it('shows an example under its request', () => {
    render(<ItemBreadcrumb itemId="x1" name="Created" nameEditor={EDITOR} {...actions} />)
    expect(crumbs().map((item) => item.textContent)).toEqual(['Campaign Manager/', 'Ads/', 'CRUD/', 'Create ad/', 'Draft name'])
    fireEvent.click(screen.getByRole('button', { name: 'Create ad' }))
    expect(actions.onOpenRequest).toHaveBeenCalledWith('r1')
  })

  it('shows only the name editor of an unsaved request', () => {
    render(<ItemBreadcrumb itemId={null} name="New Request" nameEditor={EDITOR} {...actions} />)
    expect(crumbs().map((item) => item.textContent)).toEqual(['Draft name'])
    expect(crumbs()[0]).toHaveAttribute('aria-current', 'page')
    expect(screen.getByRole('navigation', { name: 'Breadcrumb' })).toHaveAttribute('title', 'New Request')
  })
})
