import { fireEvent, render, screen, within } from '@testing-library/react'
import { beforeEach, describe, expect, it, vi } from 'vitest'
import type { TreeNode } from '@/platform/mappers/tree'

const { useSidebarStore } = await import('@/features/sidebar/store')
const { useTabsStore } = await import('@/features/tabs/store')
const { TabBreadcrumb } = await import('./TabBreadcrumb')

function node(id: string, kind: TreeNode['kind'], name: string, children: TreeNode[] = [], examples: TreeNode['examples'] = []): TreeNode {
  return { id, kind, name, method: null, url: null, error: null, warning: null, children, examples }
}

const tree = [node('c1', 'collection', 'Campaign Manager', [node('r1', 'request', 'Create ad', [], [{ id: 'x1', name: 'Created', code: 201 }])])]

function names() {
  return within(screen.getByRole('navigation', { name: 'Breadcrumb' }))
    .getAllByRole('listitem')
    .map((item) => item.firstElementChild?.textContent)
}

describe('TabBreadcrumb', () => {
  beforeEach(() => {
    useSidebarStore.setState({ tree })
    useTabsStore.setState({
      tabs: [
        { id: 't1', requestId: 'r1', pinned: true },
        { id: 't2', requestId: 'r1', pinned: true, kind: 'example', exampleId: 'x1' },
        { id: 't3', requestId: null, pinned: true },
      ],
    })
  })

  it('shows the saved request of a request tab and opens its collection tab', () => {
    const openItemTab = vi.spyOn(useTabsStore.getState(), 'openItemTab').mockImplementation(() => undefined)
    render(<TabBreadcrumb tabId="t1" />)
    expect(names()).toEqual(['Campaign Manager', 'Create ad'])
    fireEvent.click(screen.getByRole('button', { name: 'Campaign Manager' }))
    expect(openItemTab).toHaveBeenCalledWith({ kind: 'collection', itemId: 'c1' }, 'preview')
  })

  it('shows the example of an example tab under its request', () => {
    render(<TabBreadcrumb tabId="t2" />)
    expect(names()).toEqual(['Campaign Manager', 'Create ad', 'Created'])
  })

  it('shows only the title of an unsaved request', () => {
    render(<TabBreadcrumb tabId="t3" />)
    expect(names()).toEqual(['Untitled Request'])
  })
})
