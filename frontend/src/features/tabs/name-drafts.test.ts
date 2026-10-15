import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest'
import { AppError } from '@/platform/errors'
import type { Example } from '@/platform/mappers/example'
import { emptyAuth, emptyDraft } from '@/platform/mappers/request'
import type { ItemSettings, TreeNode } from '@/platform/mappers/tree'
import { defaultUIState } from '@/platform/mappers/ui-state'
import { collectionScope, type VariableRow } from '@/platform/mappers/variable'

vi.mock('@/platform/services/requests', () => ({
  applyUrl: vi.fn(),
  applyParams: vi.fn(),
  parseBulk: vi.fn(),
  formatBulk: vi.fn(),
  autoHeaders: vi.fn(),
  sendRequest: vi.fn(),
  getRequest: vi.fn(),
  saveRequest: vi.fn(),
  saveRequestAs: vi.fn(),
  lastResult: vi.fn(),
  releaseResponse: vi.fn(),
  saveResponseBody: vi.fn(),
}))
vi.mock('@/platform/services/ui-state', () => ({ getUIState: vi.fn(), saveUIState: vi.fn() }))
vi.mock('@/platform/services/environments', () => ({
  listEnvironments: vi.fn(),
  setActiveEnvironment: vi.fn(),
  getEnvironment: vi.fn(),
  updateEnvironment: vi.fn(),
}))
vi.mock('@/platform/services/variables', () => ({
  CONFIRM_REVEAL_REQUIRED: 'VAR-01409',
  getVariableTable: vi.fn(),
  saveVariableTable: vi.fn(),
  inspectVariables: vi.fn(async () => null),
  suggestVariables: vi.fn(),
}))
vi.mock('@/platform/services/vault', () => ({ PASSPHRASE_SETUP_REQUIRED: 'VLT-05409', VAULT_LOCKED: 'VLT-04409' }))
vi.mock('@/platform/services/tree', () => ({
  getTree: vi.fn(),
  searchTree: vi.fn(),
  renameItem: vi.fn(),
  getItemSettings: vi.fn(),
  saveItemSettings: vi.fn(),
  getDocumentation: vi.fn(),
  previewDocumentation: vi.fn(),
  planDocumentation: vi.fn(),
  applyDocumentation: vi.fn(),
}))
vi.mock('@/platform/services/examples', () => ({
  getExample: vi.fn(),
  saveExample: vi.fn(),
  createExampleFromResponse: vi.fn(),
  renameExample: vi.fn(),
  duplicateExample: vi.fn(),
  deleteExample: vi.fn(),
}))
vi.mock('@/platform/services/response-body', () => ({ fetchResponseText: vi.fn() }))
vi.mock('@/platform/services/history', () => ({ listHistory: vi.fn() }))
vi.mock('@/platform/services/workspaces', () => ({ listWorkspaces: vi.fn(), createWorkspace: vi.fn(), activateWorkspace: vi.fn() }))

const requests = await import('@/platform/services/requests')
const treeService = await import('@/platform/services/tree')
const variables = await import('@/platform/services/variables')
const environmentServices = await import('@/platform/services/environments')
const exampleServices = await import('@/platform/services/examples')
const { useWorkspacesStore } = await import('@/features/workspaces/store')
const { useRequestBuilderStore } = await import('@/features/request-builder/store')
const { useEnvironmentsStore } = await import('@/features/environments/store')
const { useExamplesStore } = await import('@/features/examples/store')
const { useItemSettingsStore } = await import('@/features/item-settings/store')
const { useDocumentationStore } = await import('@/features/documentation/store')
const { useSidebarStore } = await import('@/features/sidebar/store')
const { useResponseViewerStore } = await import('@/features/response-viewer/store')
const { isTabDirty, useTabsStore } = await import('./store')

function node(id: string, kind: TreeNode['kind'], name: string, children: TreeNode[] = [], examples: TreeNode['examples'] = []): TreeNode {
  return { id, kind, name, method: kind === 'request' ? 'POST' : null, url: null, error: null, warning: null, children, examples }
}

/** The names on disk; the mocked rename calls change them and the mocked tree reads them back. */
let names: Record<string, string> = {}

function diskTree(): TreeNode[] {
  return [
    node('c1', 'collection', names.c1!, [
      node('f1', 'folder', names.f1!, [node('r1', 'request', names.r1!, [], [{ id: 'x1', name: names.x1!, code: 201 }])]),
      node('r2', 'request', names.r2!),
    ]),
  ]
}

function settingsOf(id: string): ItemSettings {
  return {
    id,
    kind: id.startsWith('f') ? 'folder' : 'collection',
    name: names[id]!,
    description: '',
    auth: { ...emptyAuth(), type: 'inherit' },
    headers: [],
    scripts: { preRequest: '', postResponse: '', disabled: [] },
    inheritScripts: id.startsWith('f') ? true : null,
  }
}

function example(overrides: Partial<Example> = {}): Example {
  return {
    id: 'x1',
    requestId: 'r1',
    name: names.x1!,
    revision: 'rev-x1',
    request: { ...emptyDraft(), id: '', name: names.x1!, method: 'POST', url: 'https://x/invoices' },
    response: {
      code: 201,
      status: 'Created',
      contentType: 'application/json',
      headers: [],
      body: '{"id":1}',
      bodyEncoding: 'text',
      bodyUrl: '',
      sizeBytes: 8,
      isBinary: false,
    },
    ...overrides,
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

let environmentNames: Record<string, string> = {}

function environmentList() {
  return {
    environments: Object.entries(environmentNames).map(([id, name]) => ({ id, name, color: 'green' as const, confirmBeforeSend: false, error: null })),
    activeId: 'e1',
  }
}

let revision = 0

function tabs() {
  return useTabsStore.getState()
}

function builder() {
  return useRequestBuilderStore.getState()
}

function environments() {
  return useEnvironmentsStore.getState()
}

function deferred<T>() {
  let resolve!: (value: T) => void
  const promise = new Promise<T>((done) => (resolve = done))
  return { promise, resolve }
}

/** Whether the tab has unsaved changes, so closing it would ask first. */
function asksBeforeClosing(tabId: string): boolean {
  return isTabDirty(tabId)
}

function isOpen(tabId: string): boolean {
  return tabs().tabs.some((tab) => tab.id === tabId)
}

async function openRequest(id = 'r1'): Promise<string> {
  await tabs().openPinned(id)
  return tabs().activeTabId!
}

describe('name drafts in tabs', () => {
  beforeEach(() => {
    vi.resetAllMocks()
    names = { c1: 'Billing', f1: 'Invoices', r1: 'Create invoice', r2: 'List invoices', x1: 'Created' }
    environmentNames = { e1: 'dev', e2: 'hml' }
    revision = 0
    tabs().reset()
    builder().reset()
    useExamplesStore.getState().reset()
    useItemSettingsStore.getState().reset()
    useDocumentationStore.getState().reset()
    useResponseViewerStore.getState().reset()
    useSidebarStore.getState().reset()
    environments().reset()
    vi.mocked(treeService.previewDocumentation).mockReturnValue(new Promise(() => {}))
    vi.mocked(requests.autoHeaders).mockResolvedValue([])
    vi.mocked(requests.applyUrl).mockImplementation(async (sent) => sent)
    vi.mocked(requests.applyParams).mockImplementation(async (sent) => sent)
    vi.mocked(requests.releaseResponse).mockResolvedValue()
    vi.mocked(requests.getRequest).mockImplementation(async (_localId, id) => ({
      draft: { ...emptyDraft(), id, name: names[id]!, url: `https://x/${id}` },
      revision: `rev-${id}-${revision}`,
    }))
    vi.mocked(requests.saveRequest).mockImplementation(async (_localId, sent) => {
      revision += 1
      return { draft: sent, revision: `rev-${sent.id}-${revision}` }
    })
    vi.mocked(treeService.getTree).mockImplementation(async () => diskTree())
    vi.mocked(treeService.searchTree).mockResolvedValue({ matchIds: [], ancestorIds: [] })
    vi.mocked(treeService.renameItem).mockImplementation(async (_localId, id, name) => {
      names[id] = name
      revision += 1
      return node(id, 'request', name)
    })
    vi.mocked(treeService.getItemSettings).mockImplementation(async (_localId, id) => settingsOf(id))
    vi.mocked(treeService.saveItemSettings).mockImplementation(async (_localId, settings) => settings)
    vi.mocked(variables.getVariableTable).mockImplementation(async (_localId, scope) => ({ scope, rows: [ROW], warnings: [] }))
    vi.mocked(variables.saveVariableTable).mockImplementation(async (_localId, scope) => ({ scope, rows: [ROW], warnings: [] }))
    vi.mocked(environmentServices.listEnvironments).mockImplementation(async () => environmentList())
    vi.mocked(environmentServices.setActiveEnvironment).mockResolvedValue('e1')
    vi.mocked(environmentServices.getEnvironment).mockImplementation(async (_localId, id) => ({
      id,
      name: environmentNames[id]!,
      description: '',
      color: 'green',
      confirmBeforeSend: false,
      variables: [],
    }))
    vi.mocked(environmentServices.updateEnvironment).mockImplementation(async (_localId, update) => {
      environmentNames[update.id] = update.name
      return { ...update, variables: [] }
    })
    vi.mocked(exampleServices.getExample).mockImplementation(async () => example())
    vi.mocked(exampleServices.saveExample).mockImplementation(async (_localId, _requestId, sent) => ({ ...sent, revision: 'rev-x1-saved' }))
    vi.mocked(exampleServices.renameExample).mockImplementation(async (_localId, _requestId, id, name) => {
      names[id] = name
      return example({ name })
    })
    useWorkspacesStore.setState({ activeLocalId: 'ws' })
    useSidebarStore.setState({ tree: diskTree(), status: 'ready' })
    useEnvironmentsStore.setState({ ...environmentList(), status: 'ready' })
    tabs().hydrate('ws', defaultUIState())
  })

  afterEach(() => {
    useWorkspacesStore.setState({ activeLocalId: null })
  })

  describe('request tabs', () => {
    it('marks the tab dirty while the name differs from the saved one, and clean again after Escape', async () => {
      const tabId = await openRequest()
      expect(asksBeforeClosing(tabId)).toBe(false)
      tabs().editTabName(tabId, 'Create draft invoice')
      expect(builder().tabs[tabId]!.draft.name).toBe('Create draft invoice')
      expect(builder().tabs[tabId]!.dirty).toBe(false)
      expect(asksBeforeClosing(tabId)).toBe(true)
      tabs().editTabName(tabId, 'Create invoice  ')
      expect(asksBeforeClosing(tabId)).toBe(false)
      tabs().editTabName(tabId, 'Something else')
      tabs().cancelTabName(tabId)
      expect(builder().tabs[tabId]!.draft.name).toBe('Create invoice')
      expect(asksBeforeClosing(tabId)).toBe(false)
    })

    it('keeps the saved name everywhere outside the name editor until the save', async () => {
      const tabId = await openRequest()
      tabs().editTabName(tabId, 'Create draft invoice')
      expect(useSidebarStore.getState().tree[0]!.children[0]!.children[0]!.name).toBe('Create invoice')
      expect(treeService.renameItem).not.toHaveBeenCalled()
    })

    it('saves a name-only change with the rename call and nothing else', async () => {
      const tabId = await openRequest()
      tabs().editTabName(tabId, '  Create draft invoice ')
      await tabs().save()
      expect(treeService.renameItem).toHaveBeenCalledWith('ws', 'r1', 'Create draft invoice')
      expect(requests.saveRequest).not.toHaveBeenCalled()
      expect(builder().tabs[tabId]!.draft.name).toBe('Create draft invoice')
      expect(builder().tabs[tabId]!.baseRevision).toBe('rev-r1-1')
      expect(tabs().nameErrors[tabId]).toBeUndefined()
      expect(asksBeforeClosing(tabId)).toBe(false)
    })

    it('saves the other edits under the saved name first, then renames', async () => {
      const tabId = await openRequest()
      builder().setUrl(tabId, 'https://changed')
      tabs().editTabName(tabId, 'Create draft invoice')
      await tabs().save()
      expect(requests.saveRequest).toHaveBeenCalledWith('ws', expect.objectContaining({ name: 'Create invoice', url: 'https://changed' }), 'rev-r1-0', false)
      expect(treeService.renameItem).toHaveBeenCalledWith('ws', 'r1', 'Create draft invoice')
      expect(vi.mocked(requests.saveRequest).mock.invocationCallOrder[0]!).toBeLessThan(vi.mocked(treeService.renameItem).mock.invocationCallOrder[0]!)
      expect(asksBeforeClosing(tabId)).toBe(false)
    })

    it('stops the whole save on an empty name', async () => {
      const tabId = await openRequest()
      builder().setUrl(tabId, 'https://changed')
      tabs().editTabName(tabId, '   ')
      await tabs().save()
      expect(tabs().nameErrors[tabId]).toEqual({ kind: 'required', subject: 'request' })
      expect(requests.saveRequest).not.toHaveBeenCalled()
      expect(treeService.renameItem).not.toHaveBeenCalled()
      expect(builder().tabs[tabId]!.dirty).toBe(true)
      tabs().editTabName(tabId, 'Fixed')
      expect(tabs().nameErrors[tabId]).toBeUndefined()
    })

    it('does not rename when the other edits fail to save', async () => {
      const tabId = await openRequest()
      builder().setUrl(tabId, 'https://changed')
      tabs().editTabName(tabId, 'Create draft invoice')
      vi.mocked(requests.saveRequest).mockRejectedValue(new AppError('REQ-02409', 'changed on disk'))
      await tabs().save()
      expect(treeService.renameItem).not.toHaveBeenCalled()
      expect(builder().tabs[tabId]!.draft.name).toBe('Create draft invoice')
      expect(asksBeforeClosing(tabId)).toBe(true)
    })

    it('shows a failed rename under the name and keeps it pending', async () => {
      const tabId = await openRequest()
      builder().setUrl(tabId, 'https://changed')
      tabs().editTabName(tabId, 'Create draft invoice')
      const failure = new AppError('TRE-05500', 'The folder is read-only')
      vi.mocked(treeService.renameItem).mockRejectedValue(failure)
      await tabs().save()
      expect(requests.saveRequest).toHaveBeenCalledTimes(1)
      expect(tabs().nameErrors[tabId]).toEqual({ kind: 'failed', error: failure })
      expect(builder().tabs[tabId]!.dirty).toBe(false)
      expect(asksBeforeClosing(tabId)).toBe(true)
    })

    it('ignores a second save while one is running, and keeps a name typed meanwhile pending', async () => {
      const tabId = await openRequest()
      tabs().editTabName(tabId, 'First')
      const rename = deferred<TreeNode>()
      vi.mocked(treeService.renameItem).mockImplementation(async (_localId, id, name) => {
        await rename.promise
        names[id] = name
        return node(id, 'request', name)
      })
      const running = tabs().save()
      const again = tabs().save()
      tabs().editTabName(tabId, 'Second')
      rename.resolve(node('r1', 'request', 'First'))
      await Promise.all([running, again])
      expect(treeService.renameItem).toHaveBeenCalledTimes(1)
      expect(treeService.renameItem).toHaveBeenCalledWith('ws', 'r1', 'First')
      expect(builder().tabs[tabId]!.draft.name).toBe('Second')
      expect(asksBeforeClosing(tabId)).toBe(true)
    })

    it('keeps a pending name when the request is renamed elsewhere, and updates tabs without one', async () => {
      const mine = await openRequest()
      const sibling = tabs().duplicateTab(mine)!
      tabs().editTabName(mine, 'Mine')
      await useSidebarStore.getState().rename('r1', 'Theirs')
      await builder().adoptRename('r1', 'Create invoice')
      expect(builder().tabs[mine]!.draft.name).toBe('Mine')
      expect(builder().tabs[sibling]!.draft.name).toBe('Theirs')
      expect(builder().tabs[mine]!.baseRevision).toBe(builder().tabs[sibling]!.baseRevision)
      expect(asksBeforeClosing(mine)).toBe(true)
      expect(asksBeforeClosing(sibling)).toBe(false)
    })

    it('stops treating the draft as pending once the saved name catches up with it', async () => {
      const tabId = await openRequest()
      tabs().editTabName(tabId, 'Theirs')
      await useSidebarStore.getState().rename('r1', 'Theirs')
      await builder().adoptRename('r1', 'Create invoice')
      expect(asksBeforeClosing(tabId)).toBe(false)
    })

    it('keeps a name-only draft in the saved UI state', async () => {
      const tabId = await openRequest()
      tabs().editTabName(tabId, 'Kept')
      const saved = tabs().snapshot().tabs.find((tab) => tab.tabId === tabId)
      expect(saved?.draft?.name).toBe('Kept')
    })

    it('does not reload a tab whose only change is its name when the file changes on disk', async () => {
      const tabId = await openRequest()
      tabs().editTabName(tabId, 'Mine')
      await tabs().handleChangedOnDisk({ localId: 'ws', requestId: 'r1', deleted: false, revision: 'rev-disk' })
      expect(builder().tabs[tabId]!.draft.name).toBe('Mine')
      expect(tabs().notices[tabId]).toBe('changed')
    })

    it('treats the name of a never-saved request as part of its content', async () => {
      const tabId = tabs().activeTabId!
      expect(tabs().tabs.find((tab) => tab.id === tabId)?.requestId).toBeNull()
      tabs().editTabName(tabId, 'Ping')
      expect(builder().tabs[tabId]!.draft.name).toBe('Ping')
      expect(builder().tabs[tabId]!.dirty).toBe(true)
      await tabs().save()
      expect(tabs().saveDialog).toMatchObject({ tabId, mode: 'save' })
      expect(treeService.renameItem).not.toHaveBeenCalled()
    })
  })

  describe('collection and folder tabs', () => {
    async function openItem(itemId: 'c1' | 'f1') {
      tabs().openItemTab({ kind: itemId === 'c1' ? 'collection' : 'folder', itemId }, 'pinned')
      await useItemSettingsStore.getState().open(itemId)
      return tabs().activeTabId!
    }

    it.each(['c1', 'f1'] as const)('marks %s dirty for a pending name and clean again after Escape', async (itemId) => {
      const tabId = await openItem(itemId)
      tabs().editTabName(tabId, 'Renamed')
      expect(useItemSettingsStore.getState().items[itemId]!.draft!.name).toBe('Renamed')
      expect(asksBeforeClosing(tabId)).toBe(true)
      tabs().cancelTabName(tabId)
      expect(useItemSettingsStore.getState().items[itemId]!.draft!.name).toBe(names[itemId])
      expect(asksBeforeClosing(tabId)).toBe(false)
    })

    it('saves the settings first, then renames the collection', async () => {
      const tabId = await openItem('c1')
      useItemSettingsStore.getState().edit('c1', { description: 'edited' })
      tabs().editTabName(tabId, 'Billing v2')
      await tabs().save()
      expect(treeService.saveItemSettings).toHaveBeenCalledTimes(1)
      expect(treeService.renameItem).toHaveBeenCalledWith('ws', 'c1', 'Billing v2')
      expect(vi.mocked(treeService.saveItemSettings).mock.invocationCallOrder[0]!).toBeLessThan(
        vi.mocked(treeService.renameItem).mock.invocationCallOrder[0]!,
      )
      expect(asksBeforeClosing(tabId)).toBe(false)
    })

    it('renames a folder with only a name change', async () => {
      const tabId = await openItem('f1')
      tabs().editTabName(tabId, 'Receipts')
      await tabs().save()
      expect(treeService.renameItem).toHaveBeenCalledWith('ws', 'f1', 'Receipts')
      expect(treeService.saveItemSettings).not.toHaveBeenCalled()
      expect(asksBeforeClosing(tabId)).toBe(false)
    })

    it('does not rename while the collection variables wait for a confirmation', async () => {
      const tabId = await openItem('c1')
      const scope = collectionScope('c1')
      await environments().loadTable(scope)
      environments().editDraft(scope, (rows) => rows.map((row) => ({ ...row, secret: true })))
      tabs().editTabName(tabId, 'Billing v2')
      await tabs().save()
      expect(environments().drafts[scope]?.confirming).not.toBeNull()
      expect(treeService.renameItem).not.toHaveBeenCalled()
      expect(asksBeforeClosing(tabId)).toBe(true)
    })

    it('does not rename when the collection variables fail to save', async () => {
      const tabId = await openItem('c1')
      const scope = collectionScope('c1')
      await environments().loadTable(scope)
      environments().editDraft(scope, (rows) => rows.map((row) => ({ ...row, description: 'edited' })))
      vi.mocked(variables.saveVariableTable).mockRejectedValue(new AppError('VAR-05500', 'disk full'))
      tabs().editTabName(tabId, 'Billing v2')
      await tabs().save()
      expect(treeService.renameItem).not.toHaveBeenCalled()
      expect(useItemSettingsStore.getState().items.c1!.draft!.name).toBe('Billing v2')
    })

    it.each([
      ['c1', 'collection'],
      ['f1', 'folder'],
    ] as const)('stops the whole save of %s on an empty name', async (itemId, subject) => {
      const tabId = await openItem(itemId)
      useItemSettingsStore.getState().edit(itemId, { description: 'edited' })
      tabs().editTabName(tabId, '')
      await tabs().save()
      expect(tabs().nameErrors[tabId]).toEqual({ kind: 'required', subject })
      expect(treeService.saveItemSettings).not.toHaveBeenCalled()
      expect(treeService.renameItem).not.toHaveBeenCalled()
    })

    it('keeps a pending name when the item reloads from disk', async () => {
      const tabId = await openItem('c1')
      tabs().editTabName(tabId, 'Mine')
      await useItemSettingsStore.getState().load('c1')
      expect(useItemSettingsStore.getState().items.c1!.draft!.name).toBe('Mine')
      expect(asksBeforeClosing(tabId)).toBe(true)
    })
  })

  describe('closing a tab with a pending name', () => {
    it('asks, and Discard closes it without renaming', async () => {
      const tabId = await openRequest()
      tabs().editTabName(tabId, 'Create draft invoice')
      tabs().requestClose(tabId)
      expect(tabs().closePrompt).toBe(tabId)
      await tabs().resolveClose('discard')
      expect(isOpen(tabId)).toBe(false)
      expect(treeService.renameItem).not.toHaveBeenCalled()
      expect(tabs().nameErrors[tabId]).toBeUndefined()
      const reopened = await openRequest()
      expect(builder().tabs[reopened]!.draft.name).toBe('Create invoice')
    })

    it('keeps the draft on Cancel', async () => {
      const tabId = await openRequest()
      tabs().editTabName(tabId, 'Create draft invoice')
      tabs().requestClose(tabId)
      await tabs().resolveClose('cancel')
      expect(isOpen(tabId)).toBe(true)
      expect(builder().tabs[tabId]!.draft.name).toBe('Create draft invoice')
    })

    it('renames and closes on Save', async () => {
      const tabId = await openRequest()
      tabs().editTabName(tabId, 'Create draft invoice')
      tabs().requestClose(tabId)
      await tabs().resolveClose('save')
      expect(treeService.renameItem).toHaveBeenCalledWith('ws', 'r1', 'Create draft invoice')
      expect(isOpen(tabId)).toBe(false)
    })

    it('keeps the tab open with the error when Save meets an empty name', async () => {
      const tabId = await openRequest()
      tabs().editTabName(tabId, '')
      tabs().requestClose(tabId)
      await tabs().resolveClose('save')
      expect(isOpen(tabId)).toBe(true)
      expect(tabs().closePrompt).toBe(tabId)
      expect(tabs().closeError?.description).toBe('Type a name for the request.')
      expect(tabs().nameErrors[tabId]).toEqual({ kind: 'required', subject: 'request' })
      expect(treeService.renameItem).not.toHaveBeenCalled()
    })

    it('asks about a tab whose only change is its name when closing all tabs', async () => {
      const clean = await openRequest('r2')
      const renamed = await openRequest('r1')
      tabs().editTabName(renamed, 'Create draft invoice')
      tabs().closeAll()
      expect(tabs().closePrompt).toBe(renamed)
      expect(isOpen(clean)).toBe(false)
    })

    it.each([
      ['c1', 'collection'],
      ['f1', 'folder'],
    ] as const)('asks before closing a %s with a pending name and drops it on Discard', async (itemId, kind) => {
      tabs().openItemTab({ kind, itemId }, 'pinned')
      await useItemSettingsStore.getState().open(itemId)
      const tabId = tabs().activeTabId!
      tabs().editTabName(tabId, 'Renamed')
      tabs().requestClose(tabId)
      expect(tabs().closePrompt).toBe(tabId)
      await tabs().resolveClose('discard')
      expect(isOpen(tabId)).toBe(false)
      expect(treeService.renameItem).not.toHaveBeenCalled()
      expect(useItemSettingsStore.getState().items[itemId]).toBeUndefined()
    })

    it('renames a folder and closes it on Save', async () => {
      tabs().openItemTab({ kind: 'folder', itemId: 'f1' }, 'pinned')
      await useItemSettingsStore.getState().open('f1')
      const tabId = tabs().activeTabId!
      tabs().editTabName(tabId, 'Receipts')
      tabs().requestClose(tabId)
      await tabs().resolveClose('save')
      expect(treeService.renameItem).toHaveBeenCalledWith('ws', 'f1', 'Receipts')
      expect(isOpen(tabId)).toBe(false)
    })
  })

  describe('environment tabs', () => {
    function openEnvironment(environmentId = 'e1') {
      tabs().openEnvironmentTab({ kind: 'environment', environmentId }, 'pinned')
      return tabs().activeTabId!
    }

    it('marks the tab dirty for a pending name and clean again after Escape', async () => {
      const tabId = openEnvironment()
      tabs().editTabName(tabId, 'local')
      expect(environments().nameDrafts.e1).toBe('local')
      expect(asksBeforeClosing(tabId)).toBe(true)
      tabs().cancelTabName(tabId)
      expect(environments().nameDrafts.e1).toBeUndefined()
      expect(asksBeforeClosing(tabId)).toBe(false)
    })

    it('saves the variables first, then renames, then forgets the draft', async () => {
      const tabId = openEnvironment()
      await environments().loadTable('env:e1')
      environments().editDraft('env:e1', (rows) => rows.map((row) => ({ ...row, description: 'edited' })))
      tabs().editTabName(tabId, ' local ')
      await tabs().save()
      expect(variables.saveVariableTable).toHaveBeenCalledTimes(1)
      expect(environmentServices.updateEnvironment).toHaveBeenCalledWith('ws', expect.objectContaining({ id: 'e1', name: 'local' }))
      expect(vi.mocked(variables.saveVariableTable).mock.invocationCallOrder[0]!).toBeLessThan(
        vi.mocked(environmentServices.updateEnvironment).mock.invocationCallOrder[0]!,
      )
      expect(environments().nameDrafts.e1).toBeUndefined()
      expect(environments().environments.find((environment) => environment.id === 'e1')?.name).toBe('local')
      expect(asksBeforeClosing(tabId)).toBe(false)
    })

    it('stops the whole save when another environment has the name, ignoring case and spaces', async () => {
      const tabId = openEnvironment('e2')
      await environments().loadTable('env:e2')
      environments().editDraft('env:e2', (rows) => rows.map((row) => ({ ...row, description: 'edited' })))
      tabs().editTabName(tabId, '  DEV ')
      await tabs().save()
      expect(tabs().nameErrors[tabId]).toEqual({ kind: 'clash', name: '  DEV ' })
      expect(variables.saveVariableTable).not.toHaveBeenCalled()
      expect(environmentServices.updateEnvironment).not.toHaveBeenCalled()
      expect(asksBeforeClosing(tabId)).toBe(true)
    })

    it('lets an environment change only the case of its own name', async () => {
      const tabId = openEnvironment('e1')
      tabs().editTabName(tabId, 'DEV')
      await tabs().save()
      expect(tabs().nameErrors[tabId]).toBeUndefined()
      expect(environmentServices.updateEnvironment).toHaveBeenCalledWith('ws', expect.objectContaining({ id: 'e1', name: 'DEV' }))
    })

    it('stops the whole save on an empty name', async () => {
      const tabId = openEnvironment()
      tabs().editTabName(tabId, ' ')
      await tabs().save()
      expect(tabs().nameErrors[tabId]).toEqual({ kind: 'required', subject: 'environment' })
      expect(environmentServices.updateEnvironment).not.toHaveBeenCalled()
    })

    it('does not rename while the variables wait for a confirmation', async () => {
      const tabId = openEnvironment()
      await environments().loadTable('env:e1')
      environments().editDraft('env:e1', (rows) => rows.map((row) => ({ ...row, secret: true })))
      tabs().editTabName(tabId, 'local')
      await tabs().save()
      expect(environments().drafts['env:e1']?.confirming).not.toBeNull()
      expect(environmentServices.updateEnvironment).not.toHaveBeenCalled()
      expect(environments().nameDrafts.e1).toBe('local')
    })

    it('shows a failed rename under the name and keeps it pending', async () => {
      const tabId = openEnvironment()
      const failure = new AppError('ENV-05500', 'disk full')
      vi.mocked(environmentServices.updateEnvironment).mockRejectedValue(failure)
      tabs().editTabName(tabId, 'local')
      await tabs().save()
      expect(tabs().nameErrors[tabId]).toEqual({ kind: 'failed', error: failure })
      expect(environments().nameDrafts.e1).toBe('local')
    })

    it('drops the draft when the tab closes', async () => {
      const tabId = openEnvironment()
      tabs().editTabName(tabId, 'local')
      tabs().requestClose(tabId)
      await tabs().resolveClose('discard')
      expect(isOpen(tabId)).toBe(false)
      expect(environments().nameDrafts.e1).toBeUndefined()
      expect(environmentServices.updateEnvironment).not.toHaveBeenCalled()
    })
  })

  describe('example tabs', () => {
    async function openExample() {
      await tabs().openExamplePinned('r1', 'x1')
      return tabs().activeTabId!
    }

    it('marks the tab dirty for a pending name and clean again after Escape', async () => {
      const tabId = await openExample()
      tabs().editTabName(tabId, 'Created OK')
      expect(builder().tabs[tabId]!.draft.name).toBe('Created OK')
      expect(asksBeforeClosing(tabId)).toBe(true)
      tabs().cancelTabName(tabId)
      expect(builder().tabs[tabId]!.draft.name).toBe('Created')
      expect(asksBeforeClosing(tabId)).toBe(false)
    })

    it('saves the example under its saved name first, then renames it', async () => {
      const tabId = await openExample()
      useExamplesStore.getState().setCode(tabId, 202)
      tabs().editTabName(tabId, 'Created OK')
      await tabs().save()
      expect(exampleServices.saveExample).toHaveBeenCalledWith('ws', 'r1', expect.objectContaining({ name: 'Created' }), 'rev-x1', false)
      expect(exampleServices.renameExample).toHaveBeenCalledWith('ws', 'r1', 'x1', 'Created OK')
      expect(vi.mocked(exampleServices.saveExample).mock.invocationCallOrder[0]!).toBeLessThan(
        vi.mocked(exampleServices.renameExample).mock.invocationCallOrder[0]!,
      )
      expect(asksBeforeClosing(tabId)).toBe(false)
    })

    it('does not rename when the example fails to save', async () => {
      const tabId = await openExample()
      useExamplesStore.getState().setCode(tabId, 99)
      vi.mocked(exampleServices.saveExample).mockRejectedValue(new AppError('EXM-03400', 'The status code must be between 100 and 599'))
      tabs().editTabName(tabId, 'Created OK')
      await tabs().save()
      expect(exampleServices.renameExample).not.toHaveBeenCalled()
      expect(builder().tabs[tabId]!.draft.name).toBe('Created OK')
    })

    it('stops the whole save on an empty name', async () => {
      const tabId = await openExample()
      useExamplesStore.getState().setCode(tabId, 202)
      tabs().editTabName(tabId, '')
      await tabs().save()
      expect(tabs().nameErrors[tabId]).toEqual({ kind: 'required', subject: 'example' })
      expect(exampleServices.saveExample).not.toHaveBeenCalled()
      expect(exampleServices.renameExample).not.toHaveBeenCalled()
    })

    it('drops the draft on Discard without renaming', async () => {
      const tabId = await openExample()
      tabs().editTabName(tabId, 'Created OK')
      tabs().requestClose(tabId)
      expect(tabs().closePrompt).toBe(tabId)
      await tabs().resolveClose('discard')
      expect(isOpen(tabId)).toBe(false)
      expect(exampleServices.renameExample).not.toHaveBeenCalled()
    })
  })
})
