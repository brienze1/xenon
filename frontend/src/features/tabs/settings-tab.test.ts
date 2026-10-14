import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest'
import { AppError } from '@/platform/errors'
import { defaultAppSettings } from '@/platform/mappers/settings'
import { defaultUIState } from '@/platform/mappers/ui-state'

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
vi.mock('@/platform/services/environments', () => ({ listEnvironments: vi.fn(), setActiveEnvironment: vi.fn() }))
vi.mock('@/platform/services/tree', () => ({ getTree: vi.fn(), searchTree: vi.fn() }))
vi.mock('@/platform/services/history', () => ({ listHistory: vi.fn() }))
vi.mock('@/platform/services/settings', () => ({
  SCRIPT_TIMEOUT_OUT_OF_RANGE: 'SCR-01400',
  HISTORY_USER_CAP_OUT_OF_RANGE: 'SET-02400',
  HISTORY_AI_CAP_OUT_OF_RANGE: 'SET-03400',
  AI_PORT_OUT_OF_RANGE: 'SET-04400',
  getSettings: vi.fn(),
  saveSettings: vi.fn(),
}))
vi.mock('@/platform/services/ai-access', () => ({ getAIAccessStatus: vi.fn(), copyText: vi.fn() }))

const requests = await import('@/platform/services/requests')
const uiState = await import('@/platform/services/ui-state')
const settingsService = await import('@/platform/services/settings')
const { useWorkspacesStore } = await import('@/features/workspaces/store')
const { useRequestBuilderStore } = await import('@/features/request-builder/store')
const { useSettingsStore } = await import('@/features/settings/store')
const { useTabsStore } = await import('./store')

function tabs() {
  return useTabsStore.getState()
}

function settings() {
  return useSettingsStore.getState()
}

function settingsTabs() {
  return tabs().tabs.filter((tab) => tab.kind === 'settings')
}

describe('settings tab', () => {
  beforeEach(() => {
    vi.resetAllMocks()
    tabs().reset()
    useRequestBuilderStore.getState().reset()
    useSettingsStore.setState({ settings: defaultAppSettings(), fieldErrors: {}, saveError: null, saving: false, pageOpen: false })
    settings().discard()
    vi.mocked(requests.autoHeaders).mockResolvedValue([])
    vi.mocked(requests.releaseResponse).mockResolvedValue()
    vi.mocked(uiState.saveUIState).mockResolvedValue()
    useWorkspacesStore.setState({ activeLocalId: 'ws' })
    tabs().hydrate('ws', { ...defaultUIState(), tabs: [{ tabId: 'start', requestId: null, pinned: true, baseRevision: null, draft: null }], activeTabId: 'start' })
  })

  afterEach(() => {
    useWorkspacesStore.setState({ activeLocalId: null })
  })

  it('opens one pinned Settings tab and focuses it on a second open', () => {
    tabs().openSettingsTab()
    const [opened] = settingsTabs()
    expect(opened).toMatchObject({ kind: 'settings', pinned: true, requestId: null })
    expect(tabs().activeTabId).toBe(opened!.id)
    expect(useRequestBuilderStore.getState().tabs[opened!.id]).toBeUndefined()
    tabs().activate('start')
    tabs().openSettingsTab()
    expect(settingsTabs()).toHaveLength(1)
    expect(tabs().activeTabId).toBe(opened!.id)
  })

  it('opens the full-page Settings instead when no workspace is open', () => {
    tabs().reset()
    tabs().openSettingsTab()
    expect(tabs().tabs).toEqual([])
    expect(settings().pageOpen).toBe(true)
  })

  it('is persisted in the UI state and restored without its unsaved edits', async () => {
    tabs().openSettingsTab()
    const tabId = tabs().activeTabId!
    settings().edit({ aiPort: '50000' })
    const snapshot = tabs().snapshot()
    expect(snapshot.tabs.find((tab) => tab.tabId === tabId)).toEqual({ tabId, kind: 'settings', requestId: null, pinned: true, baseRevision: null, draft: null })
    await tabs().flushPersist()
    expect(uiState.saveUIState).toHaveBeenCalledWith('ws', expect.objectContaining({ activeTabId: tabId }))
    tabs().reset()
    settings().discard()
    tabs().hydrate('ws', snapshot)
    expect(tabs().tabs.find((tab) => tab.id === tabId)).toMatchObject({ kind: 'settings', pinned: true })
    expect(tabs().activeTabId).toBe(tabId)
    expect(useRequestBuilderStore.getState().tabs[tabId]).toBeUndefined()
    expect(settings().draft.aiPort).toBe('47821')
  })

  it('asks before closing with unsaved edits: Cancel keeps them, Don’t Save drops them', async () => {
    tabs().openSettingsTab()
    const tabId = tabs().activeTabId!
    tabs().requestClose(tabId)
    expect(tabs().closePrompt).toBeNull()
    expect(settingsTabs()).toHaveLength(0)
    tabs().openSettingsTab()
    const reopened = tabs().activeTabId!
    settings().edit({ historyAiCap: '500' })
    tabs().requestClose(reopened)
    expect(tabs().closePrompt).toBe(reopened)
    await tabs().resolveClose('cancel')
    expect(settingsTabs()).toHaveLength(1)
    expect(settings().draft.historyAiCap).toBe('500')
    tabs().requestClose(reopened)
    await tabs().resolveClose('discard')
    expect(settingsTabs()).toHaveLength(0)
    expect(settings().draft.historyAiCap).toBe('1000')
  })

  it('saves from the close prompt, or keeps the tab open on an invalid field or a failed save', async () => {
    tabs().openSettingsTab()
    const tabId = tabs().activeTabId!
    settings().edit({ aiPort: '80' })
    tabs().requestClose(tabId)
    await tabs().resolveClose('save')
    expect(settingsService.saveSettings).not.toHaveBeenCalled()
    expect(tabs()).toMatchObject({ closePrompt: null, activeTabId: tabId })
    expect(settings().fieldErrors.aiPort).toBeDefined()
    settings().edit({ aiPort: '50000' })
    vi.mocked(settingsService.saveSettings).mockRejectedValueOnce(new AppError('SCR-03500', 'app.json could not be written'))
    tabs().requestClose(tabId)
    await tabs().resolveClose('save')
    expect(tabs()).toMatchObject({ closePrompt: tabId, closeError: { code: 'SCR-03500' } })
    vi.mocked(settingsService.saveSettings).mockImplementation(async (saved) => saved)
    await tabs().resolveClose('save')
    expect(settingsService.saveSettings).toHaveBeenLastCalledWith(expect.objectContaining({ aiAccess: expect.objectContaining({ port: 50000 }) }))
    expect(settingsTabs()).toHaveLength(0)
    expect(tabs().closePrompt).toBeNull()
  })

  it('saves with Cmd/Ctrl+S, never offers Save As, and reopens as the single Settings tab', async () => {
    tabs().openSettingsTab()
    const tabId = tabs().activeTabId!
    vi.mocked(settingsService.saveSettings).mockImplementation(async (saved) => saved)
    settings().edit({ scriptTimeoutSeconds: '30' })
    await tabs().save()
    expect(settingsService.saveSettings).toHaveBeenCalledWith(expect.objectContaining({ scriptTimeoutMs: 30000 }))
    tabs().saveAs()
    expect(tabs().saveDialog).toBeNull()
    tabs().closeTab(tabId)
    await tabs().reopenClosed()
    expect(settingsTabs()).toHaveLength(1)
    expect(tabs().activeTabId).toBe(settingsTabs()[0]!.id)
  })
})
