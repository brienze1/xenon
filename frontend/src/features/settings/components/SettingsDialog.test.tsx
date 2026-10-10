import { act, render, screen } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest'
import { AppError } from '@/platform/errors'

vi.mock('@/platform/services/settings', () => ({ SCRIPT_TIMEOUT_OUT_OF_RANGE: 'SCR-01400', getSettings: vi.fn(), saveSettings: vi.fn() }))

const service = await import('@/platform/services/settings')
const { useSettingsStore } = await import('@/features/settings/store')
const { SettingsDialog } = await import('./SettingsDialog')
const { SettingsButton } = await import('./SettingsButton')

function renderSettings() {
  return render(
    <>
      <SettingsButton />
      <SettingsDialog />
    </>,
  )
}

describe('SettingsDialog', () => {
  beforeEach(() => {
    vi.resetAllMocks()
    vi.mocked(service.getSettings).mockResolvedValue({ scriptTimeoutMs: 10000 })
  })

  afterEach(() => {
    act(() => useSettingsStore.setState({ open: false, settings: { scriptTimeoutMs: 10000 }, loading: false, error: null }))
  })

  it('opens from the settings button with the saved timeout in seconds', async () => {
    vi.mocked(service.getSettings).mockResolvedValue({ scriptTimeoutMs: 2500 })
    renderSettings()
    expect(screen.queryByRole('dialog')).not.toBeInTheDocument()
    await userEvent.click(screen.getByRole('button', { name: 'Settings' }))
    expect(await screen.findByRole('dialog', { name: 'Settings' })).toBeInTheDocument()
    const field = screen.getByRole('spinbutton', { name: /Script Timeout/ })
    expect(field).toHaveValue(2.5)
    expect(field).toHaveAccessibleDescription(/including waiting for its timers and requests/)
  })

  it('refuses values outside 1 to 60 seconds before saving', async () => {
    renderSettings()
    await userEvent.click(screen.getByRole('button', { name: 'Settings' }))
    const field = await screen.findByRole('spinbutton', { name: /Script Timeout/ })
    for (const value of ['0', '61', '1.5', ' ']) {
      await userEvent.clear(field)
      await userEvent.type(field, value)
      await userEvent.click(screen.getByRole('button', { name: 'Save' }))
      expect(screen.getByRole('alert')).toHaveTextContent('Enter a whole number of seconds from 1 to 60.')
      expect(field).toHaveAttribute('aria-invalid', 'true')
      expect(field).toHaveFocus()
    }
    expect(service.saveSettings).not.toHaveBeenCalled()
    await userEvent.type(field, '5')
    expect(screen.queryByRole('alert')).not.toBeInTheDocument()
  })

  it('saves the timeout in milliseconds and closes', async () => {
    vi.mocked(service.saveSettings).mockResolvedValue({ scriptTimeoutMs: 30000 })
    renderSettings()
    await userEvent.click(screen.getByRole('button', { name: 'Settings' }))
    const field = await screen.findByRole('spinbutton', { name: /Script Timeout/ })
    await userEvent.clear(field)
    await userEvent.type(field, '30')
    await userEvent.click(screen.getByRole('button', { name: 'Save' }))
    expect(service.saveSettings).toHaveBeenCalledWith({ scriptTimeoutMs: 30000 })
    expect(screen.queryByRole('dialog')).not.toBeInTheDocument()
    expect(useSettingsStore.getState().settings).toEqual({ scriptTimeoutMs: 30000 })
  })

  it('shows the range message when Go refuses the value and a banner for other errors', async () => {
    renderSettings()
    await userEvent.click(screen.getByRole('button', { name: 'Settings' }))
    const field = await screen.findByRole('spinbutton', { name: /Script Timeout/ })
    vi.mocked(service.saveSettings).mockRejectedValueOnce(new AppError('SCR-01400', 'out of range'))
    await userEvent.click(screen.getByRole('button', { name: 'Save' }))
    expect(await screen.findByRole('alert')).toHaveTextContent('Enter a whole number of seconds from 1 to 60.')
    expect(field).toHaveFocus()
    vi.mocked(service.saveSettings).mockRejectedValueOnce(new AppError('SCR-03500', 'app.json could not be written'))
    await userEvent.click(screen.getByRole('button', { name: 'Save' }))
    expect(await screen.findByText('app.json could not be written. Try again.')).toBeInTheDocument()
    expect(screen.getByRole('dialog', { name: 'Settings' })).toBeInTheDocument()
  })

  it('shows why the settings could not be read and closes on Cancel', async () => {
    vi.mocked(service.getSettings).mockRejectedValue(new AppError('UNK-00500', 'Settings are unreadable'))
    renderSettings()
    await userEvent.click(screen.getByRole('button', { name: 'Settings' }))
    expect(await screen.findByText('Settings are unreadable. Close and open Settings again.')).toBeInTheDocument()
    await userEvent.click(screen.getByRole('button', { name: 'Cancel' }))
    expect(screen.queryByRole('dialog')).not.toBeInTheDocument()
  })
})
