import { act, fireEvent, render, screen, waitFor, within } from '@testing-library/react'
import { afterAll, beforeAll, beforeEach, describe, expect, it, vi } from 'vitest'
import type { TreeNode } from '@/platform/mappers/tree'

vi.mock('@/platform/services/tree', () => ({ getTree: vi.fn(), searchTree: vi.fn(), moveItem: vi.fn() }))
vi.mock('@/platform/services/workspaces', () => ({ listWorkspaces: vi.fn(), createWorkspace: vi.fn(), activateWorkspace: vi.fn() }))

const treeService = await import('@/platform/services/tree')
const { useWorkspacesStore } = await import('@/features/workspaces/store')
const { useSidebarStore } = await import('@/features/sidebar/store')
const { CollectionTree, TREE_ROW_HEIGHT } = await import('./CollectionTree')

function node(id: string, kind: TreeNode['kind'], children: TreeNode[] = []): TreeNode {
  return { id, kind, name: `Name ${id}`, method: kind === 'request' ? 'GET' : null, url: null, error: null, warning: null, children }
}

const tree = [node('c1', 'collection', [node('f1', 'folder', [node('r1', 'request')]), node('r2', 'request')]), node('c2', 'collection')]

const actions = {
  onOpenRequest: vi.fn(),
  onPinRequest: vi.fn(),
  onOpenExample: vi.fn(),
  onPinExample: vi.fn(),
  onOpenSettings: vi.fn(),
  onDelete: vi.fn(),
  onRenamed: vi.fn(),
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

function rows() {
  const items = within(screen.getByRole('tree', { name: 'Collections' })).getAllByRole('treeitem')
  items.forEach((item, index) => {
    vi.spyOn(item, 'getBoundingClientRect').mockImplementation(() =>
      DOMRect.fromRect({ x: 0, y: index * TREE_ROW_HEIGHT, width: 240, height: TREE_ROW_HEIGHT }),
    )
  })
  return items
}

function row(id: string) {
  return rows().find((item) => item.dataset.treeId === id)!
}

async function dragTo(from: HTMLElement, y: number) {
  fireEvent.pointerDown(from, { isPrimary: true, button: 0, clientX: 20, clientY: Number(from.getBoundingClientRect().y) + 5 })
  fireEvent.pointerMove(document, { isPrimary: true, clientX: 20, clientY: Number(from.getBoundingClientRect().y) + 15 })
  await act(() => new Promise((resolve) => setTimeout(resolve)))
  fireEvent.pointerMove(document, { isPrimary: true, clientX: 20, clientY: y })
}

describe('CollectionTree drag and drop', () => {
  beforeEach(() => {
    vi.resetAllMocks()
    useSidebarStore.getState().reset()
    useSidebarStore.setState({ tree, status: 'ready', expanded: { c1: true } })
    useWorkspacesStore.setState({ activeLocalId: 'L1' })
  })

  it('shows where the item drops and moves it there', async () => {
    vi.mocked(treeService.moveItem).mockResolvedValue(tree)
    render(<CollectionTree {...actions} />)
    const items = rows()
    expect(items.map((item) => item.dataset.treeId)).toEqual(['c1', 'f1', 'r2', 'c2'])
    await dragTo(row('r2'), 3 * TREE_ROW_HEIGHT + 14)
    await waitFor(() => expect(row('c2')).toHaveAttribute('data-drop', 'inside'))
    expect(row('r2').className).toContain('dragging')
    fireEvent.pointerMove(document, { isPrimary: true, clientX: 20, clientY: TREE_ROW_HEIGHT + 2 })
    await waitFor(() => expect(row('f1')).toHaveAttribute('data-drop', 'before'))
    expect(row('c2')).not.toHaveAttribute('data-drop')
    fireEvent.pointerUp(document, { isPrimary: true, clientX: 20, clientY: TREE_ROW_HEIGHT + 2 })
    await waitFor(() => expect(treeService.moveItem).toHaveBeenCalledWith('L1', 'r2', 'c1', 0))
    expect(row('f1')).not.toHaveAttribute('data-drop')
  })

  it('cancels with Escape without moving anything', async () => {
    render(<CollectionTree {...actions} />)
    await dragTo(row('r2'), 3 * TREE_ROW_HEIGHT + 14)
    await waitFor(() => expect(row('c2')).toHaveAttribute('data-drop', 'inside'))
    fireEvent.keyDown(document, { code: 'Escape', key: 'Escape' })
    await waitFor(() => expect(row('c2')).not.toHaveAttribute('data-drop'))
    fireEvent.pointerUp(document, { isPrimary: true, clientX: 20, clientY: 3 * TREE_ROW_HEIGHT + 14 })
    expect(treeService.moveItem).not.toHaveBeenCalled()
  })

  it('refuses to drop a folder into itself', async () => {
    render(<CollectionTree {...actions} />)
    await dragTo(row('f1'), TREE_ROW_HEIGHT + 14)
    await act(() => new Promise((resolve) => setTimeout(resolve)))
    expect(row('f1')).not.toHaveAttribute('data-drop')
    fireEvent.pointerUp(document, { isPrimary: true, clientX: 20, clientY: TREE_ROW_HEIGHT + 14 })
    await act(() => new Promise((resolve) => setTimeout(resolve)))
    expect(treeService.moveItem).not.toHaveBeenCalled()
  })

  it('keeps a click on a row a click', () => {
    render(<CollectionTree {...actions} />)
    fireEvent.click(row('r2'))
    expect(actions.onOpenRequest).toHaveBeenCalledWith('r2')
  })
})
