import { act, renderHook } from '@testing-library/react'
import type { DragEndEvent, DragMoveEvent, DragStartEvent } from '@dnd-kit/core'
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest'
import type { TreeNode } from '@/platform/mappers/tree'

vi.mock('@/platform/services/tree', () => ({ getTree: vi.fn(), searchTree: vi.fn(), moveItem: vi.fn() }))
vi.mock('@/platform/services/workspaces', () => ({ listWorkspaces: vi.fn(), createWorkspace: vi.fn(), activateWorkspace: vi.fn() }))

const treeService = await import('@/platform/services/tree')
const { useWorkspacesStore } = await import('@/features/workspaces/store')
const { useSidebarStore } = await import('@/features/sidebar/store')
const { AUTO_EXPAND_DELAY_MS, useTreeDrag } = await import('./use-tree-drag')

function node(id: string, kind: TreeNode['kind'], children: TreeNode[] = []): TreeNode {
  return { id, kind, name: id, method: null, url: null, error: null, warning: null, children }
}

const tree = [node('c1', 'collection', [node('f1', 'folder', [node('r1', 'request')]), node('r2', 'request')]), node('c2', 'collection')]

const ROW = { top: 100, height: 28, left: 0, right: 200, bottom: 128, width: 200 }

function moveEvent(active: string, over: string | null, pointerY: number): DragMoveEvent {
  return {
    active: { id: active },
    over: over ? { id: over, rect: ROW } : null,
    delta: { x: 0, y: pointerY - 10 },
    activatorEvent: new MouseEvent('pointerdown', { clientY: 10 }),
  } as unknown as DragMoveEvent
}

function sidebar() {
  return useSidebarStore.getState()
}

describe('useTreeDrag', () => {
  beforeEach(() => {
    vi.useFakeTimers()
    vi.resetAllMocks()
    useSidebarStore.getState().reset()
    useSidebarStore.setState({ tree, status: 'ready' })
    useWorkspacesStore.setState({ activeLocalId: 'L1' })
  })

  afterEach(() => {
    vi.useRealTimers()
  })

  it('shows the drop position under the pointer and moves on drop', async () => {
    vi.mocked(treeService.moveItem).mockResolvedValue([node('c2', 'collection', [node('r2', 'request')])])
    const { result } = renderHook(() => useTreeDrag())
    act(() => result.current.onDragStart({ active: { id: 'r2' } } as unknown as DragStartEvent))
    expect(sidebar().draggingId).toBe('r2')
    expect(document.documentElement).toHaveAttribute('data-dragging')
    act(() => result.current.onDragMove(moveEvent('r2', 'f1', 102)))
    expect(sidebar().dropTarget).toEqual({ id: 'f1', position: 'before' })
    act(() => result.current.onDragMove(moveEvent('r2', 'f1', 114)))
    expect(sidebar().dropTarget).toEqual({ id: 'f1', position: 'inside' })
    act(() => result.current.onDragMove(moveEvent('r2', 'c2', 114)))
    expect(sidebar().dropTarget).toEqual({ id: 'c2', position: 'inside' })
    await act(async () => result.current.onDragEnd(moveEvent('r2', 'c2', 114) as unknown as DragEndEvent))
    expect(treeService.moveItem).toHaveBeenCalledWith('L1', 'r2', 'c2', 0)
    expect(sidebar()).toMatchObject({ draggingId: null, dropTarget: null, selectedId: 'r2', expanded: { c2: true } })
    expect(sidebar().tree[0]?.id).toBe('c2')
    expect(document.documentElement).not.toHaveAttribute('data-dragging')
  })

  it('shows no indicator and moves nothing for refused or empty drops', async () => {
    const { result } = renderHook(() => useTreeDrag())
    act(() => result.current.onDragStart({ active: { id: 'f1' } } as unknown as DragStartEvent))
    act(() => result.current.onDragMove(moveEvent('f1', 'r1', 102)))
    expect(sidebar().dropTarget).toBeNull()
    act(() => result.current.onDragMove(moveEvent('f1', null, 102)))
    expect(sidebar().dropTarget).toBeNull()
    act(() => result.current.onDragMove({ ...moveEvent('f1', 'c2', 114), activatorEvent: new KeyboardEvent('keydown') } as unknown as DragMoveEvent))
    expect(sidebar().dropTarget).toBeNull()
    await act(async () => result.current.onDragEnd(moveEvent('f1', 'r1', 102) as unknown as DragEndEvent))
    await act(async () => result.current.onDragEnd(moveEvent('missing', 'c2', 114) as unknown as DragEndEvent))
    expect(treeService.moveItem).not.toHaveBeenCalled()
  })

  it('opens a closed container held under the pointer for 600 ms', () => {
    const { result } = renderHook(() => useTreeDrag())
    act(() => result.current.onDragStart({ active: { id: 'r2' } } as unknown as DragStartEvent))
    act(() => result.current.onDragMove(moveEvent('r2', 'f1', 114)))
    act(() => vi.advanceTimersByTime(AUTO_EXPAND_DELAY_MS - 1))
    act(() => result.current.onDragMove(moveEvent('r2', 'f1', 115)))
    expect(sidebar().expanded.f1).toBeUndefined()
    act(() => vi.advanceTimersByTime(1))
    expect(sidebar().expanded.f1).toBe(true)
    act(() => result.current.onDragMove(moveEvent('r2', 'f1', 115)))
    act(() => vi.advanceTimersByTime(AUTO_EXPAND_DELAY_MS))
    expect(sidebar().expanded.f1).toBe(true)
  })

  it('stops opening when the pointer leaves, the drag is cancelled or the tree unmounts', () => {
    const { result, unmount } = renderHook(() => useTreeDrag())
    act(() => result.current.onDragStart({ active: { id: 'r2' } } as unknown as DragStartEvent))
    act(() => result.current.onDragMove(moveEvent('r2', 'f1', 114)))
    act(() => result.current.onDragMove(moveEvent('r2', 'f1', 102)))
    act(() => vi.advanceTimersByTime(AUTO_EXPAND_DELAY_MS))
    expect(sidebar().expanded.f1).toBeUndefined()
    act(() => result.current.onDragMove(moveEvent('r2', 'f1', 114)))
    act(() => result.current.onDragCancel())
    act(() => vi.advanceTimersByTime(AUTO_EXPAND_DELAY_MS))
    expect(sidebar()).toMatchObject({ draggingId: null, dropTarget: null })
    expect(sidebar().expanded.f1).toBeUndefined()
    act(() => result.current.onDragStart({ active: { id: 'r2' } } as unknown as DragStartEvent))
    act(() => result.current.onDragMove(moveEvent('r2', 'f1', 114)))
    unmount()
    act(() => vi.advanceTimersByTime(AUTO_EXPAND_DELAY_MS))
    expect(sidebar().expanded.f1).toBeUndefined()
    expect(document.documentElement).not.toHaveAttribute('data-dragging')
  })
})
