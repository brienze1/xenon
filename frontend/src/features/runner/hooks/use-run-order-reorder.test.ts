import type { Active, Over } from '@dnd-kit/core'
import { act, renderHook } from '@testing-library/react'
import { describe, expect, it, vi } from 'vitest'
import { planRow } from '@/features/runner/test-fixtures'
import { announcementsFor, useRunOrderReorder } from './use-run-order-reorder'

const rows = [planRow('r1', 1, { name: 'Login' }), planRow('r2', 2, { name: 'Get order' }), planRow('r3', 3, { name: 'Delete order' })]

function item(id: string): Active & Over {
  return { id } as Active & Over
}

describe('run order announcements', () => {
  it('names the request and its position, and skips the drag-over of the pick-up itself', () => {
    const announcements = announcementsFor(rows, { overId: null })
    expect(announcements.onDragStart({ active: item('r3') })).toBe('Picked up Delete order.')
    expect(announcements.onDragOver({ active: item('r3'), over: item('r3') })).toBeUndefined()
    expect(announcements.onDragOver({ active: item('r3'), over: item('r1') })).toBe('Delete order moved to position 1 of 3.')
    expect(announcements.onDragOver({ active: item('r3'), over: null })).toBe('Delete order is no longer over a position.')
    expect(announcements.onDragEnd({ active: item('r3'), over: item('r2') })).toBe('Delete order dropped at position 2 of 3.')
    expect(announcements.onDragCancel({ active: item('r3'), over: null })).toBe('Reordering cancelled. Delete order is back at position 3 of 3.')
  })
})

describe('useRunOrderReorder', () => {
  it('moves the row on a drop over another row, and ignores a drop in place', () => {
    const onMove = vi.fn()
    const { result } = renderHook(() => useRunOrderReorder(rows, onMove))
    act(() => {
      result.current.onDragStart()
      result.current.onDragEnd({ active: item('r3'), over: item('r1') } as never)
    })
    expect(onMove).toHaveBeenCalledWith('r3', 'r1')
    act(() => result.current.onDragEnd({ active: item('r1'), over: item('r1') } as never))
    act(() => result.current.onDragEnd({ active: item('r1'), over: null } as never))
    expect(onMove).toHaveBeenCalledTimes(1)
    expect(document.documentElement.hasAttribute('data-dragging')).toBe(false)
  })
})
