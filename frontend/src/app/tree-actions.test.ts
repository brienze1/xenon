import { beforeEach, describe, expect, it, vi } from 'vitest'
import { emptyAuth } from '@/platform/mappers/request'
import type { TreeNode } from '@/platform/mappers/tree'

vi.mock('@/platform/services/tree', () => ({ getTree: vi.fn(), searchTree: vi.fn(), getItemSettings: vi.fn(), saveItemSettings: vi.fn() }))
vi.mock('@/platform/services/ui-state', () => ({ getUIState: vi.fn(), saveUIState: vi.fn(async () => undefined) }))
vi.mock('@/platform/services/workspaces', () => ({ listWorkspaces: vi.fn(), createWorkspace: vi.fn(), activateWorkspace: vi.fn() }))
vi.mock('@/platform/services/runner', async (importOriginal) => ({
  ...(await importOriginal<typeof import('@/platform/services/runner')>()),
  getRun: vi.fn(),
}))

const tree = await import('@/platform/services/tree')
const runner = await import('@/platform/services/runner')
const { useWorkspacesStore } = await import('@/features/workspaces/store')
const { useTabsStore } = await import('@/features/tabs/store')
const { useSidebarStore } = await import('@/features/sidebar/store')
const { useItemSettingsStore } = await import('@/features/item-settings/store')
const { useRunnerStore } = await import('@/features/runner/store')
const { run, runSummaryView } = await import('@/features/runner/test-fixtures')
const { openRunResults, runItemTab } = await import('./tree-actions')

function node(id: string, kind: TreeNode['kind'], children: TreeNode[] = []): TreeNode {
  return { id, kind, name: id, method: null, url: null, error: null, warning: null, children }
}

describe('Run… and Open results', () => {
  beforeEach(() => {
    vi.resetAllMocks()
    useTabsStore.getState().reset()
    useItemSettingsStore.getState().reset()
    useRunnerStore.getState().reset()
    useWorkspacesStore.setState({ activeLocalId: 'L1' })
    useSidebarStore.setState({ tree: [node('c1', 'collection', [node('f1', 'folder')])] })
    vi.mocked(tree.getItemSettings).mockImplementation(async (_localId, id) => ({
      id,
      kind: id === 'c1' ? 'collection' : 'folder',
      name: id,
      description: '',
      auth: emptyAuth(),
      headers: [],
      scripts: { preRequest: '', postResponse: '', disabled: [] },
      inheritScripts: null,
    }))
  })

  it('Run… opens the tab pinned on the Runner, on Setup', () => {
    runItemTab('f1', 'folder')
    const tab = useTabsStore.getState().tabs.find((candidate) => candidate.itemId === 'f1')
    expect(tab).toMatchObject({ kind: 'folder', pinned: true })
    expect(useTabsStore.getState().activeTabId).toBe(tab?.id)
    expect(useItemSettingsStore.getState().items.f1?.section).toBe('runner')
    expect(useRunnerStore.getState().targets.f1?.view).toBe('setup')
  })

  it('Run… shows Results while a run is active on the target', () => {
    useRunnerStore.setState({ active: { c1: runSummaryView() } })
    runItemTab('c1', 'collection')
    expect(useRunnerStore.getState().targets.c1).toMatchObject({ view: 'results', runId: 'R1' })
  })

  it('Open results focuses the target tab on Runner › Results for that run, finding the kind in the tree', async () => {
    vi.mocked(runner.getRun).mockResolvedValue(run({ id: 'R7' }))
    openRunResults({ itemId: 'f1', runId: 'R7', kind: null, targetName: 'f1', status: 'completed', passed: 1, failed: 0 })
    expect(useTabsStore.getState().tabs.find((candidate) => candidate.itemId === 'f1')).toMatchObject({ kind: 'folder', pinned: true })
    expect(useItemSettingsStore.getState().items.f1?.section).toBe('runner')
    await vi.waitFor(() => expect(useRunnerStore.getState().targets.f1).toMatchObject({ view: 'results', runId: 'R7', run: { id: 'R7' } }))
    openRunResults({ itemId: 'gone', runId: 'R8', kind: null, targetName: 'x', status: 'completed', passed: 0, failed: 0 })
    expect(useTabsStore.getState().tabs.some((candidate) => candidate.itemId === 'gone')).toBe(false)
  })
})
