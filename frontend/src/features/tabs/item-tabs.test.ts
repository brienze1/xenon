import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest'
import { AppError } from '@/platform/errors'
import { emptyAuth, emptyDraft } from '@/platform/mappers/request'
import type { ItemSettings, TreeNode } from '@/platform/mappers/tree'
import { defaultUIState } from '@/platform/mappers/ui-state'
import type { VariableRow } from '@/platform/mappers/variable'

vi.mock('@/platform/services/requests', () => ({
  applyUrl: vi.fn(),
  applyParams: vi.fn(),
  autoHeaders: vi.fn(),
  getRequest: vi.fn(),
  saveRequest: vi.fn(),
  saveRequestAs: vi.fn(),
  releaseResponse: vi.fn(),
  parseBulk: vi.fn(),
  formatBulk: vi.fn(),
}))
vi.mock('@/platform/services/ui-state', () => ({ getUIState: vi.fn(), saveUIState: vi.fn() }))
vi.mock('@/platform/services/environments', () => ({ listEnvironments: vi.fn(), setActiveEnvironment: vi.fn() }))
vi.mock('@/platform/services/variables', () => ({
  CONFIRM_REVEAL_REQUIRED: 'VAR-01409',
  getVariableTable: vi.fn(),
  saveVariableTable: vi.fn(),
}))
vi.mock('@/platform/services/vault', () => ({ PASSPHRASE_SETUP_REQUIRED: 'VLT-05409', VAULT_LOCKED: 'VLT-04409' }))
vi.mock('@/platform/services/tree', () => ({
  getTree: vi.fn(),
  searchTree: vi.fn(),
  getItemSettings: vi.fn(),
  saveItemSettings: vi.fn(),
  getDocumentation: vi.fn(),
  previewDocumentation: vi.fn(),
  planDocumentation: vi.fn(),
  applyDocumentation: vi.fn(),
}))
vi.mock('@/platform/services/history', () => ({ listHistory: vi.fn() }))

const requests = await import('@/platform/services/requests')
const tree = await import('@/platform/services/tree')
const variables = await import('@/platform/services/variables')
const { useWorkspacesStore } = await import('@/features/workspaces/store')
const { useRequestBuilderStore } = await import('@/features/request-builder/store')
const { useEnvironmentsStore } = await import('@/features/environments/store')
const { useItemSettingsStore } = await import('@/features/item-settings/store')
const { useDocumentationStore } = await import('@/features/documentation/store')
const { useSidebarStore } = await import('@/features/sidebar/store')
const { useResponseViewerStore } = await import('@/features/response-viewer/store')
const { useTabsStore } = await import('./store')

function node(id: string, kind: TreeNode['kind'], name: string, children: TreeNode[] = []): TreeNode {
  return { id, kind, name, method: null, url: null, error: null, warning: null, children }
}

const TREE = [node('c1', 'collection', 'Billing', [node('f1', 'folder', 'Invoices')]), node('c2', 'collection', 'Users')]

function settingsOf(id: string): ItemSettings {
  const folder = id.startsWith('f')
  return {
    id,
    kind: folder ? 'folder' : 'collection',
    name: `Item ${id}`,
    description: '',
    auth: { ...emptyAuth(), type: 'inherit' },
    headers: [],
    scripts: { preRequest: '', postResponse: '', disabled: [] },
    inheritScripts: folder ? true : null,
  }
}

const ROW: VariableRow = {
  key: 'base_url',
  initialValue: 'https://dev',
  currentValue: 'https://dev',
  hasCurrentOverride: false,
  secret: false,
  secretState: null,
  enabled: true,
  description: '',
  duplicate: false,
  surroundingSpaces: false,
}

function tabs() {
  return useTabsStore.getState()
}

function items() {
  return useItemSettingsStore.getState()
}

async function openAndEdit(itemId: string, kind: 'collection' | 'folder', mode: 'preview' | 'pinned' = 'pinned') {
  tabs().openItemTab({ kind, itemId }, mode)
  await items().open(itemId)
  items().edit(itemId, { description: 'edited' })
  return tabs().activeTabId!
}

describe('collection and folder tabs', () => {
  beforeEach(() => {
    vi.resetAllMocks()
    // The debounced documentation preview is not under test here: keep it pending so it never lands after a test
    vi.mocked(tree.previewDocumentation).mockReturnValue(new Promise(() => {}))
    tabs().reset()
    items().reset()
    useDocumentationStore.getState().reset()
    useRequestBuilderStore.getState().reset()
    useResponseViewerStore.getState().reset()
    useSidebarStore.getState().reset()
    useEnvironmentsStore.getState().reset()
    vi.mocked(requests.autoHeaders).mockResolvedValue([])
    vi.mocked(requests.releaseResponse).mockResolvedValue()
    vi.mocked(requests.getRequest).mockImplementation(async (_localId, id) => ({ draft: { ...emptyDraft(), id, name: `Request ${id}` }, revision: `rev-${id}` }))
    vi.mocked(tree.getItemSettings).mockImplementation(async (_localId, id) => settingsOf(id))
    vi.mocked(tree.saveItemSettings).mockImplementation(async (_localId, settings) => settings)
    vi.mocked(variables.getVariableTable).mockImplementation(async (_localId, scope) => ({ scope, rows: [ROW], warnings: [] }))
    useWorkspacesStore.setState({ activeLocalId: 'ws' })
    useSidebarStore.setState({ tree: TREE, status: 'ready' })
    tabs().hydrate('ws', { ...defaultUIState(), tabs: [{ tabId: 'start', requestId: 'r0', pinned: true, baseRevision: null, draft: null }], activeTabId: 'start' })
  })

  afterEach(() => {
    useWorkspacesStore.setState({ activeLocalId: null })
  })

  it('previews a collection on a single click, reuses the preview for a folder and pins on a double click', () => {
    tabs().openItemTab({ kind: 'collection', itemId: 'c1' }, 'preview')
    const preview = tabs().previewTabId!
    expect(tabs().tabs.find((tab) => tab.id === preview)).toMatchObject({ kind: 'collection', itemId: 'c1', pinned: false, requestId: null })
    expect(tabs().activeTabId).toBe(preview)
    expect(useRequestBuilderStore.getState().tabs[preview]).toBeUndefined()
    tabs().openItemTab({ kind: 'folder', itemId: 'f1' }, 'preview')
    expect(tabs().tabs).toHaveLength(2)
    expect(tabs().tabs[1]).toMatchObject({ id: preview, kind: 'folder', itemId: 'f1' })
    tabs().openItemTab({ kind: 'folder', itemId: 'f1' }, 'pinned')
    expect(tabs().previewTabId).toBeNull()
    expect(tabs().tabs[1]?.pinned).toBe(true)
  })

  it('focuses the tab that already shows the item instead of opening another', () => {
    tabs().openItemTab({ kind: 'collection', itemId: 'c1' }, 'pinned')
    const collection = tabs().activeTabId
    tabs().activate('start')
    tabs().openItemTab({ kind: 'collection', itemId: 'c1' }, 'preview')
    expect(tabs().activeTabId).toBe(collection)
    expect(tabs().tabs).toHaveLength(2)
  })

  it('turns a clean request preview into a collection preview and back, dropping what the preview held', async () => {
    await tabs().openPreview('r1')
    const preview = tabs().previewTabId!
    tabs().openItemTab({ kind: 'collection', itemId: 'c1' }, 'preview')
    expect(tabs().tabs.find((tab) => tab.id === preview)).toMatchObject({ kind: 'collection', itemId: 'c1', requestId: null })
    expect(useRequestBuilderStore.getState().tabs[preview]).toBeUndefined()
    await items().open('c1')
    await tabs().openPreview('r2')
    expect(tabs().tabs.find((tab) => tab.id === preview)).toMatchObject({ requestId: 'r2', kind: undefined, itemId: undefined })
    expect(items().items.c1).toBeUndefined()
  })

  it('pins a previewed collection on its first edit and asks before closing it', async () => {
    const tabId = await openAndEdit('c1', 'collection', 'preview')
    tabs().pinIfEdited()
    expect(tabs().previewTabId).toBeNull()
    tabs().requestClose(tabId)
    expect(tabs().closePrompt).toBe(tabId)
    await tabs().resolveClose('cancel')
    expect(tabs().tabs.some((tab) => tab.id === tabId)).toBe(true)
    tabs().requestClose(tabId)
    await tabs().resolveClose('discard')
    expect(tabs().tabs.some((tab) => tab.id === tabId)).toBe(false)
    expect(items().items.c1).toBeUndefined()
    expect(tree.saveItemSettings).not.toHaveBeenCalled()
  })

  it('saves the settings from the close prompt and closes the tab', async () => {
    const tabId = await openAndEdit('f1', 'folder')
    tabs().requestClose(tabId)
    await tabs().resolveClose('save')
    expect(tree.saveItemSettings).toHaveBeenCalledWith('ws', expect.objectContaining({ id: 'f1', description: 'edited', inheritScripts: true }))
    expect(tabs().closePrompt).toBeNull()
    expect(tabs().tabs.some((tab) => tab.id === tabId)).toBe(false)
  })

  it('keeps the tab and shows the error when saving from the close prompt fails', async () => {
    const tabId = await openAndEdit('c1', 'collection')
    vi.mocked(tree.saveItemSettings).mockRejectedValue(new AppError('TRE-05500', 'The folder is read-only'))
    tabs().requestClose(tabId)
    await tabs().resolveClose('save')
    expect(tabs().closePrompt).toBe(tabId)
    expect(tabs().closeError?.code).toBe('TRE-05500')
    expect(items().items.c1?.failure?.error.code).toBe('TRE-05500')
  })

  it('saves the settings and the collection variables of the active tab with Save', async () => {
    await openAndEdit('c1', 'collection')
    const environments = useEnvironmentsStore.getState()
    await environments.loadTable('collection:c1')
    environments.editDraft('collection:c1', (rows) => rows.map((row) => ({ ...row, description: 'edited' })))
    expect(tabs().closePrompt).toBeNull()
    vi.mocked(variables.saveVariableTable).mockResolvedValue({ scope: 'collection:c1', rows: [ROW], warnings: [] })
    await tabs().save()
    expect(tree.saveItemSettings).toHaveBeenCalledWith('ws', expect.objectContaining({ id: 'c1', description: 'edited' }))
    expect(variables.saveVariableTable).toHaveBeenCalledWith('ws', 'collection:c1', expect.any(Array), false)
    expect(items().items.c1?.dirty).toBe(false)
    tabs().saveAs()
    expect(tabs().saveDialog).toBeNull()
    expect(tabs().duplicateTab(tabs().activeTabId!)).toBeNull()
  })

  it('counts edited documentation as unsaved and asks to confirm its plan before closing the collection tab', async () => {
    const plan = { name: 'Billing', collectionChanged: true, updated: [], created: [], deleted: [] }
    vi.mocked(tree.getDocumentation).mockResolvedValue({ yaml: 'openapi: 3.1.0\n', document: { title: 'Billing', description: '', version: '1.0.0', servers: [], tags: [], tagGroups: [], operations: [], securitySchemes: [] } })
    vi.mocked(tree.planDocumentation).mockResolvedValue(plan)
    tabs().openItemTab({ kind: 'collection', itemId: 'c1' }, 'preview')
    const tabId = tabs().activeTabId!
    await items().open('c1')
    const documentation = useDocumentationStore.getState()
    await documentation.open('c1')
    documentation.edit('c1', 'openapi: 3.1.0\ninfo:\n  title: Payments\n')
    tabs().pinIfEdited()
    expect(tabs().tabs.find((tab) => tab.id === tabId)?.pinned).toBe(true)
    tabs().requestClose(tabId)
    expect(tabs().closePrompt).toBe(tabId)
    await tabs().resolveClose('save')
    expect(tree.planDocumentation).toHaveBeenCalledWith('ws', 'c1', 'openapi: 3.1.0\ninfo:\n  title: Payments\n')
    expect(useDocumentationStore.getState().entries.c1?.plan).toEqual(plan)
    expect(items().items.c1?.section).toBe('documentation')
    expect(tabs().tabs.some((tab) => tab.id === tabId)).toBe(true)
    expect(tabs().activeTabId).toBe(tabId)
    tabs().forceClose(tabId)
    expect(useDocumentationStore.getState().entries.c1).toBeUndefined()
  })

  it('counts edited folder documentation as unsaved and plans its save from the folder tab', async () => {
    const plan = { name: 'Invoices', collectionChanged: false, updated: [], created: [], deleted: [] }
    vi.mocked(tree.getDocumentation).mockResolvedValue({ yaml: 'openapi: 3.1.0\n', document: { title: 'Invoices', description: '', version: '1.0.0', servers: [], tags: [], tagGroups: [], operations: [], securitySchemes: [] } })
    vi.mocked(tree.planDocumentation).mockResolvedValue({ ...plan, created: [{ id: 'n1', name: 'New', folder: '', method: 'GET', path: '/new', revision: '' }] })
    tabs().openItemTab({ kind: 'folder', itemId: 'f1' }, 'pinned')
    const tabId = tabs().activeTabId!
    await items().open('f1')
    const documentation = useDocumentationStore.getState()
    await documentation.open('f1')
    documentation.edit('f1', 'openapi: 3.1.0\ninfo:\n  title: Invoices\n')
    tabs().requestClose(tabId)
    expect(tabs().closePrompt).toBe(tabId)
    await tabs().resolveClose('save')
    expect(tree.planDocumentation).toHaveBeenCalledWith('ws', 'f1', 'openapi: 3.1.0\ninfo:\n  title: Invoices\n')
    expect(items().items.f1?.section).toBe('documentation')
    expect(tabs().tabs.some((tab) => tab.id === tabId)).toBe(true)
  })

  it('counts unsaved collection variables as unsaved changes of the collection tab', async () => {
    tabs().openItemTab({ kind: 'collection', itemId: 'c1' }, 'pinned')
    const tabId = tabs().activeTabId!
    await items().open('c1')
    await useEnvironmentsStore.getState().loadTable('collection:c1')
    useEnvironmentsStore.getState().editDraft('collection:c1', (rows) => rows.map((row) => ({ ...row, secret: true })))
    tabs().requestClose(tabId)
    expect(tabs().closePrompt).toBe(tabId)
    tabs().activate('start')
    await tabs().resolveClose('save')
    expect(tabs().activeTabId).toBe(tabId)
    expect(items().items.c1?.section).toBe('variables')
    expect(useEnvironmentsStore.getState().drafts['collection:c1']?.confirming).not.toBeNull()
  })

  it('follows the tree: closes the tabs of deleted items and keeps the rest', async () => {
    tabs().openItemTab({ kind: 'collection', itemId: 'c1' }, 'pinned')
    tabs().openItemTab({ kind: 'folder', itemId: 'f1' }, 'pinned')
    tabs().openItemTab({ kind: 'collection', itemId: 'c2' }, 'pinned')
    await openAndEdit('c2', 'collection')
    tabs().handleTreeChanged([node('c1', 'collection', 'Billing')])
    expect(tabs().tabs.map((tab) => tab.itemId ?? 'request')).toEqual(['request', 'c1'])
    expect(tabs().closedTabs).toHaveLength(0)
    expect(items().items.c2).toBeUndefined()
  })

  it('updates the kind when a moved folder became a collection', () => {
    tabs().openItemTab({ kind: 'folder', itemId: 'f1' }, 'pinned')
    tabs().handleTreeChanged([node('f1', 'collection', 'Invoices')])
    expect(tabs().tabs.at(-1)).toMatchObject({ kind: 'collection', itemId: 'f1' })
  })

  it('persists collection and folder tabs and restores them', () => {
    tabs().openItemTab({ kind: 'collection', itemId: 'c1' }, 'preview')
    tabs().openItemTab({ kind: 'folder', itemId: 'f1' }, 'pinned')
    const snapshot = tabs().snapshot()
    expect(snapshot.tabs.slice(1)).toEqual([
      { tabId: expect.any(String), kind: 'collection', requestId: null, itemId: 'c1', pinned: false, baseRevision: null, draft: null },
      { tabId: expect.any(String), kind: 'folder', requestId: null, itemId: 'f1', pinned: true, baseRevision: null, draft: null },
    ])
    tabs().reset()
    tabs().hydrate('ws', snapshot)
    expect(tabs().tabs.map((tab) => [tab.kind ?? 'request', tab.itemId ?? null])).toEqual([
      ['request', null],
      ['collection', 'c1'],
      ['folder', 'f1'],
    ])
    expect(tabs().previewTabId).toBe(snapshot.previewTabId)
    expect(Object.keys(useRequestBuilderStore.getState().tabs)).toEqual(['start'])
  })

  it('reopens a closed collection tab unless the collection is gone', () => {
    tabs().openItemTab({ kind: 'folder', itemId: 'f1' }, 'pinned')
    tabs().requestClose()
    expect(tabs().closedTabs.at(-1)?.tab).toMatchObject({ kind: 'folder', itemId: 'f1' })
    void tabs().reopenClosed()
    expect(tabs().tabs.at(-1)).toMatchObject({ kind: 'folder', itemId: 'f1', pinned: true })
    tabs().requestClose()
    useSidebarStore.setState({ tree: [node('c2', 'collection', 'Users')] })
    void tabs().reopenClosed()
    expect(tabs().tabs.some((tab) => tab.itemId === 'f1')).toBe(false)
  })

  it('reveals a folder tab in the Collections view', () => {
    useSidebarStore.setState({ view: 'history' })
    tabs().openItemTab({ kind: 'folder', itemId: 'f1' }, 'pinned')
    tabs().revealInSidebar(tabs().activeTabId!)
    expect(useSidebarStore.getState()).toMatchObject({ view: 'collections', selectedId: 'f1', expanded: { c1: true } })
  })
})
