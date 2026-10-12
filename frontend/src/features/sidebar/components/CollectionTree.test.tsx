import { act, fireEvent, render, screen, waitFor, within } from '@testing-library/react'
import { afterAll, beforeAll, beforeEach, describe, expect, it, vi } from 'vitest'
import type { TreeNode } from '@/platform/mappers/tree'

vi.mock('@/platform/services/tree', () => ({ getTree: vi.fn(), searchTree: vi.fn(), moveItem: vi.fn() }))
vi.mock('@/platform/services/workspaces', () => ({ listWorkspaces: vi.fn(), createWorkspace: vi.fn(), activateWorkspace: vi.fn() }))
vi.mock('@/platform/services/export', () => ({ getExportOptions: vi.fn(), pickExportTarget: vi.fn(), runExport: vi.fn(), showExportInFolder: vi.fn() }))

const treeService = await import('@/platform/services/tree')
const exportService = await import('@/platform/services/export')
const { useExportStore } = await import('@/features/export/store')
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
  onOpenItem: vi.fn(),
  onDelete: vi.fn(),
  onRenamed: vi.fn(),
}

// dnd-kit swallows clicks for 50 ms after a drag ends; the previous test's drag must not eat this one.
const DND_CLICK_GUARD_MS = 60

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

  it('keeps a click on a row a click', async () => {
    await act(() => new Promise((resolve) => setTimeout(resolve, DND_CLICK_GUARD_MS)))
    render(<CollectionTree {...actions} />)
    fireEvent.click(row('r2'))
    expect(actions.onOpenRequest).toHaveBeenCalledWith('r2')
  })

  it('opens a collection or folder in a preview tab on click, pinned on double click, and expands a closed one', async () => {
    await act(() => new Promise((resolve) => setTimeout(resolve, DND_CLICK_GUARD_MS)))
    render(<CollectionTree {...actions} />)
    fireEvent.click(row('f1'))
    expect(actions.onOpenItem).toHaveBeenLastCalledWith('f1', 'folder', 'preview')
    expect(useSidebarStore.getState().expanded.f1).toBe(true)
    fireEvent.click(row('c1'))
    expect(actions.onOpenItem).toHaveBeenLastCalledWith('c1', 'collection', 'preview')
    expect(useSidebarStore.getState().expanded.c1).toBe(true)
    fireEvent.doubleClick(row('c2'))
    expect(actions.onOpenItem).toHaveBeenLastCalledWith('c2', 'collection', 'pinned')
  })

  it('only expands and collapses from the chevron, without opening a tab', async () => {
    await act(() => new Promise((resolve) => setTimeout(resolve, DND_CLICK_GUARD_MS)))
    render(<CollectionTree {...actions} />)
    const chevron = row('c1').querySelector('[aria-hidden="true"]')!
    fireEvent.click(chevron)
    fireEvent.doubleClick(chevron)
    expect(useSidebarStore.getState().expanded.c1).toBe(false)
    expect(actions.onOpenItem).not.toHaveBeenCalled()
  })

  it('opens a collection or folder pinned from its context menu', () => {
    render(<CollectionTree {...actions} />)
    fireEvent.contextMenu(row('f1'), { clientX: 10, clientY: 10 })
    expect(screen.queryByRole('menuitem', { name: 'Settings…' })).not.toBeInTheDocument()
    fireEvent.click(screen.getByRole('menuitem', { name: 'Open' }))
    expect(actions.onOpenItem).toHaveBeenCalledWith('f1', 'folder', 'pinned')
  })

  it('exports collections, folders and requests from their context menu, but not examples', async () => {
    vi.mocked(exportService.getExportOptions).mockResolvedValue({ suggestedFileName: 'x', formats: [], hasSecrets: false, hasCurrentValues: false })
    useSidebarStore.setState({ tree: [node('c1', 'collection', [node('f1', 'folder'), { ...node('r2', 'request'), examples: [{ id: 'e1', name: 'Name e1', code: 200 }] }])], expanded: { c1: true, r2: true } })
    render(<CollectionTree {...actions} />)
    for (const [id, kind] of [['c1', 'collection'], ['f1', 'folder'], ['r2', 'request']] as const) {
      useExportStore.getState().reset()
      fireEvent.contextMenu(row(id), { clientX: 10, clientY: 10 })
      fireEvent.click(screen.getByRole('menuitem', { name: 'Export…' }))
      await waitFor(() => expect(useExportStore.getState().step).toBe('configure'))
      expect(exportService.getExportOptions).toHaveBeenLastCalledWith({ localId: 'L1', kind, itemId: id, name: `Name ${id}` })
    }
    fireEvent.contextMenu(row('e1'), { clientX: 10, clientY: 10 })
    expect(screen.getByRole('menu', { name: 'Name e1 actions' })).toBeInTheDocument()
    expect(screen.queryByRole('menuitem', { name: 'Export…' })).not.toBeInTheDocument()
    useExportStore.getState().reset()
  })

  it('adds from a + button on collection and folder rows and offers the rest from a … button on every row, without opening the row', async () => {
    await act(() => new Promise((resolve) => setTimeout(resolve, DND_CLICK_GUARD_MS)))
    render(<CollectionTree {...actions} />)
    expect(within(row('r2')).queryByRole('button', { name: 'Add to Name r2' })).not.toBeInTheDocument()
    fireEvent.click(within(row('f1')).getByRole('button', { name: 'Add to Name f1' }))
    const add = screen.getByRole('menu', { name: 'Add to Name f1' })
    expect(within(add).getAllByRole('menuitem').map((item) => item.textContent)).toEqual(['Add Request', 'Add Folder'])
    fireEvent.keyDown(add, { key: 'Escape' })
    fireEvent.click(within(row('f1')).getByRole('button', { name: 'More actions for Name f1' }))
    const more = screen.getByRole('menu', { name: 'More actions for Name f1' })
    expect(within(more).getAllByRole('menuitem').map((item) => item.textContent)).toEqual(['Open', 'Rename', 'Duplicate', 'Export…', 'Delete'])
    fireEvent.keyDown(more, { key: 'Escape' })
    fireEvent.click(within(row('r2')).getByRole('button', { name: 'More actions for Name r2' }))
    expect(within(screen.getByRole('menu', { name: 'More actions for Name r2' })).getAllByRole('menuitem').map((item) => item.textContent)).toEqual([
      'Rename',
      'Duplicate',
      'Export…',
      'Delete',
    ])
    expect(actions.onOpenItem).not.toHaveBeenCalled()
    expect(actions.onOpenRequest).not.toHaveBeenCalled()
    expect(useSidebarStore.getState().expanded.f1).toBeFalsy()
  })
})
