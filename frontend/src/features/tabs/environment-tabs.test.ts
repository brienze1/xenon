import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest'
import { AppError } from '@/platform/errors'
import { emptyDraft } from '@/platform/mappers/request'
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
}))
vi.mock('@/platform/services/ui-state', () => ({ getUIState: vi.fn(), saveUIState: vi.fn() }))
vi.mock('@/platform/services/environments', () => ({
  listEnvironments: vi.fn(),
  setActiveEnvironment: vi.fn(),
}))
vi.mock('@/platform/services/variables', () => ({
  CONFIRM_REVEAL_REQUIRED: 'VAR-01409',
  getVariableTable: vi.fn(),
  saveVariableTable: vi.fn(),
}))
vi.mock('@/platform/services/vault', () => ({ PASSPHRASE_SETUP_REQUIRED: 'VLT-05409', VAULT_LOCKED: 'VLT-04409' }))
vi.mock('@/platform/services/tree', () => ({ getTree: vi.fn(), searchTree: vi.fn() }))
vi.mock('@/platform/services/history', () => ({ listHistory: vi.fn() }))

const requests = await import('@/platform/services/requests')
const variables = await import('@/platform/services/variables')
const { useWorkspacesStore } = await import('@/features/workspaces/store')
const { useRequestBuilderStore } = await import('@/features/request-builder/store')
const { useEnvironmentsStore } = await import('@/features/environments/store')
const { useSidebarStore } = await import('@/features/sidebar/store')
const { useResponseViewerStore } = await import('@/features/response-viewer/store')
const { useTabsStore } = await import('./store')

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

const dev = { id: 'e1', name: 'dev', color: 'green', confirmBeforeSend: false, error: null } as const
const hml = { id: 'e2', name: 'hml', color: 'yellow', confirmBeforeSend: false, error: null } as const

function tabs() {
  return useTabsStore.getState()
}

function environments() {
  return useEnvironmentsStore.getState()
}

async function editVariables(scope: string) {
  await environments().loadTable(scope)
  environments().editDraft(scope, (rows) => rows.map((row) => ({ ...row, description: 'edited' })))
}

describe('environment tabs', () => {
  beforeEach(() => {
    vi.resetAllMocks()
    tabs().reset()
    useRequestBuilderStore.getState().reset()
    useResponseViewerStore.getState().reset()
    useSidebarStore.getState().reset()
    environments().reset()
    vi.mocked(requests.autoHeaders).mockResolvedValue([])
    vi.mocked(requests.releaseResponse).mockResolvedValue()
    vi.mocked(requests.getRequest).mockImplementation(async (_localId, id) => ({ draft: { ...emptyDraft(), id, name: `Request ${id}` }, revision: `rev-${id}`, aliases: [] }))
    vi.mocked(variables.getVariableTable).mockImplementation(async (_localId, scope) => ({ scope, rows: [ROW], warnings: [] }))
    useWorkspacesStore.setState({ activeLocalId: 'ws' })
    useEnvironmentsStore.setState({ environments: [dev, hml], activeId: 'e1', status: 'ready' })
    tabs().hydrate('ws', { ...defaultUIState(), tabs: [{ tabId: 'start', requestId: 'r0', pinned: true, baseRevision: null, draft: null }], activeTabId: 'start' })
  })

  afterEach(() => {
    environments().reset()
    useWorkspacesStore.setState({ activeLocalId: null })
  })

  it('previews an environment on a single click and pins it on a double click', () => {
    tabs().openEnvironmentTab({ kind: 'environment', environmentId: 'e1' }, 'preview')
    const preview = tabs().previewTabId
    expect(tabs().tabs.find((tab) => tab.id === preview)).toMatchObject({ kind: 'environment', environmentId: 'e1', pinned: false, requestId: null })
    expect(tabs().activeTabId).toBe(preview)
    expect(useRequestBuilderStore.getState().tabs[preview!]).toBeUndefined()
    tabs().openEnvironmentTab({ kind: 'environment', environmentId: 'e2' }, 'preview')
    expect(tabs().tabs).toHaveLength(2)
    expect(tabs().tabs[1]).toMatchObject({ id: preview, environmentId: 'e2' })
    tabs().openEnvironmentTab({ kind: 'environment', environmentId: 'e2' }, 'pinned')
    expect(tabs().previewTabId).toBeNull()
    expect(tabs().tabs[1]?.pinned).toBe(true)
  })

  it('turns a clean request preview into an environment preview and back', async () => {
    await tabs().openPreview('r1')
    const preview = tabs().previewTabId!
    tabs().openEnvironmentTab({ kind: 'globals' }, 'preview')
    expect(tabs().tabs.find((tab) => tab.id === preview)).toMatchObject({ kind: 'globals', requestId: null })
    expect(useRequestBuilderStore.getState().tabs[preview]).toBeUndefined()
    await tabs().openPreview('r2')
    expect(tabs().tabs.find((tab) => tab.id === preview)).toMatchObject({ requestId: 'r2', kind: undefined, environmentId: undefined })
  })

  it('focuses the tab that already shows an environment or the globals', () => {
    tabs().openEnvironmentTab({ kind: 'globals' }, 'pinned')
    const globals = tabs().activeTabId
    tabs().activate('start')
    tabs().openEnvironmentTab({ kind: 'globals' }, 'preview')
    expect(tabs().activeTabId).toBe(globals)
    expect(tabs().tabs).toHaveLength(2)
  })

  it('pins a previewed environment once its variables are edited, and asks before closing it', async () => {
    tabs().openEnvironmentTab({ kind: 'environment', environmentId: 'e1' }, 'preview')
    const tabId = tabs().activeTabId!
    await editVariables('env:e1')
    tabs().pinIfEdited()
    expect(tabs().previewTabId).toBeNull()
    tabs().requestClose(tabId)
    expect(tabs().closePrompt).toBe(tabId)
    await tabs().resolveClose('cancel')
    expect(tabs().tabs.some((tab) => tab.id === tabId)).toBe(true)
    tabs().requestClose(tabId)
    await tabs().resolveClose('discard')
    expect(tabs().tabs.some((tab) => tab.id === tabId)).toBe(false)
    expect(environments().drafts['env:e1']?.dirty).toBe(false)
  })

  it('saves the variables from the close prompt and closes the tab', async () => {
    tabs().openEnvironmentTab({ kind: 'globals' }, 'pinned')
    const tabId = tabs().activeTabId!
    await editVariables('globals')
    vi.mocked(variables.saveVariableTable).mockResolvedValue({ scope: 'globals', rows: [{ ...ROW, description: 'edited' }], warnings: [] })
    tabs().requestClose(tabId)
    await tabs().resolveClose('save')
    expect(variables.saveVariableTable).toHaveBeenCalledWith('ws', 'globals', [expect.objectContaining({ key: 'base_url', description: 'edited' })], false)
    expect(tabs().closePrompt).toBeNull()
    expect(tabs().tabs.some((tab) => tab.id === tabId)).toBe(false)
  })

  it('keeps the tab and shows the error when saving from the close prompt fails', async () => {
    tabs().openEnvironmentTab({ kind: 'environment', environmentId: 'e1' }, 'pinned')
    const tabId = tabs().activeTabId!
    await editVariables('env:e1')
    vi.mocked(variables.saveVariableTable).mockRejectedValue(new AppError('VAR-01400', 'bad name'))
    tabs().requestClose(tabId)
    await tabs().resolveClose('save')
    expect(tabs().closePrompt).toBe(tabId)
    expect(tabs().closeError?.code).toBe('VAR-01400')
  })

  it('leaves the tab open to confirm a type switch when saving from the close prompt', async () => {
    tabs().openEnvironmentTab({ kind: 'environment', environmentId: 'e1' }, 'pinned')
    const tabId = tabs().activeTabId!
    await environments().loadTable('env:e1')
    environments().editDraft('env:e1', (rows) => rows.map((row) => ({ ...row, secret: true })))
    tabs().activate('start')
    tabs().requestClose(tabId)
    await tabs().resolveClose('save')
    expect(tabs().closePrompt).toBeNull()
    expect(tabs().activeTabId).toBe(tabId)
    expect(environments().drafts['env:e1']?.confirming).not.toBeNull()
  })

  it('saves the variables of the active environment tab with Save', async () => {
    tabs().openEnvironmentTab({ kind: 'environment', environmentId: 'e1' }, 'pinned')
    await editVariables('env:e1')
    vi.mocked(variables.saveVariableTable).mockResolvedValue({ scope: 'env:e1', rows: [ROW], warnings: [] })
    await tabs().save()
    expect(variables.saveVariableTable).toHaveBeenCalledWith('ws', 'env:e1', expect.any(Array), false)
    tabs().saveAs()
    expect(tabs().saveDialog).toBeNull()
    expect(tabs().duplicateTab(tabs().activeTabId!)).toBeNull()
  })

  it('closes the tabs of environments that were deleted', () => {
    tabs().openEnvironmentTab({ kind: 'environment', environmentId: 'e1' }, 'pinned')
    tabs().openEnvironmentTab({ kind: 'environment', environmentId: 'e2' }, 'pinned')
    tabs().openEnvironmentTab({ kind: 'globals' }, 'pinned')
    tabs().handleEnvironmentsChanged(['e2'])
    expect(tabs().tabs.map((tab) => tab.environmentId ?? tab.kind ?? 'request')).toEqual(['request', 'e2', 'globals'])
    expect(tabs().closedTabs).toHaveLength(0)
  })

  it('persists environment tabs and restores them', () => {
    tabs().openEnvironmentTab({ kind: 'environment', environmentId: 'e1' }, 'preview')
    tabs().openEnvironmentTab({ kind: 'globals' }, 'pinned')
    const snapshot = tabs().snapshot()
    expect(snapshot.tabs.slice(1)).toEqual([
      { tabId: expect.any(String), kind: 'environment', requestId: null, environmentId: 'e1', pinned: false, baseRevision: null, draft: null },
      { tabId: expect.any(String), kind: 'globals', requestId: null, environmentId: null, pinned: true, baseRevision: null, draft: null },
    ])
    tabs().reset()
    tabs().hydrate('ws', { ...snapshot, previewTabId: snapshot.previewTabId })
    expect(tabs().tabs.map((tab) => tab.kind ?? 'request')).toEqual(['request', 'environment', 'globals'])
    expect(tabs().previewTabId).toBe(snapshot.previewTabId)
    expect(Object.keys(useRequestBuilderStore.getState().tabs)).toEqual(['start'])
  })

  it('reopens a closed environment tab unless the environment is gone', () => {
    tabs().openEnvironmentTab({ kind: 'environment', environmentId: 'e1' }, 'pinned')
    tabs().requestClose()
    expect(tabs().closedTabs.at(-1)?.tab).toMatchObject({ kind: 'environment', environmentId: 'e1' })
    void tabs().reopenClosed()
    expect(tabs().tabs.at(-1)).toMatchObject({ kind: 'environment', environmentId: 'e1', pinned: true })
    tabs().requestClose()
    useEnvironmentsStore.setState({ environments: [hml] })
    void tabs().reopenClosed()
    expect(tabs().tabs.some((tab) => tab.kind === 'environment')).toBe(false)
  })

  it('reveals an environment tab in the sidebar Environments view', () => {
    tabs().openEnvironmentTab({ kind: 'environment', environmentId: 'e1' }, 'pinned')
    tabs().revealInSidebar(tabs().activeTabId!)
    expect(useSidebarStore.getState().view).toBe('environments')
  })
})
