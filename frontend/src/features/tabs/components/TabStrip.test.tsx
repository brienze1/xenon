import { act, fireEvent, render, screen, waitFor, within } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest'
import { AppError } from '@/platform/errors'
import { emptyDraft } from '@/platform/mappers/request'
import { defaultUIState } from '@/platform/mappers/ui-state'

vi.mock('@/platform/services/requests', () => ({
  applyUrl: vi.fn(async (draft) => draft),
  applyParams: vi.fn(async (draft) => draft),
  autoHeaders: vi.fn(async () => []),
  getRequest: vi.fn(),
  saveRequest: vi.fn(),
  saveRequestAs: vi.fn(),
  releaseResponse: vi.fn(async () => undefined),
  lastResult: vi.fn(async () => null),
}))
vi.mock('@/platform/services/ui-state', () => ({ getUIState: vi.fn(), saveUIState: vi.fn(async () => undefined) }))
vi.mock('@/platform/services/environments', () => ({
  listEnvironments: vi.fn(async () => ({ environments: [], activeId: null })),
  setActiveEnvironment: vi.fn(async () => null),
}))
vi.mock('@/platform/services/workspaces', () => ({ listWorkspaces: vi.fn(), createWorkspace: vi.fn(), activateWorkspace: vi.fn() }))
vi.mock('@/platform/services/dialogs', () => ({ pickFolder: vi.fn(), pickFile: vi.fn(), pickSaveFile: vi.fn() }))
vi.mock('@/platform/services/tree', () => ({ getTree: vi.fn(async () => []), searchTree: vi.fn() }))

const requests = await import('@/platform/services/requests')
const { useWorkspacesStore } = await import('@/features/workspaces/store')
const { useRequestBuilderStore } = await import('@/features/request-builder/store')
const { useSidebarStore } = await import('@/features/sidebar/store')
const { useTabsStore } = await import('@/features/tabs/store')
const { useResponseViewerStore } = await import('@/features/response-viewer/store')
const { TabStrip } = await import('./TabStrip')
const { CloseTabPrompt } = await import('./CloseTabPrompt')
const { SaveRequestDialog } = await import('./SaveRequestDialog')
const { TabNotice } = await import('./TabNotice')

function resetWorkspaceStores() {
  useTabsStore.getState().reset()
  useRequestBuilderStore.getState().reset()
  useResponseViewerStore.getState().reset()
  useSidebarStore.getState().reset()
}

function strip() {
  return screen.getByRole('tablist', { name: 'Open requests' })
}

function renderStrip() {
  render(
    <>
      <TabStrip />
      <CloseTabPrompt />
      <SaveRequestDialog />
    </>,
  )
}

describe('TabStrip', () => {
  beforeEach(() => {
    resetWorkspaceStores()
    vi.clearAllMocks()
    vi.mocked(requests.getRequest).mockImplementation(async (_localId, id) => ({ draft: { ...emptyDraft(), id, name: `Request ${id}`, method: 'POST' }, revision: 'rev' }))
    useWorkspacesStore.setState({ activeLocalId: 'ws' })
    useTabsStore.getState().hydrate('ws', defaultUIState())
  })

  afterEach(() => resetWorkspaceStores())

  it('shows tabs with method, italic preview title and dirty dot, and switches on click', async () => {
    await act(() => useTabsStore.getState().openPinned('A'))
    await act(() => useTabsStore.getState().openPreview('B'))
    renderStrip()
    const tabs = within(strip()).getAllByRole('tab')
    expect(tabs).toHaveLength(3)
    expect(tabs[1]).toHaveTextContent('POST')
    expect(tabs[1]).toHaveTextContent('Request A')
    expect(tabs[2]).toHaveAttribute('title', 'Request B (preview)')
    expect(tabs[2]).toHaveAttribute('aria-selected', 'true')
    await userEvent.click(tabs[1]!)
    expect(tabs[1]).toHaveAttribute('aria-selected', 'true')
    act(() => useRequestBuilderStore.getState().setUrl(useTabsStore.getState().activeTabId!, 'https://x'))
    expect(within(tabs[1]!).getByRole('img', { name: 'Unsaved changes' })).toBeInTheDocument()
  })

  it('pins a preview tab on double click', async () => {
    await act(() => useTabsStore.getState().openPreview('B'))
    renderStrip()
    await userEvent.dblClick(within(strip()).getAllByRole('tab')[1]!)
    expect(useTabsStore.getState().previewTabId).toBeNull()
    expect(within(strip()).getAllByRole('tab')[1]).toHaveAttribute('title', 'Request B')
  })

  it('opens a new tab with the + button', async () => {
    renderStrip()
    await userEvent.click(screen.getByRole('button', { name: 'New Tab' }))
    expect(within(strip()).getAllByRole('tab')).toHaveLength(2)
  })

  it('closes clean tabs at once and asks before closing a dirty one', async () => {
    await act(() => useTabsStore.getState().openPinned('A'))
    renderStrip()
    await userEvent.click(screen.getByRole('button', { name: 'Close Request A' }))
    expect(within(strip()).getAllByRole('tab')).toHaveLength(1)
    const tabId = useTabsStore.getState().activeTabId!
    act(() => useRequestBuilderStore.getState().setUrl(tabId, 'https://x'))
    await userEvent.click(screen.getByRole('button', { name: 'Close Untitled Request' }))
    const prompt = screen.getByRole('dialog', { name: 'Save Changes?' })
    await userEvent.click(within(prompt).getByRole('button', { name: 'Cancel' }))
    expect(within(strip()).getAllByRole('tab')).toHaveLength(1)
    fireEvent.keyDown(within(strip()).getByRole('tab'), { key: 'Delete' })
    await userEvent.click(within(screen.getByRole('dialog')).getByRole('button', { name: 'Don’t Save' }))
    expect(within(strip()).queryAllByRole('tab')).toHaveLength(0)
  })

  it('is keyboard accessible and closes with the middle button', async () => {
    await act(() => useTabsStore.getState().openPinned('A'))
    renderStrip()
    const tabs = within(strip()).getAllByRole('tab')
    expect(tabs[1]).toHaveAttribute('tabindex', '0')
    expect(tabs[0]).toHaveAttribute('tabindex', '-1')
    fireEvent.keyDown(tabs[0]!, { key: 'Enter' })
    expect(useTabsStore.getState().activeTabId).toBe(useTabsStore.getState().tabs[0]!.id)
    fireEvent.keyDown(tabs[0]!, { key: ' ' })
    fireEvent(tabs[1]!, new MouseEvent('auxclick', { bubbles: true, button: 1 }))
    expect(within(strip()).getAllByRole('tab')).toHaveLength(1)
  })

  it('moves focus across tabs with the arrow, Home and End keys', async () => {
    await act(() => useTabsStore.getState().openPinned('A'))
    await act(() => useTabsStore.getState().openPinned('B'))
    renderStrip()
    const tabs = within(strip()).getAllByRole('tab')
    tabs[2]!.focus()
    await userEvent.keyboard('{ArrowRight}')
    expect(tabs[0]).toHaveFocus()
    await userEvent.keyboard('{ArrowLeft}')
    expect(tabs[2]).toHaveFocus()
    await userEvent.keyboard('{Home}')
    expect(tabs[0]).toHaveFocus()
    await userEvent.keyboard('{End}')
    expect(tabs[2]).toHaveFocus()
    await userEvent.keyboard('{ArrowLeft}')
    expect(tabs[1]).toHaveFocus()
  })

  it('reorders tabs from the keyboard: Space picks up, arrows move, Space drops, keeping the tab focusable (never inert) during a keyboard drag', async () => {
    await act(() => useTabsStore.getState().openPinned('A'))
    renderStrip()
    const tabs = within(strip()).getAllByRole('tab')
    tabs.forEach((tab, index) => {
      vi.spyOn(tab, 'getBoundingClientRect').mockImplementation(() => DOMRect.fromRect({ x: index * 100, y: 0, width: 100, height: 30 }))
    })
    const [first, second] = useTabsStore.getState().tabs.map((tab) => tab.id)
    tabs[0]!.focus()
    await userEvent.keyboard('[Space]')
    expect(useTabsStore.getState().activeTabId).toBe(first)
    await act(() => new Promise((resolve) => setTimeout(resolve)))
    await userEvent.keyboard('[ArrowRight]')
    expect(tabs[0]).toHaveFocus()
    expect(tabs[0]).not.toHaveAttribute('inert')
    expect(document.documentElement).toHaveAttribute('data-dragging')
    await userEvent.keyboard('[Space]')
    expect(useTabsStore.getState().tabs.map((tab) => tab.id)).toEqual([second, first])
    expect(document.documentElement).not.toHaveAttribute('data-dragging')
  })

  it('makes the dragged tab inert and turns off text selection during a pointer drag', async () => {
    await act(() => useTabsStore.getState().openPinned('A'))
    renderStrip()
    const tabs = within(strip()).getAllByRole('tab')
    tabs.forEach((tab, index) => {
      vi.spyOn(tab, 'getBoundingClientRect').mockImplementation(() => DOMRect.fromRect({ x: index * 100, y: 0, width: 100, height: 30 }))
    })
    fireEvent.pointerDown(tabs[0]!, { isPrimary: true, button: 0, clientX: 10, clientY: 10 })
    fireEvent.pointerMove(document, { isPrimary: true, clientX: 60, clientY: 10 })
    await waitFor(() => expect(tabs[0]).toHaveAttribute('inert'))
    expect(tabs[1]).not.toHaveAttribute('inert')
    expect(document.documentElement).toHaveAttribute('data-dragging')
    fireEvent.pointerUp(document, { isPrimary: true, clientX: 60, clientY: 10 })
    await waitFor(() => expect(tabs[0]).not.toHaveAttribute('inert'))
    expect(document.documentElement).not.toHaveAttribute('data-dragging')
  })

  it('keeps roving arrow navigation and Enter activation when no drag is in progress', async () => {
    await act(() => useTabsStore.getState().openPinned('A'))
    renderStrip()
    const tabs = within(strip()).getAllByRole('tab')
    tabs[1]!.focus()
    await userEvent.keyboard('{ArrowLeft}{Enter}')
    expect(tabs[0]).toHaveFocus()
    expect(useTabsStore.getState().activeTabId).toBe(useTabsStore.getState().tabs[0]!.id)
  })

  it('saves a new tab into a chosen collection', async () => {
    useSidebarStore.setState({
      tree: [
        {
          id: 'c1',
          kind: 'collection',
          name: 'Billing',
          method: null,
          url: null,
          error: null,
          warning: null,
          children: [{ id: 'f1', kind: 'folder', name: 'Invoices', method: null, url: null, error: null, warning: null, children: [] }],
        },
      ],
    })
    vi.mocked(requests.saveRequestAs).mockResolvedValue({ draft: { ...emptyDraft(), id: 'N', name: 'Users' }, revision: 'rev-N' })
    renderStrip()
    await act(() => useTabsStore.getState().save())
    const dialog = screen.getByRole('dialog', { name: 'Save Request' })
    const name = within(dialog).getByRole('textbox', { name: 'Request name' })
    await userEvent.clear(name)
    await userEvent.type(name, 'Users')
    await userEvent.click(within(dialog).getByRole('radio', { name: 'Invoices' }))
    await userEvent.click(within(dialog).getByRole('button', { name: 'Save Request' }))
    expect(requests.saveRequestAs).toHaveBeenCalledWith('ws', 'f1', 'Users', expect.anything())
    expect(screen.queryByRole('dialog')).not.toBeInTheDocument()
    expect(useTabsStore.getState().tabs[0]?.requestId).toBe('N')
  })

  it('shows the save in progress and focuses the chosen place when it is missing', async () => {
    useSidebarStore.setState({
      tree: [{ id: 'c1', kind: 'collection', name: 'Billing', method: null, url: null, error: null, warning: null, children: [] }],
    })
    let finish: () => void = () => undefined
    vi.mocked(requests.saveRequestAs).mockImplementation(
      () => new Promise((resolve) => (finish = () => resolve({ draft: { ...emptyDraft(), id: 'N', name: 'Users' }, revision: 'rev-N' }))),
    )
    renderStrip()
    await act(() => useTabsStore.getState().save())
    const dialog = screen.getByRole('dialog', { name: 'Save Request' })
    expect(within(dialog).getByRole('radio', { name: 'Billing' })).toBeChecked()
    await userEvent.click(within(dialog).getByRole('button', { name: 'Save Request' }))
    expect(within(dialog).getByRole('button', { name: 'Saving…' })).toBeDisabled()
    await act(async () => finish())
    expect(screen.queryByRole('dialog')).not.toBeInTheDocument()
  })

  it('shows a failed save inside the dialog and keeps it open', async () => {
    useSidebarStore.setState({
      tree: [{ id: 'c1', kind: 'collection', name: 'Billing', method: null, url: null, error: null, warning: null, children: [] }],
    })
    vi.mocked(requests.saveRequestAs).mockRejectedValue(new AppError('TRE-01404', 'The collection no longer exists'))
    renderStrip()
    await act(() => useTabsStore.getState().save())
    const dialog = screen.getByRole('dialog', { name: 'Save Request' })
    await userEvent.click(within(dialog).getByRole('button', { name: 'Save Request' }))
    const alert = await within(dialog).findByRole('alert')
    expect(alert).toHaveTextContent('The collection no longer exists. Choose another name or place, then try again.')
    expect(alert.parentElement).toHaveFocus()
    expect(useRequestBuilderStore.getState().tabs[useTabsStore.getState().activeTabId!]!.error).toBeNull()
    expect(useTabsStore.getState().saveDialog).not.toBeNull()
  })

  it('keeps the close prompt open and shows why saving failed', async () => {
    await act(() => useTabsStore.getState().openPinned('A'))
    const tabId = useTabsStore.getState().activeTabId!
    act(() => useRequestBuilderStore.getState().setUrl(tabId, 'https://changed'))
    vi.mocked(requests.saveRequest).mockRejectedValue(new AppError('REQ-13500', 'The disk is full'))
    renderStrip()
    act(() => useTabsStore.getState().requestClose(tabId))
    const dialog = screen.getByRole('dialog', { name: 'Save Changes?' })
    await userEvent.click(within(dialog).getByRole('button', { name: 'Save Changes' }))
    expect(await within(dialog).findByRole('alert')).toHaveTextContent('The disk is full. Try again, or choose Don’t Save')
    expect(useTabsStore.getState().tabs.some((tab) => tab.id === tabId)).toBe(true)
    await userEvent.click(within(dialog).getByRole('button', { name: 'Don’t Save' }))
    expect(screen.queryByRole('dialog')).not.toBeInTheDocument()
    expect(useTabsStore.getState().tabs.some((tab) => tab.id === tabId)).toBe(false)
  })

  it('suggests a copy name for save as and can be cancelled', async () => {
    await act(() => useTabsStore.getState().openPinned('A'))
    renderStrip()
    act(() => useTabsStore.getState().saveAs())
    const dialog = screen.getByRole('dialog', { name: 'Save Request As' })
    expect(within(dialog).getByRole('textbox', { name: 'Request name' })).toHaveValue('Request A Copy')
    expect(within(dialog).getByText('Create a collection in the sidebar first.')).toBeInTheDocument()
    expect(within(dialog).getByRole('button', { name: 'Save Request' })).toBeEnabled()
    await userEvent.click(within(dialog).getByRole('button', { name: 'Save Request' }))
    expect(within(dialog).getByRole('alert')).toHaveTextContent('Choose where to save the request.')
    expect(within(dialog).getByRole('group', { name: 'Save to' })).toHaveFocus()
    expect(within(dialog).getByRole('group', { name: 'Save to' })).toHaveAccessibleDescription('Choose where to save the request.')
    expect(requests.saveRequestAs).not.toHaveBeenCalled()
    const name = within(dialog).getByRole('textbox', { name: 'Request name' })
    await userEvent.clear(name)
    await userEvent.click(within(dialog).getByRole('button', { name: 'Save Request' }))
    expect(name).toHaveFocus()
    expect(name).toHaveAttribute('aria-invalid', 'true')
    expect(name).toHaveAccessibleDescription('Enter a name for the request.')
    await userEvent.click(within(dialog).getByRole('button', { name: 'Cancel' }))
    expect(useTabsStore.getState().saveDialog).toBeNull()
  })
})

describe('TabNotice', () => {
  beforeEach(() => {
    resetWorkspaceStores()
    vi.clearAllMocks()
    vi.mocked(requests.getRequest).mockImplementation(async (_localId, id) => ({ draft: { ...emptyDraft(), id, name: 'Disk' }, revision: 'rev-disk' }))
    useWorkspacesStore.setState({ activeLocalId: 'ws' })
    useTabsStore.getState().hydrate('ws', defaultUIState())
  })

  afterEach(() => resetWorkspaceStores())

  it('offers keep mine or load from disk for a changed file', async () => {
    await act(() => useTabsStore.getState().openPinned('A'))
    const tabId = useTabsStore.getState().activeTabId!
    act(() => useRequestBuilderStore.getState().setUrl(tabId, 'https://mine'))
    await act(() => useTabsStore.getState().handleChangedOnDisk({ localId: 'ws', requestId: 'A', deleted: false, revision: 'new' }))
    render(<TabNotice tabId={tabId} />)
    expect(screen.getByRole('status')).toHaveTextContent('changed on disk')
    vi.mocked(requests.saveRequest).mockResolvedValue({ draft: { ...emptyDraft(), id: 'A' }, revision: 'rev-2' })
    await userEvent.click(screen.getByRole('button', { name: 'Keep Your Version' }))
    expect(requests.saveRequest).toHaveBeenCalledWith('ws', expect.anything(), 'rev-disk', true)
    expect(screen.queryByRole('status')).not.toBeInTheDocument()
  })

  it('offers load from disk', async () => {
    await act(() => useTabsStore.getState().openPinned('A'))
    const tabId = useTabsStore.getState().activeTabId!
    act(() => useRequestBuilderStore.getState().setUrl(tabId, 'https://mine'))
    await act(() => useTabsStore.getState().handleChangedOnDisk({ localId: 'ws', requestId: 'A', deleted: false, revision: 'new' }))
    render(<TabNotice tabId={tabId} />)
    await userEvent.click(screen.getByRole('button', { name: 'Load from Disk' }))
    expect(useRequestBuilderStore.getState().tabs[tabId]!.dirty).toBe(false)
  })

  it('offers keeping a deleted request as an unsaved tab or closing it', async () => {
    await act(() => useTabsStore.getState().openPinned('A'))
    const tabId = useTabsStore.getState().activeTabId!
    act(() => useRequestBuilderStore.getState().setUrl(tabId, 'https://mine'))
    act(() => useTabsStore.getState().handleDeleted(['A']))
    const { unmount } = render(<TabNotice tabId={tabId} />)
    expect(screen.getByRole('status')).toHaveTextContent('no longer exists on disk')
    await userEvent.click(screen.getByRole('button', { name: 'Keep as Unsaved Tab' }))
    expect(useTabsStore.getState().tabs.find((tab) => tab.id === tabId)?.requestId).toBeNull()
    unmount()
    await act(() => useTabsStore.getState().openPinned('B'))
    const other = useTabsStore.getState().activeTabId!
    act(() => useRequestBuilderStore.getState().setUrl(other, 'https://x'))
    act(() => useTabsStore.getState().handleDeleted(['B']))
    render(
      <>
        <TabNotice tabId={other} />
        <CloseTabPrompt />
      </>,
    )
    await userEvent.click(screen.getByRole('button', { name: 'Close Tab' }))
    const prompt = screen.getByRole('dialog', { name: 'Save Changes?' })
    expect(useTabsStore.getState().tabs.some((tab) => tab.id === other)).toBe(true)
    await userEvent.click(within(prompt).getByRole('button', { name: 'Don’t Save' }))
    expect(useTabsStore.getState().tabs.some((tab) => tab.id === other)).toBe(false)
  })

})

describe('TabContextMenu', () => {
  beforeEach(() => {
    resetWorkspaceStores()
    vi.clearAllMocks()
    vi.mocked(requests.getRequest).mockImplementation(async (_localId, id) => ({ draft: { ...emptyDraft(), id, name: `Request ${id}`, method: 'POST' }, revision: 'rev' }))
    useWorkspacesStore.setState({ activeLocalId: 'ws' })
    useTabsStore.getState().hydrate('ws', defaultUIState())
  })

  afterEach(() => resetWorkspaceStores())

  it('offers the Postman tab menu on right click and runs each item', async () => {
    await act(() => useTabsStore.getState().openPinned('A'))
    await act(() => useTabsStore.getState().openPinned('B'))
    renderStrip()
    const [, first] = within(strip()).getAllByRole('tab')
    fireEvent.contextMenu(first!, { clientX: 10, clientY: 10 })
    const menu = screen.getByRole('menu', { name: 'Request A tab actions' })
    expect(within(menu).getAllByRole('menuitem').map((item) => item.textContent)).toEqual([
      'New Request',
      'Duplicate Tab',
      'Close Tab',
      'Force Close Tab',
      'Close Other Tabs',
      'Close All Tabs',
      'Close Tabs to the Left',
      'Close Tabs to the Right',
      'Force Close All Tabs',
      'Reveal in Sidebar',
    ])
    await userEvent.click(within(menu).getByRole('menuitem', { name: 'Duplicate Tab' }))
    expect(within(strip()).getAllByRole('tab').map((tab) => tab.textContent)).toEqual(['GETUntitled Request×', 'POSTRequest A×', 'POSTRequest A×', 'POSTRequest B×'])
    expect(screen.queryByRole('menu')).not.toBeInTheDocument()

    fireEvent.contextMenu(within(strip()).getAllByRole('tab')[2]!, { clientX: 10, clientY: 10 })
    await userEvent.click(screen.getByRole('menuitem', { name: 'Close Tabs to the Right' }))
    expect(within(strip()).getAllByRole('tab')).toHaveLength(3)
    fireEvent.contextMenu(within(strip()).getAllByRole('tab')[2]!, { clientX: 10, clientY: 10 })
    await userEvent.click(screen.getByRole('menuitem', { name: 'Close Tabs to the Left' }))
    expect(within(strip()).getAllByRole('tab').map((tab) => tab.textContent)).toEqual(['POSTRequest A×'])
  })

  it('disables the items that do not apply and asks before closing dirty tabs from the menu', async () => {
    renderStrip()
    const [only] = within(strip()).getAllByRole('tab')
    fireEvent.contextMenu(only!, { clientX: 10, clientY: 10 })
    expect(screen.getByRole('menuitem', { name: 'Close Other Tabs' })).toHaveAttribute('aria-disabled', 'true')
    expect(screen.getByRole('menuitem', { name: 'Close Tabs to the Left' })).toHaveAttribute('aria-disabled', 'true')
    expect(screen.getByRole('menuitem', { name: 'Close Tabs to the Right' })).toHaveAttribute('aria-disabled', 'true')
    expect(screen.getByRole('menuitem', { name: 'Reveal in Sidebar' })).toHaveAttribute('aria-disabled', 'true')
    await userEvent.click(screen.getByRole('menuitem', { name: 'Reveal in Sidebar' }))
    expect(screen.getByRole('menu')).toBeInTheDocument()
    await userEvent.keyboard('{Escape}')
    act(() => useRequestBuilderStore.getState().setUrl(useTabsStore.getState().activeTabId!, 'https://edited'))
    fireEvent.contextMenu(within(strip()).getAllByRole('tab')[0]!, { clientX: 10, clientY: 10 })
    await userEvent.click(screen.getByRole('menuitem', { name: 'Close All Tabs' }))
    expect(screen.getByRole('dialog')).toBeInTheDocument()
  })

  it('force closes every tab from the menu without asking', async () => {
    await act(() => useTabsStore.getState().openPinned('A'))
    act(() => useRequestBuilderStore.getState().setUrl(useTabsStore.getState().activeTabId!, 'https://edited'))
    renderStrip()
    fireEvent.contextMenu(within(strip()).getAllByRole('tab')[1]!, { clientX: 10, clientY: 10 })
    await userEvent.click(screen.getByRole('menuitem', { name: 'Force Close All Tabs' }))
    expect(within(strip()).queryAllByRole('tab')).toHaveLength(0)
    expect(screen.queryByRole('dialog')).not.toBeInTheDocument()
  })
})

