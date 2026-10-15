import { act, render, screen, within } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest'
import { defaultAppSettings } from '@/platform/mappers/settings'
import { idleUpdateStatus, type UpdateStatus } from '@/platform/mappers/update-status'

vi.mock('@/platform/services/settings', () => ({
  SCRIPT_TIMEOUT_OUT_OF_RANGE: 'SCR-01400',
  HISTORY_USER_CAP_OUT_OF_RANGE: 'SET-02400',
  HISTORY_AI_CAP_OUT_OF_RANGE: 'SET-03400',
  AI_PORT_OUT_OF_RANGE: 'SET-04400',
  getSettings: vi.fn(),
  saveSettings: vi.fn(),
}))
vi.mock('@/platform/services/ai-access', () => ({ getAIAccessStatus: vi.fn(), copyText: vi.fn() }))
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
const { useSettingsStore } = await import('@/features/settings/store')
const { useUpdatesStore } = await import('@/features/updates/store')
const { AboutUpdatesSection } = await import('./AboutUpdatesSection')

function updateStatus(patch: Partial<UpdateStatus> = {}): UpdateStatus {
  return { ...idleUpdateStatus(), currentVersion: '0.1.4', latestVersion: '0.1.7', releaseUrl: 'https://github.com/brienze1/xenon/releases/tag/v0.1.7', ...patch }
}

const NOT_WRITABLE = "Xenon can't replace itself at /Applications. Move Xenon to a folder you own, such as ~/Applications, or install it from the release page."

interface Row {
  name: string
  status: Partial<UpdateStatus>
  lines: string[]
  latest: boolean
  whatsNew: boolean
  buttons: string[]
}

const ROWS: Row[] = [
  { name: 'idle', status: { state: 'idle' }, lines: [], latest: false, whatsNew: false, buttons: ['Check for updates'] },
  { name: 'checking', status: { state: 'checking' }, lines: ['Checking…'], latest: false, whatsNew: false, buttons: ['Check for updates'] },
  { name: 'up_to_date', status: { state: 'up_to_date', latestVersion: '0.1.4' }, lines: ['Xenon is up to date'], latest: true, whatsNew: false, buttons: ['Check for updates'] },
  { name: 'available', status: { state: 'available' }, lines: ['Update available'], latest: true, whatsNew: true, buttons: ['Update', 'Check for updates'] },
  { name: 'no_package', status: { state: 'no_package' }, lines: ['No update package for this system'], latest: true, whatsNew: true, buttons: ['Check for updates'] },
  { name: 'check_failed', status: { state: 'check_failed', reason: 'unreachable' }, lines: ["Couldn't check for updates"], latest: false, whatsNew: false, buttons: ['Check for updates'] },
  { name: 'downloading', status: { state: 'downloading', received: 2_500, total: 10_000 }, lines: ['Downloading… 25%'], latest: true, whatsNew: true, buttons: [] },
  { name: 'ready', status: { state: 'ready' }, lines: ['Update ready'], latest: true, whatsNew: true, buttons: ['Restart now'] },
  { name: 'installing', status: { state: 'installing' }, lines: ['Installing…'], latest: false, whatsNew: false, buttons: [] },
  {
    name: 'failed (download_failed)',
    status: { state: 'failed', reason: 'download_failed' },
    lines: ['The update could not be downloaded'],
    latest: false,
    whatsNew: false,
    buttons: ['Update', 'Check for updates'],
  },
  {
    name: 'failed (unverified)',
    status: { state: 'failed', reason: 'unverified' },
    lines: ['The update could not be verified — nothing was changed'],
    latest: false,
    whatsNew: false,
    buttons: ['Update', 'Check for updates'],
  },
  {
    name: 'failed (not_packaged)',
    status: { state: 'failed', reason: 'not_packaged' },
    lines: ['Updates install only in a packaged Xenon'],
    latest: false,
    whatsNew: false,
    buttons: ['Check for updates'],
  },
  {
    name: 'failed (not_writable)',
    status: { state: 'failed', reason: 'not_writable', folder: '/Applications' },
    lines: [NOT_WRITABLE],
    latest: false,
    whatsNew: false,
    buttons: ['Open the release page', 'Check for updates'],
  },
  {
    name: 'failed (install_failed)',
    status: { state: 'failed', reason: 'install_failed' },
    lines: ['The update could not be installed — Xenon was not changed'],
    latest: false,
    whatsNew: false,
    buttons: ['Update', 'Check for updates'],
  },
]

function show(patch: Partial<UpdateStatus>) {
  act(() => useUpdatesStore.setState({ status: updateStatus(patch) }))
}

function section() {
  return screen.getByRole('group', { name: 'About & Updates' })
}

function buttonNames() {
  return within(section())
    .queryAllByRole('button')
    .map((button) => button.textContent)
    .filter((name) => name !== "What's new")
}

describe('AboutUpdatesSection', () => {
  beforeEach(() => {
    vi.resetAllMocks()
    useUpdatesStore.setState({ status: idleUpdateStatus(), toast: 'unseen' })
    useSettingsStore.setState({ settings: defaultAppSettings(), fieldErrors: {} })
    useSettingsStore.getState().discard()
  })

  afterEach(() => {
    act(() => useUpdatesStore.setState({ status: idleUpdateStatus() }))
  })

  it.each(ROWS)('shows the $name row', ({ status, lines, latest, whatsNew, buttons }) => {
    render(<AboutUpdatesSection />)
    show(status)
    expect(within(section()).getByText('Version 0.1.4')).toBeInTheDocument()
    const statusArea = screen.getByRole('status', { name: 'Update status' })
    for (const line of lines) {
      expect(within(statusArea).getByText(line)).toBeInTheDocument()
    }
    if (lines.length === 0) {
      expect(statusArea).toBeEmptyDOMElement()
    }
    const latestLine = within(section()).queryByText(/^Latest version/)
    if (latest) {
      expect(latestLine).toHaveTextContent(`Latest version ${status.latestVersion ?? '0.1.7'}`)
    } else {
      expect(latestLine).not.toBeInTheDocument()
    }
    expect(within(section()).queryByRole('button', { name: "What's new" }) !== null).toBe(whatsNew)
    expect(buttonNames()).toEqual(buttons)
  })

  it('disables Check for updates while checking', () => {
    render(<AboutUpdatesSection />)
    show({ state: 'checking' })
    expect(screen.getByRole('button', { name: 'Check for updates' })).toBeDisabled()
    show({ state: 'up_to_date' })
    expect(screen.getByRole('button', { name: 'Check for updates' })).toBeEnabled()
  })

  it('shows the download progress as a bar', () => {
    render(<AboutUpdatesSection />)
    show({ state: 'downloading', received: 7_500, total: 10_000 })
    expect(screen.getByRole('progressbar', { name: 'Download progress' })).toHaveAttribute('value', '75')
    expect(screen.getByText('Downloading… 75%')).toBeInTheDocument()
    show({ state: 'downloading', received: 0, total: 0 })
    expect(screen.getByText('Downloading… 0%')).toBeInTheDocument()
  })

  it('checks for updates', async () => {
    vi.mocked(service.checkForUpdates).mockResolvedValue(updateStatus({ state: 'up_to_date', latestVersion: '0.1.4' }))
    render(<AboutUpdatesSection />)
    show({ state: 'idle' })
    await userEvent.click(screen.getByRole('button', { name: 'Check for updates' }))
    expect(service.checkForUpdates).toHaveBeenCalledTimes(1)
    expect(await screen.findByText('Xenon is up to date')).toBeInTheDocument()
  })

  it.each(["What's new", 'Open the release page'])('opens the release page through Go from "%s"', async (name) => {
    vi.mocked(service.openReleasePage).mockImplementation(async () => useUpdatesStore.getState().status)
    render(<AboutUpdatesSection />)
    show(name === "What's new" ? { state: 'available' } : { state: 'failed', reason: 'not_writable', folder: '/Applications' })
    await userEvent.click(screen.getByRole('button', { name }))
    expect(service.openReleasePage).toHaveBeenCalledTimes(1)
  })

  it('updates and then restarts into the verified update', async () => {
    vi.mocked(service.startUpdate).mockResolvedValue(updateStatus({ state: 'ready' }))
    vi.mocked(service.restartToUpdate).mockResolvedValue(updateStatus({ state: 'installing' }))
    render(<AboutUpdatesSection />)
    show({ state: 'available' })
    await userEvent.click(screen.getByRole('button', { name: 'Update' }))
    expect(service.startUpdate).toHaveBeenCalledTimes(1)
    expect(service.restartToUpdate).toHaveBeenCalledTimes(1)
    expect(await screen.findByText('Installing…')).toBeInTheDocument()
  })

  it('restarts from Restart now', async () => {
    vi.mocked(service.restartToUpdate).mockResolvedValue(updateStatus({ state: 'failed', reason: 'install_failed' }))
    render(<AboutUpdatesSection />)
    show({ state: 'ready' })
    await userEvent.click(screen.getByRole('button', { name: 'Restart now' }))
    expect(service.restartToUpdate).toHaveBeenCalledTimes(1)
    expect(await screen.findByText('The update could not be installed — Xenon was not changed')).toBeInTheDocument()
  })

  it('edits Update automatically in the form draft, on by default', async () => {
    render(<AboutUpdatesSection />)
    const toggle = screen.getByRole('switch', { name: 'Update automatically' })
    expect(toggle).toBeChecked()
    await userEvent.click(toggle)
    expect(toggle).not.toBeChecked()
    expect(useSettingsStore.getState().draft.autoUpdate).toBe(false)
  })
})
