import { act, fireEvent, render, screen, within } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { beforeEach, describe, expect, it, vi } from 'vitest'
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
vi.mock('@/platform/services/workspaces', () => ({ listWorkspaces: vi.fn(), createWorkspace: vi.fn(), activateWorkspace: vi.fn() }))
vi.mock('@/platform/services/dialogs', () => ({ pickFolder: vi.fn(), pickFile: vi.fn(), pickSaveFile: vi.fn() }))
vi.mock('@/platform/services/tree', () => ({ getTree: vi.fn(async () => []), searchTree: vi.fn() }))

const requests = await import('@/platform/services/requests')
const { useWorkspacesStore } = await import('@/features/workspaces/store')
const { useRequestBuilderStore } = await import('@/features/request-builder/store')
const { useSidebarStore } = await import('@/features/sidebar/store')
const { useTabsStore } = await import('@/features/tabs/store')
const { TabStrip } = await import('./TabStrip')
const { CloseTabPrompt } = await import('./CloseTabPrompt')
const { SaveRequestDialog } = await import('./SaveRequestDialog')
const { TabNotice } = await import('./TabNotice')

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
    vi.clearAllMocks()
    vi.mocked(requests.getRequest).mockImplementation(async (_localId, id) => ({ draft: { ...emptyDraft(), id, name: `Request ${id}`, method: 'POST' }, revision: 'rev' }))
    useWorkspacesStore.setState({ activeLocalId: 'ws' })
    useRequestBuilderStore.getState().reset()
    useTabsStore.getState().reset()
    useTabsStore.getState().hydrate('ws', defaultUIState())
  })

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
    await userEvent.click(screen.getByRole('button', { name: 'New tab' }))
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
    const prompt = screen.getByRole('dialog', { name: 'Save changes?' })
    await userEvent.click(within(prompt).getByRole('button', { name: 'Cancel' }))
    expect(within(strip()).getAllByRole('tab')).toHaveLength(1)
    fireEvent.keyDown(within(strip()).getByRole('tab'), { key: 'Delete' })
    await userEvent.click(within(screen.getByRole('dialog')).getByRole('button', { name: 'Don’t save' }))
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
    const dialog = screen.getByRole('dialog', { name: 'Save request' })
    const name = within(dialog).getByRole('textbox', { name: 'Request name' })
    await userEvent.clear(name)
    await userEvent.type(name, 'Users')
    await userEvent.click(within(dialog).getByRole('radio', { name: 'Invoices' }))
    await userEvent.click(within(dialog).getByRole('button', { name: 'Save' }))
    expect(requests.saveRequestAs).toHaveBeenCalledWith('ws', 'f1', 'Users', expect.anything())
    expect(screen.queryByRole('dialog')).not.toBeInTheDocument()
    expect(useTabsStore.getState().tabs[0]?.requestId).toBe('N')
  })

  it('suggests a copy name for save as and can be cancelled', async () => {
    await act(() => useTabsStore.getState().openPinned('A'))
    renderStrip()
    act(() => useTabsStore.getState().saveAs())
    const dialog = screen.getByRole('dialog', { name: 'Save request as' })
    expect(within(dialog).getByRole('textbox', { name: 'Request name' })).toHaveValue('Request A Copy')
    expect(within(dialog).getByText('Create a collection in the sidebar first.')).toBeInTheDocument()
    expect(within(dialog).getByRole('button', { name: 'Save' })).toBeEnabled()
    await userEvent.click(within(dialog).getByRole('button', { name: 'Save' }))
    expect(within(dialog).getByText('Choose where to save the request.')).toBeInTheDocument()
    expect(requests.saveRequestAs).not.toHaveBeenCalled()
    const name = within(dialog).getByRole('textbox', { name: 'Request name' })
    await userEvent.clear(name)
    await userEvent.click(within(dialog).getByRole('button', { name: 'Save' }))
    expect(name).toHaveFocus()
    expect(name).toHaveAttribute('aria-invalid', 'true')
    expect(name).toHaveAccessibleDescription('Enter a name for the request.')
    await userEvent.click(within(dialog).getByRole('button', { name: 'Cancel' }))
    expect(useTabsStore.getState().saveDialog).toBeNull()
  })
})

describe('TabNotice', () => {
  beforeEach(() => {
    vi.clearAllMocks()
    vi.mocked(requests.getRequest).mockImplementation(async (_localId, id) => ({ draft: { ...emptyDraft(), id, name: 'Disk' }, revision: 'rev-disk' }))
    useWorkspacesStore.setState({ activeLocalId: 'ws' })
    useRequestBuilderStore.getState().reset()
    useTabsStore.getState().reset()
    useTabsStore.getState().hydrate('ws', defaultUIState())
  })

  it('offers keep mine or load from disk for a changed file', async () => {
    await act(() => useTabsStore.getState().openPinned('A'))
    const tabId = useTabsStore.getState().activeTabId!
    act(() => useRequestBuilderStore.getState().setUrl(tabId, 'https://mine'))
    await act(() => useTabsStore.getState().handleChangedOnDisk({ localId: 'ws', requestId: 'A', deleted: false, revision: 'new' }))
    render(<TabNotice tabId={tabId} />)
    expect(screen.getByRole('status')).toHaveTextContent('changed on disk')
    vi.mocked(requests.saveRequest).mockResolvedValue({ draft: { ...emptyDraft(), id: 'A' }, revision: 'rev-2' })
    await userEvent.click(screen.getByRole('button', { name: 'Keep mine' }))
    expect(requests.saveRequest).toHaveBeenCalledWith('ws', expect.anything(), 'rev-disk', true)
    expect(screen.queryByRole('status')).not.toBeInTheDocument()
  })

  it('offers load from disk', async () => {
    await act(() => useTabsStore.getState().openPinned('A'))
    const tabId = useTabsStore.getState().activeTabId!
    act(() => useRequestBuilderStore.getState().setUrl(tabId, 'https://mine'))
    await act(() => useTabsStore.getState().handleChangedOnDisk({ localId: 'ws', requestId: 'A', deleted: false, revision: 'new' }))
    render(<TabNotice tabId={tabId} />)
    await userEvent.click(screen.getByRole('button', { name: 'Load from disk' }))
    expect(useRequestBuilderStore.getState().tabs[tabId]!.dirty).toBe(false)
  })

  it('offers keeping a deleted request as an unsaved tab or closing it', async () => {
    await act(() => useTabsStore.getState().openPinned('A'))
    const tabId = useTabsStore.getState().activeTabId!
    act(() => useRequestBuilderStore.getState().setUrl(tabId, 'https://mine'))
    act(() => useTabsStore.getState().handleDeleted(['A']))
    const { unmount } = render(<TabNotice tabId={tabId} />)
    expect(screen.getByRole('status')).toHaveTextContent('was deleted')
    await userEvent.click(screen.getByRole('button', { name: 'Keep as unsaved tab' }))
    expect(useTabsStore.getState().tabs.find((tab) => tab.id === tabId)?.requestId).toBeNull()
    unmount()
    await act(() => useTabsStore.getState().openPinned('B'))
    const other = useTabsStore.getState().activeTabId!
    act(() => useRequestBuilderStore.getState().setUrl(other, 'https://x'))
    act(() => useTabsStore.getState().handleDeleted(['B']))
    render(<TabNotice tabId={other} />)
    await userEvent.click(screen.getByRole('button', { name: 'Close' }))
    expect(useTabsStore.getState().tabs.some((tab) => tab.id === other)).toBe(false)
  })
})
