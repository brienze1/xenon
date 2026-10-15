import { act, fireEvent, render, screen, within } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { beforeEach, describe, expect, it, vi } from 'vitest'
import { emptyDraft } from '@/platform/mappers/request'
import type { TreeNode } from '@/platform/mappers/tree'

vi.mock('@/platform/services/requests', () => ({
  applyUrl: vi.fn(async (draft) => draft),
  applyParams: vi.fn(async (draft) => draft),
  autoHeaders: vi.fn(async () => []),
  getRequest: vi.fn(),
  saveRequest: vi.fn(),
  releaseResponse: vi.fn(async () => undefined),
}))
vi.mock('@/platform/services/ui-state', () => ({ getUIState: vi.fn(), saveUIState: vi.fn(async () => undefined) }))
vi.mock('@/platform/services/workspaces', () => ({ listWorkspaces: vi.fn(), createWorkspace: vi.fn(), activateWorkspace: vi.fn() }))

const { useSidebarStore } = await import('@/features/sidebar/store')
const { useTabsStore } = await import('@/features/tabs/store')
const { useRequestBuilderStore } = await import('@/features/request-builder/store')
const { TabBreadcrumb } = await import('./TabBreadcrumb')

function node(id: string, kind: TreeNode['kind'], name: string, children: TreeNode[] = [], examples: TreeNode['examples'] = []): TreeNode {
  return { id, kind, name, method: null, url: null, error: null, warning: null, children, examples }
}

const tree = [node('c1', 'collection', 'Campaign Manager', [node('r1', 'request', 'Create ad', [], [{ id: 'x1', name: 'Created', code: 201 }])])]

function names() {
  return within(screen.getByRole('navigation', { name: 'Breadcrumb' }))
    .getAllByRole('listitem')
    .map((item) => item.textContent?.replace(/\/$/, ''))
}

function draftName(tabId: string) {
  return useRequestBuilderStore.getState().tabs[tabId]?.draft.name
}

describe('TabBreadcrumb', () => {
  beforeEach(() => {
    useRequestBuilderStore.getState().reset()
    useSidebarStore.setState({ tree })
    useTabsStore.setState({
      tabs: [
        { id: 't1', requestId: 'r1', pinned: true },
        { id: 't2', requestId: 'r1', pinned: true, kind: 'example', exampleId: 'x1' },
        { id: 't3', requestId: null, pinned: true },
      ],
      nameErrors: {},
    })
    useRequestBuilderStore.getState().loadDraft('t1', { ...emptyDraft(), id: 'r1', name: 'Create ad' }, 'rev-1', false)
    useRequestBuilderStore.getState().loadDraft('t2', { ...emptyDraft(), name: 'Created' }, null, false)
    useRequestBuilderStore.getState().loadDraft('t3', emptyDraft(), null, false)
  })

  it('shows the saved request of a request tab and opens its collection tab', () => {
    const openItemTab = vi.spyOn(useTabsStore.getState(), 'openItemTab').mockImplementation(() => undefined)
    render(<TabBreadcrumb tabId="t1" />)
    expect(names()).toEqual(['Campaign Manager', 'Create ad'])
    fireEvent.click(screen.getByRole('button', { name: 'Campaign Manager' }))
    expect(openItemTab).toHaveBeenCalledWith({ kind: 'collection', itemId: 'c1' }, 'preview')
  })

  it('edits the request name in place as a draft, without renaming anything yet', async () => {
    render(<TabBreadcrumb tabId="t1" />)
    await userEvent.click(screen.getByRole('button', { name: /^Rename request/ }))
    const input = screen.getByRole('textbox', { name: 'Rename request' }) as HTMLInputElement
    expect(input).toHaveValue('Create ad')
    expect([input.selectionStart, input.selectionEnd]).toEqual([0, 'Create ad'.length])
    await userEvent.type(input, ' v2')
    expect(draftName('t1')).toBe('Create ad v2')
    expect(useSidebarStore.getState().tree[0]!.children[0]!.name).toBe('Create ad')
    expect(screen.getByRole('navigation', { name: 'Breadcrumb' })).toHaveAttribute('title', 'Campaign Manager / Create ad')
    await userEvent.keyboard('{Escape}')
    expect(draftName('t1')).toBe('Create ad')
  })

  it('opens the name editor from the keyboard', async () => {
    render(<TabBreadcrumb tabId="t1" />)
    screen.getByRole('button', { name: /^Rename request/ }).focus()
    await userEvent.keyboard('{Enter}')
    expect(screen.getByRole('textbox', { name: 'Rename request' })).toHaveFocus()
  })

  it('shows a save error under the name', () => {
    useTabsStore.setState({ nameErrors: { t1: { kind: 'required', subject: 'request' } } })
    render(<TabBreadcrumb tabId="t1" />)
    expect(screen.getByRole('alert')).toHaveTextContent('Type a name for the request.')
    act(() => useTabsStore.getState().editTabName('t1', 'Fixed'))
    expect(screen.queryByRole('alert')).not.toBeInTheDocument()
  })

  it('shows the example of an example tab under its request, and edits its name in place', async () => {
    render(<TabBreadcrumb tabId="t2" />)
    expect(names()).toEqual(['Campaign Manager', 'Create ad', 'Created'])
    screen.getByRole('button', { name: /^Rename example/ }).focus()
    await userEvent.keyboard('{Enter}')
    const input = screen.getByRole('textbox', { name: 'Rename example' }) as HTMLInputElement
    expect([input.selectionStart, input.selectionEnd]).toEqual([0, 'Created'.length])
    await userEvent.clear(input)
    await userEvent.type(input, 'Created OK')
    expect(draftName('t2')).toBe('Created OK')
    expect(useRequestBuilderStore.getState().tabs.t2!.dirty).toBe(false)
  })

  it('shows and edits the name of an unsaved request', async () => {
    render(<TabBreadcrumb tabId="t3" />)
    expect(names()).toEqual(['Untitled Request'])
    await userEvent.click(screen.getByRole('button', { name: /^Rename request/ }))
    await userEvent.clear(screen.getByRole('textbox', { name: 'Rename request' }))
    await userEvent.type(screen.getByRole('textbox', { name: 'Rename request' }), 'Ping')
    expect(draftName('t3')).toBe('Ping')
    expect(useRequestBuilderStore.getState().tabs.t3!.dirty).toBe(true)
  })
})
