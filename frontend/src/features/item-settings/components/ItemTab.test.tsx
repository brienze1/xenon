import { act, render, screen, within } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { beforeEach, describe, expect, it, vi } from 'vitest'
import { AppError } from '@/platform/errors'
import { emptyAuth } from '@/platform/mappers/request'
import type { ItemSettings, TreeNode } from '@/platform/mappers/tree'
import type { ItemAuthSlot, ItemScriptsSlot } from '@/features/item-settings/model'

vi.mock('@/platform/services/tree', () => ({
  getTree: vi.fn(),
  searchTree: vi.fn(),
  renameItem: vi.fn(),
  getItemSettings: vi.fn(),
  saveItemSettings: vi.fn(),
}))
vi.mock('@/platform/services/requests', () => ({ parseBulk: vi.fn(), formatBulk: vi.fn() }))
vi.mock('@/platform/services/workspaces', () => ({ listWorkspaces: vi.fn(), createWorkspace: vi.fn(), activateWorkspace: vi.fn() }))

const tree = await import('@/platform/services/tree')
const requests = await import('@/platform/services/requests')
const { useWorkspacesStore } = await import('@/features/workspaces/store')
const { useSidebarStore } = await import('@/features/sidebar/store')
const { useItemSettingsStore } = await import('@/features/item-settings/store')
const { useDocumentationStore } = await import('@/features/documentation/store')
const { useTabsStore } = await import('@/features/tabs/store')
const { ItemTab } = await import('./ItemTab')

function node(id: string, kind: TreeNode['kind'], name: string, children: TreeNode[] = []): TreeNode {
  return { id, kind, name, method: null, url: null, error: null, warning: null, children }
}

const collection: ItemSettings = {
  id: 'c1',
  kind: 'collection',
  name: 'Billing',
  description: 'The **billing** API',
  auth: { ...emptyAuth(), type: 'inherit' },
  headers: [],
  scripts: { preRequest: 'console.log(1)', postResponse: '', disabled: [] },
  inheritScripts: null,
}

const folder: ItemSettings = { ...collection, id: 'f1', kind: 'folder', name: 'Invoices', description: '', inheritScripts: true }

function authEditor(slot: ItemAuthSlot) {
  return (
    <label>
      Token for {slot.kind} {slot.itemId} ({slot.auth.type})
      <input value={slot.auth.bearer.token} onChange={(event) => slot.onChange({ ...slot.auth, type: 'bearer', bearer: { token: event.target.value } })} />
    </label>
  )
}

function scriptsEditor(slot: ItemScriptsSlot) {
  return (
    <div>
      <p>
        Editing {slot.itemId}: {slot.scripts.preRequest} (inherit: {String(slot.inheritScripts)})
      </p>
      <button type="button" onClick={() => slot.onChange({ ...slot.scripts, preRequest: 'edited()' })}>
        Edit script
      </button>
      <button type="button" onClick={() => slot.onInheritChange(false)}>
        Stop inheriting
      </button>
    </div>
  )
}

function collectionVariables(collectionId: string) {
  return <p>Variables of {collectionId}</p>
}

function documentation(itemId: string, kind: string) {
  return (
    <p>
      Documentation of {kind} {itemId}
    </p>
  )
}

const EDITED = {
  status: 'ready',
  loadError: null,
  base: null,
  draft: 'edited',
  dirty: true,
  view: 'yaml',
  preview: null,
  problem: null,
  plan: null,
  saving: false,
  failure: null,
  stale: false,
} as const

function runner(itemId: string, kind: string) {
  return (
    <p>
      Runner of {kind} {itemId}
    </p>
  )
}

function renderTab(settings: ItemSettings, onSave = vi.fn(), withRunner = false) {
  vi.mocked(tree.getItemSettings).mockResolvedValue(settings)
  render(
    <ItemTab
      itemId={settings.id}
      kind={settings.kind}
      onSave={onSave}
      authEditor={authEditor}
      scriptsEditor={scriptsEditor}
      collectionVariables={collectionVariables}
      documentation={documentation}
      runner={withRunner ? runner : undefined}
    />,
  )
  return onSave
}

function store() {
  return useItemSettingsStore.getState()
}

describe('ItemTab', () => {
  beforeEach(() => {
    vi.resetAllMocks()
    store().reset()
    useSidebarStore.getState().reset()
    useWorkspacesStore.setState({ activeLocalId: 'ws' })
    useSidebarStore.setState({ tree: [node('c1', 'collection', 'Billing', [node('f1', 'folder', 'Invoices')])], status: 'ready' })
    vi.mocked(tree.saveItemSettings).mockImplementation(async (_localId, settings) => settings)
    useTabsStore.setState({
      tabs: [
        { id: 'tc', requestId: null, pinned: true, kind: 'collection', itemId: 'c1' },
        { id: 'tf', requestId: null, pinned: true, kind: 'folder', itemId: 'f1' },
      ],
      activeTabId: 'tc',
      nameErrors: {},
    })
  })

  it('opens on the Overview with the name and the description rendered as markdown', async () => {
    renderTab(collection)
    const tab = await screen.findByRole('region', { name: 'Billing collection' })
    expect(within(tab).getByRole('tab', { name: 'Overview' })).toHaveAttribute('aria-selected', 'true')
    expect(within(tab).getAllByRole('tab').map((item) => item.textContent)).toEqual(['Overview', 'Authorization', 'Headers', 'Scripts', 'Variables', 'Documentation'])
    expect(within(tab).getByRole('heading', { level: 2 })).toHaveTextContent('Billing')
    expect(within(tab).getByRole('button', { name: /^Rename collection/ })).toHaveTextContent('Billing')
    expect(within(tab).queryByRole('textbox', { name: /name/i })).not.toBeInTheDocument()
    expect(await within(tab).findByText('billing', { selector: 'strong' })).toBeInTheDocument()
    expect(within(tab).getByRole('button', { name: 'Save' })).toBeDisabled()
  })

  it('adds the Runner section last for collections and folders, on the section the store asks for', async () => {
    renderTab(collection, vi.fn(), true)
    const tab = await screen.findByRole('region', { name: 'Billing collection' })
    expect(within(tab).getAllByRole('tab').map((item) => item.textContent)).toEqual(['Overview', 'Authorization', 'Headers', 'Scripts', 'Variables', 'Documentation', 'Runner'])
    await userEvent.click(within(tab).getByRole('tab', { name: 'Runner' }))
    expect(store().items.c1?.section).toBe('runner')
    expect(within(tab).getByRole('tabpanel')).toHaveTextContent('Runner of collection c1')
  })

  it('opens a folder on its Runner when Run… asked for it', async () => {
    vi.mocked(tree.getItemSettings).mockResolvedValue(folder)
    store().setSection('f1', 'runner')
    renderTab(folder, vi.fn(), true)
    const tab = await screen.findByRole('region', { name: 'Invoices folder' })
    expect(within(tab).getByRole('tab', { name: 'Runner' })).toHaveAttribute('aria-selected', 'true')
    expect(within(tab).getByRole('tabpanel')).toHaveTextContent('Runner of folder f1')
  })

  it('falls back to the Overview when no Runner is composed', async () => {
    vi.mocked(tree.getItemSettings).mockResolvedValue(collection)
    store().setSection('c1', 'runner')
    renderTab(collection)
    const tab = await screen.findByRole('region', { name: 'Billing collection' })
    expect(within(tab).queryByRole('tab', { name: 'Runner' })).not.toBeInTheDocument()
    expect(within(tab).getByRole('tab', { name: 'Overview' })).toHaveAttribute('aria-selected', 'true')
  })

  it('edits the description in markdown and saves it', async () => {
    const onSave = renderTab(collection)
    await screen.findByRole('region', { name: 'Billing collection' })
    await userEvent.click(screen.getByRole('button', { name: 'Edit Description' }))
    const description = screen.getByRole('textbox', { name: 'Description' })
    await userEvent.clear(description)
    await userEvent.type(description, 'New text')
    expect(store().items.c1).toMatchObject({ dirty: true, draft: { description: 'New text' } })
    await userEvent.click(screen.getByRole('button', { name: 'Preview' }))
    expect(await screen.findByText('New text')).toBeInTheDocument()
    await userEvent.click(screen.getByRole('button', { name: 'Save' }))
    expect(onSave).toHaveBeenCalled()
  })

  it('edits the collection name in place from its heading as a draft, and enables Save', async () => {
    const onSave = renderTab(collection)
    const tab = await screen.findByRole('region', { name: 'Billing collection' })
    await userEvent.click(within(tab).getByRole('button', { name: /^Rename collection/ }))
    const input = within(tab).getByRole('textbox', { name: 'Rename collection' }) as HTMLInputElement
    expect([input.selectionStart, input.selectionEnd]).toEqual([0, 'Billing'.length])
    await userEvent.clear(input)
    await userEvent.type(input, 'Payments{Enter}')
    expect(store().items.c1?.draft?.name).toBe('Payments')
    expect(store().items.c1?.dirty).toBe(false)
    expect(tree.renameItem).not.toHaveBeenCalled()
    expect(within(tab).getByRole('button', { name: /^Rename collection/ })).toHaveTextContent('Payments')
    expect(within(tab).getByRole('button', { name: 'Save' })).toBeEnabled()
    await userEvent.click(within(tab).getByRole('button', { name: 'Save' }))
    expect(onSave).toHaveBeenCalledTimes(1)
  })

  it('edits a folder name in place from the keyboard, and Escape restores it', async () => {
    renderTab(folder)
    const tab = await screen.findByRole('region', { name: 'Invoices folder' })
    within(tab).getByRole('button', { name: /^Rename folder/ }).focus()
    await userEvent.keyboard('{Enter}')
    const input = within(tab).getByRole('textbox', { name: 'Rename folder' })
    expect(input).toHaveFocus()
    await userEvent.type(input, ' 2024')
    expect(store().items.f1?.draft?.name).toBe('Invoices 2024')
    await userEvent.keyboard('{Escape}')
    expect(store().items.f1?.draft?.name).toBe('Invoices')
    expect(within(tab).getByRole('button', { name: /^Rename folder/ })).toHaveFocus()
  })

  it('shows a name error under the heading', async () => {
    useTabsStore.setState({ nameErrors: { tc: { kind: 'failed', error: new AppError('TRE-05500', 'The folder is read-only') } } })
    renderTab(collection)
    const tab = await screen.findByRole('region', { name: 'Billing collection' })
    expect(within(tab).getByRole('alert')).toHaveTextContent('The folder is read-only. Choose another name, then save.')
    expect(within(tab).getByRole('button', { name: /^Rename collection/ })).toHaveAttribute('aria-invalid', 'true')
  })

  it('edits the authorization through the composed editor', async () => {
    renderTab(collection)
    await userEvent.click(await screen.findByRole('tab', { name: 'Authorization' }))
    await userEvent.type(screen.getByRole('textbox', { name: 'Token for collection c1 (inherit)' }), 'abc')
    expect(store().items.c1?.draft?.auth).toMatchObject({ type: 'bearer', bearer: { token: 'abc' } })
    expect(screen.getByRole('button', { name: 'Save' })).toBeEnabled()
  })

  it('edits the headers in the table and in bulk', async () => {
    vi.mocked(requests.formatBulk).mockResolvedValue('')
    renderTab(collection)
    await userEvent.click(await screen.findByRole('tab', { name: 'Headers' }))
    await userEvent.click(screen.getByRole('button', { name: 'Bulk Edit' }))
    expect(store().items.c1?.bulkText).toBe('')
    await userEvent.type(screen.getByRole('textbox', { name: /Billing headers/ }), 'X-Team:payments')
    expect(store().items.c1).toMatchObject({ bulkText: 'X-Team:payments', dirty: true })
  })

  it('edits a folder’s scripts and its inherit toggle; folders have no variables', async () => {
    renderTab(folder)
    const tab = await screen.findByRole('region', { name: 'Invoices folder' })
    expect(within(tab).queryByRole('tab', { name: 'Variables' })).not.toBeInTheDocument()
    expect(within(tab).getByRole('tablist', { name: 'Folder settings' })).toBeInTheDocument()
    await userEvent.click(within(tab).getByRole('tab', { name: 'Scripts' }))
    expect(screen.getByText('Editing f1: console.log(1) (inherit: true)')).toBeInTheDocument()
    await userEvent.click(screen.getByRole('button', { name: 'Edit script' }))
    await userEvent.click(screen.getByRole('button', { name: 'Stop inheriting' }))
    expect(store().items.f1?.draft).toMatchObject({ scripts: { preRequest: 'edited()' }, inheritScripts: false })
    expect(store().items.f1?.section).toBe('scripts')
  })

  it('shows the collection variables in their own inner tab', async () => {
    renderTab(collection)
    await userEvent.click(await screen.findByRole('tab', { name: 'Variables' }))
    expect(screen.getByRole('tabpanel', { name: 'Variables' })).toHaveTextContent('Variables of c1')
  })

  it('warns when the settings changed on disk during an edit and can load the disk version', async () => {
    const onSave = renderTab(collection)
    await screen.findByRole('region', { name: 'Billing collection' })
    act(() => store().edit('c1', { description: 'Mine' }))
    vi.mocked(tree.getItemSettings).mockResolvedValue({ ...collection, description: 'Theirs' })
    await act(() => store().refreshAll())
    expect(screen.getByText('These settings changed on disk while you were editing them.')).toBeInTheDocument()
    await userEvent.click(screen.getByRole('button', { name: 'Keep Your Version' }))
    expect(onSave).toHaveBeenCalled()
    await userEvent.click(screen.getByRole('button', { name: 'Load from Disk' }))
    expect(store().items.c1).toMatchObject({ dirty: false, stale: false, draft: { description: 'Theirs' } })
    expect(screen.queryByText(/changed on disk/)).not.toBeInTheDocument()
  })

  it('shows a save failure and a load failure with a way to try again', async () => {
    vi.mocked(tree.getItemSettings).mockRejectedValueOnce(new AppError('TRE-02404', 'The item does not exist'))
    render(<ItemTab itemId="c1" kind="collection" onSave={vi.fn()} />)
    expect(await screen.findByText(/The item does not exist. Check that the workspace folder exists/)).toBeInTheDocument()
    vi.mocked(tree.getItemSettings).mockResolvedValue(collection)
    await userEvent.click(screen.getByRole('button', { name: 'Try Again' }))
    await screen.findByRole('region', { name: 'Billing collection' })
    act(() => store().edit('c1', { description: 'Mine' }))
    vi.mocked(tree.saveItemSettings).mockRejectedValue(new AppError('TRE-05500', 'The folder is read-only'))
    await act(() => store().save('c1'))
    expect(screen.getByText(/The folder is read-only. Check that the workspace folder exists and is writable/)).toBeInTheDocument()
  })

  it('shows the documentation of a collection and enables Save while its YAML is edited', async () => {
    renderTab(collection)
    const tab = await screen.findByRole('region', { name: 'Billing collection' })
    await userEvent.click(within(tab).getByRole('tab', { name: 'Documentation' }))
    expect(within(tab).getByText('Documentation of collection c1')).toBeInTheDocument()
    act(() =>
      useDocumentationStore.setState({
        entries: {
          c1: {
            status: 'ready',
            loadError: null,
            base: null,
            draft: 'edited',
            dirty: true,
            view: 'yaml',
            preview: null,
            problem: null,
            plan: null,
            saving: false,
            failure: null,
            stale: false,
          },
        },
      }),
    )
    expect(within(tab).getByRole('button', { name: 'Save' })).toBeEnabled()
    act(() => useDocumentationStore.getState().reset())
  })

  it('shows the documentation of a folder and enables Save while its YAML is edited', async () => {
    renderTab(folder)
    const tab = await screen.findByRole('region', { name: 'Invoices folder' })
    await userEvent.click(within(tab).getByRole('tab', { name: 'Documentation' }))
    expect(within(tab).getByText('Documentation of folder f1')).toBeInTheDocument()
    expect(within(tab).getByRole('button', { name: 'Save' })).toBeDisabled()
    act(() => useDocumentationStore.setState({ entries: { f1: { ...useDocumentationStore.getState().entries.f1!, ...EDITED } } }))
    expect(within(tab).getByRole('button', { name: 'Save' })).toBeEnabled()
    act(() => useDocumentationStore.getState().reset())
  })
})
