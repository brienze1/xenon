import { act, render, screen } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest'
import { defaultAppSettings } from '@/platform/mappers/settings'
import { idleUpdateStatus, type UpdateState } from '@/platform/mappers/update-status'
import { defaultUIState } from '@/platform/mappers/ui-state'
import { updateStatus } from '@/features/updates/test-fixtures'

vi.mock('@/platform/services/updates', () => ({
  getUpdateStatus: vi.fn(),
  checkForUpdates: vi.fn(),
  checkOnStartup: vi.fn(),
  startUpdate: vi.fn(),
  restartToUpdate: vi.fn(),
  openReleasePage: vi.fn(),
}))
vi.mock('@/platform/services/ui-state', () => ({ getUIState: vi.fn(), saveUIState: vi.fn() }))
vi.mock('@/platform/services/requests', () => ({ autoHeaders: vi.fn(), releaseResponse: vi.fn() }))

const service = await import('@/platform/services/updates')
const uiState = await import('@/platform/services/ui-state')
const { useUpdatesStore } = await import('@/features/updates/store')
const { useTabsStore } = await import('@/features/tabs/store')
const { useSettingsStore } = await import('@/features/settings/store')
const { FooterUpdateNotice } = await import('./FooterUpdateNotice')

function showStatus(state: UpdateState) {
  act(() => useUpdatesStore.setState({ status: updateStatus({ state }) }))
}

function openWorkspace() {
  useTabsStore.getState().hydrate('ws', { ...defaultUIState(), tabs: [{ tabId: 'start', kind: 'globals', requestId: null, pinned: true, baseRevision: null, draft: null }], activeTabId: 'start' })
}

function settingsTabs() {
  return useTabsStore.getState().tabs.filter((tab) => tab.kind === 'settings')
}

describe('FooterUpdateNotice', () => {
  beforeEach(() => {
    vi.resetAllMocks()
    vi.mocked(uiState.saveUIState).mockResolvedValue()
    useUpdatesStore.setState({ status: idleUpdateStatus(), toast: 'unseen' })
  })

  afterEach(() => {
    act(() => {
      useTabsStore.getState().reset()
      useSettingsStore.setState({ settings: defaultAppSettings(), pageOpen: false })
      useSettingsStore.getState().discard()
    })
  })

  it.each([false, true])('restarts into a ready update (workspace open: %s), saving the tabs first', async (withWorkspace) => {
    if (withWorkspace) {
      openWorkspace()
    }
    vi.mocked(service.restartToUpdate).mockResolvedValue(updateStatus({ state: 'installing' }))
    render(<FooterUpdateNotice />)
    showStatus('ready')
    await userEvent.click(screen.getByRole('button', { name: 'Update ready — Restart' }))
    expect(service.restartToUpdate).toHaveBeenCalledTimes(1)
    if (withWorkspace) {
      expect(uiState.saveUIState).toHaveBeenCalledBefore(vi.mocked(service.restartToUpdate))
    }
  })

  it.each([true, false])('shows "Update available" with Update automatically %s and opens the Settings tab', async (autoUpdate) => {
    act(() => useSettingsStore.setState({ settings: { ...defaultAppSettings(), autoUpdate } }))
    openWorkspace()
    render(<FooterUpdateNotice />)
    showStatus('available')
    await userEvent.click(screen.getByRole('button', { name: 'Update available' }))
    expect(settingsTabs()).toHaveLength(1)
    expect(useTabsStore.getState().activeTabId).toBe(settingsTabs()[0]!.id)
  })

  it('opens the full-page Settings from "Update available" with no workspace open', async () => {
    render(<FooterUpdateNotice />)
    showStatus('available')
    await userEvent.click(screen.getByRole('button', { name: 'Update available' }))
    expect(useSettingsStore.getState().pageOpen).toBe(true)
  })

  it.each<UpdateState>(['idle', 'checking', 'up_to_date', 'no_package', 'check_failed', 'downloading', 'installing', 'failed'])('shows nothing while %s', (state) => {
    const { container } = render(<FooterUpdateNotice />)
    showStatus(state)
    expect(container).toBeEmptyDOMElement()
  })
})
