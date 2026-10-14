import { act, render, screen, within } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest'
import { AppError } from '@/platform/errors'
import type { AppSettings } from '@/platform/mappers/settings'
import { defaultAppSettings } from '@/platform/mappers/settings'
import { defaultUIState } from '@/platform/mappers/ui-state'

vi.mock('@/platform/services/settings', () => ({
  SCRIPT_TIMEOUT_OUT_OF_RANGE: 'SCR-01400',
  HISTORY_USER_CAP_OUT_OF_RANGE: 'SET-02400',
  HISTORY_AI_CAP_OUT_OF_RANGE: 'SET-03400',
  AI_PORT_OUT_OF_RANGE: 'SET-04400',
  getSettings: vi.fn(),
  saveSettings: vi.fn(),
}))
vi.mock('@/platform/services/ai-access', () => ({ getAIAccessStatus: vi.fn(), copyText: vi.fn() }))
vi.mock('@/platform/services/git-accounts', () => ({
  GITHUB_SIGN_IN_NOT_CONFIGURED: 'ACC-01503',
  listGitAccounts: vi.fn(),
  removeGitAccount: vi.fn(),
}))
vi.mock('@/platform/services/ui-state', () => ({ getUIState: vi.fn(), saveUIState: vi.fn() }))
vi.mock('@/platform/services/requests', () => ({ autoHeaders: vi.fn(), releaseResponse: vi.fn() }))

const service = await import('@/platform/services/settings')
const aiAccess = await import('@/platform/services/ai-access')
const gitAccounts = await import('@/platform/services/git-accounts')
const { isSettingsDirty, useGitAccountsStore, useSettingsStore } = await import('@/features/settings/store')
const { useTabsStore } = await import('@/features/tabs/store')
const { SettingsTab } = await import('./SettingsTab')
const { SettingsButton } = await import('./SettingsButton')

const saved: AppSettings = { scriptTimeoutMs: 3000, historyUserCap: 5000, historyAiCap: 200, aiAccess: { enabled: true, port: 47821, skipApprovals: false } }

const SHORTCUTS = [{ label: 'Search requests', combos: ['Mod+K'] }]

async function renderTab(props: Parameters<typeof SettingsTab>[0] = {}) {
  render(<SettingsTab shortcuts={SHORTCUTS} {...props} />)
  await screen.findByDisplayValue('3')
}

function field(name: RegExp) {
  return screen.getByRole('spinbutton', { name })
}

async function retype(input: HTMLElement, value: string) {
  await userEvent.clear(input)
  if (value !== '') {
    await userEvent.type(input, value)
  }
}

describe('SettingsTab', () => {
  beforeEach(() => {
    vi.resetAllMocks()
    vi.mocked(service.getSettings).mockResolvedValue(saved)
    vi.mocked(aiAccess.getAIAccessStatus).mockResolvedValue({ state: 'off', address: null, reason: null, code: null, clients: 0, stdioCommand: '', httpConfig: '', stdioConfig: '' })
    vi.mocked(gitAccounts.listGitAccounts).mockResolvedValue([])
  })

  afterEach(() => {
    act(() => {
      useSettingsStore.setState({ settings: defaultAppSettings(), status: 'idle', loadError: null, fieldErrors: {}, saving: false, saveError: null, aiStatus: null, pageOpen: false })
      useSettingsStore.getState().discard()
      useGitAccountsStore.setState({ accounts: [], error: null })
      useTabsStore.getState().reset()
    })
  })

  it('holds the Scripts, History, AI access, Git accounts and Shortcuts sections with the saved values', async () => {
    await renderTab()
    expect(screen.getByRole('heading', { name: 'Settings' })).toBeInTheDocument()
    for (const section of ['Scripts', 'History', 'AI access (MCP)', 'Git accounts']) {
      expect(screen.getByRole('group', { name: section })).toBeInTheDocument()
    }
    expect(screen.getByRole('heading', { name: 'Shortcuts' })).toBeInTheDocument()
    expect(field(/Script Timeout/)).toHaveValue(3)
    expect(field(/Script Timeout/)).toHaveAccessibleDescription(/including waiting for its timers and requests/)
    expect(field(/User Cap/)).toHaveValue(5000)
    expect(field(/AI Cap/)).toHaveValue(200)
    expect(field(/Port/)).toHaveValue(47821)
  })

  it('validates every field on Save, shows each error inline, focuses the first and saves nothing', async () => {
    await renderTab()
    await retype(field(/Script Timeout/), '61')
    await retype(field(/User Cap/), '99')
    await retype(field(/AI Cap/), '20001')
    await retype(field(/Port/), '80')
    await userEvent.click(screen.getByRole('button', { name: 'Save' }))
    expect(service.saveSettings).not.toHaveBeenCalled()
    expect(screen.getAllByRole('alert').map((alert) => alert.textContent)).toEqual([
      'Enter a whole number of seconds from 1 to 60.',
      'Enter a whole number from 100 to 20,000.',
      'Enter a whole number from 100 to 20,000.',
      'Enter a port from 1024 to 65535.',
    ])
    expect(field(/Port/)).toHaveAttribute('aria-invalid', 'true')
    expect(field(/Script Timeout/)).toHaveFocus()
    await retype(field(/Script Timeout/), '5')
    expect(field(/Script Timeout/)).not.toHaveAttribute('aria-invalid')
    expect(screen.getAllByRole('alert')).toHaveLength(3)
  })

  it('saves every form field together with one Save', async () => {
    vi.mocked(service.saveSettings).mockImplementation(async (value) => value)
    await renderTab()
    expect(isSettingsDirty(useSettingsStore.getState())).toBe(false)
    await retype(field(/Script Timeout/), '30')
    await retype(field(/User Cap/), '10000')
    await retype(field(/AI Cap/), '100')
    await retype(field(/Port/), '50000')
    await userEvent.click(screen.getByRole('switch', { name: 'AI Access' }))
    await userEvent.click(screen.getByRole('checkbox', { name: 'Dangerously skip approvals' }))
    expect(isSettingsDirty(useSettingsStore.getState())).toBe(true)
    await userEvent.click(screen.getByRole('button', { name: 'Save' }))
    expect(service.saveSettings).toHaveBeenCalledTimes(1)
    expect(service.saveSettings).toHaveBeenCalledWith({ scriptTimeoutMs: 30000, historyUserCap: 10000, historyAiCap: 100, aiAccess: { enabled: false, port: 50000, skipApprovals: true } })
    expect(isSettingsDirty(useSettingsStore.getState())).toBe(false)
  })

  it('shows a range refused by Go inline and other save failures in a banner', async () => {
    await renderTab()
    vi.mocked(service.saveSettings).mockRejectedValueOnce(new AppError('SET-04400', 'port out of range'))
    await userEvent.click(screen.getByRole('button', { name: 'Save' }))
    expect(await screen.findByRole('alert')).toHaveTextContent('Enter a port from 1024 to 65535.')
    expect(field(/Port/)).toHaveFocus()
    vi.mocked(service.saveSettings).mockRejectedValueOnce(new AppError('SCR-03500', 'app.json could not be written'))
    await userEvent.click(screen.getByRole('button', { name: 'Save' }))
    expect(await screen.findByText('app.json could not be written. Try again.')).toBeInTheDocument()
  })

  it('shows why the settings could not be read', async () => {
    vi.mocked(service.getSettings).mockRejectedValue(new AppError('UNK-00500', 'Settings are unreadable'))
    render(<SettingsTab />)
    expect(await screen.findByText('Settings are unreadable. Close Settings and open it again.')).toBeInTheDocument()
  })

  it('removes a Git account at once, without Save', async () => {
    vi.mocked(gitAccounts.listGitAccounts).mockResolvedValue([
      { id: 'a1', kind: 'token', provider: 'gitlab', sshMode: '', displayName: 'luis @ gitlab.com', host: 'gitlab.com', username: 'luis', commitName: 'Luis', commitEmail: '', avatarUrl: '', publicKey: '', hasPassphrase: false },
    ])
    vi.mocked(gitAccounts.removeGitAccount).mockResolvedValue()
    await renderTab()
    const list = await screen.findByRole('list', { name: 'Git accounts' })
    await userEvent.click(within(list).getByRole('button', { name: 'Remove luis @ gitlab.com' }))
    expect(gitAccounts.removeGitAccount).toHaveBeenCalledWith('a1')
    expect(service.saveSettings).not.toHaveBeenCalled()
    expect(screen.queryByRole('list', { name: 'Git accounts' })).not.toBeInTheDocument()
  })

  it('opens the Settings tab from the gear button and focuses it on a second click', async () => {
    useTabsStore.getState().hydrate('ws', { ...defaultUIState(), tabs: [{ tabId: 'start', kind: 'globals', requestId: null, pinned: true, baseRevision: null, draft: null }], activeTabId: 'start' })
    render(<SettingsButton />)
    await userEvent.click(screen.getByRole('button', { name: 'Settings' }))
    const opened = useTabsStore.getState().tabs.filter((tab) => tab.kind === 'settings')
    expect(opened).toHaveLength(1)
    expect(useTabsStore.getState().activeTabId).toBe(opened[0]!.id)
    useTabsStore.getState().activate('start')
    await userEvent.click(screen.getByRole('button', { name: 'Settings' }))
    expect(useTabsStore.getState().tabs.filter((tab) => tab.kind === 'settings')).toHaveLength(1)
    expect(useTabsStore.getState().activeTabId).toBe(opened[0]!.id)
  })

  it('opens the full-page Settings from the gear button with no workspace, with a close button', async () => {
    render(<SettingsButton />)
    await userEvent.click(screen.getByRole('button', { name: 'Settings' }))
    expect(useSettingsStore.getState().pageOpen).toBe(true)
    const onClose = vi.fn()
    await renderTab({ onClose })
    await userEvent.click(screen.getByRole('button', { name: 'Close Settings' }))
    expect(onClose).toHaveBeenCalledTimes(1)
  })

  it('has no close button inside a tab', async () => {
    await renderTab()
    expect(screen.queryByRole('button', { name: 'Close Settings' })).not.toBeInTheDocument()
  })
})
