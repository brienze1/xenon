import { act, render, screen, within } from '@testing-library/react'
import { beforeEach, describe, expect, it, vi } from 'vitest'
import type { RunSummaryView } from '@/platform/mappers/runner'
import type { TreeNode } from '@/platform/mappers/tree'

vi.mock('@/platform/services/ui-state', () => ({ getUIState: vi.fn(), saveUIState: vi.fn(async () => undefined) }))
vi.mock('@/platform/services/workspaces', () => ({ listWorkspaces: vi.fn(), createWorkspace: vi.fn(), activateWorkspace: vi.fn() }))
vi.mock('@/platform/services/tree', () => ({ getTree: vi.fn(async () => []), searchTree: vi.fn() }))

const { useTabsStore } = await import('@/features/tabs/store')
const { useSidebarStore } = await import('@/features/sidebar/store')
const { useRunnerStore } = await import('@/features/runner/store')
const { TabStrip } = await import('./TabStrip')

function node(id: string, kind: TreeNode['kind'], name: string, children: TreeNode[] = []): TreeNode {
  return { id, kind, name, method: null, url: null, error: null, warning: null, children }
}

const progress = { currentRow: 1, currentName: 'Login', iteration: 0, completed: 1, planned: 2, testsPassed: 0, testsFailed: 0, elapsedMs: 0, metrics: null }
const summary = { iterations: 0, totalRequests: 0, testsPassed: 0, testsFailed: 0, testsSkipped: 0, avgResponseMs: 0, durationMs: 0, passed: 0, failed: 0, skipped: 0, errors: 0, cancelled: 0 }
const folderRun: RunSummaryView = {
  id: 'R1',
  itemId: 'f1',
  targetKind: 'folder',
  targetName: 'Auth',
  type: 'functional',
  status: 'running',
  environmentName: '',
  startedAt: '',
  finishedAt: '',
  progress,
  summary,
}

describe('TabItem running indicator', () => {
  beforeEach(() => {
    useRunnerStore.getState().reset()
    useSidebarStore.setState({ tree: [node('c1', 'collection', 'Orders API', [node('f1', 'folder', 'Auth')])], status: 'ready' })
    useTabsStore.setState({
      tabs: [
        { id: 'tc', requestId: null, pinned: true, kind: 'collection', itemId: 'c1' },
        { id: 'tf', requestId: null, pinned: true, kind: 'folder', itemId: 'f1' },
      ],
      activeTabId: 'tc',
      previewTabId: null,
    })
  })

  it('pulses on the tab of a collection or folder with an active run and names it for assistive tech', () => {
    render(<TabStrip />)
    const strip = screen.getByRole('tablist', { name: 'Open requests' })
    expect(within(strip).queryByRole('tab', { name: /run in progress/ })).not.toBeInTheDocument()
    act(() => useRunnerStore.setState({ active: { f1: folderRun } }))
    const folder = within(strip).getByRole('tab', { name: 'Auth, run in progress' })
    expect(folder.querySelector('[class*="running"]')).not.toBeNull()
    act(() => useRunnerStore.setState({ active: {} }))
    expect(within(strip).queryByRole('tab', { name: /run in progress/ })).not.toBeInTheDocument()
  })
})
