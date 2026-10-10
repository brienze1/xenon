import { render, screen, within } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest'

vi.mock('@/platform/services/requests', async (importOriginal) => {
  const { confirmationOf, missingSecretsOf } = await importOriginal<typeof import('@/platform/services/requests')>()
  return {
    confirmationOf,
    missingSecretsOf,
    sendRequest: vi.fn(),
    autoHeaders: vi.fn(),
  }
})
vi.mock('@/platform/services/environments', () => ({ listEnvironments: vi.fn(), setActiveEnvironment: vi.fn() }))

const requests = await import('@/platform/services/requests')
const { useRequestBuilderStore } = await import('@/features/request-builder/store')
const { useEnvironmentsStore } = await import('@/features/environments/store')
const { useWorkspacesStore } = await import('@/features/workspaces/store')
const { MissingSecretsDialog } = await import('./MissingSecretsDialog')

const TAB = 'tab-1'

function refuse(missing: string[], locked: string[]) {
  useRequestBuilderStore.getState().ensureTab(TAB)
  useRequestBuilderStore.setState((state) => ({ tabs: { ...state.tabs, [TAB]: { ...state.tabs[TAB]!, missingSecrets: { missing, locked } } } }))
}

describe('MissingSecretsDialog', () => {
  beforeEach(() => {
    vi.resetAllMocks()
    useRequestBuilderStore.getState().reset()
    useWorkspacesStore.setState({ activeLocalId: 'ws' })
  })

  afterEach(() => {
    useRequestBuilderStore.getState().reset()
    useEnvironmentsStore.setState({ passphrasePrompt: null })
    useWorkspacesStore.setState({ activeLocalId: null })
  })

  it('stays closed until a send is refused', () => {
    useRequestBuilderStore.getState().ensureTab(TAB)
    render(<MissingSecretsDialog tabId={TAB} />)
    expect(screen.queryByRole('dialog')).not.toBeInTheDocument()
  })

  it('lists missing and locked secrets and cancels without sending', async () => {
    refuse(['api_key'], ['token'])
    render(<MissingSecretsDialog tabId={TAB} />)
    const dialog = screen.getByRole('dialog', { name: 'Send without some secrets?' })
    expect(within(dialog).getByText('No value on this machine')).toBeInTheDocument()
    expect(within(dialog).getByText('{{api_key}}')).toBeInTheDocument()
    expect(within(dialog).getByText('In a locked vault')).toBeInTheDocument()
    expect(within(dialog).getByText('{{token}}')).toBeInTheDocument()
    await userEvent.click(screen.getByRole('button', { name: 'Cancel' }))
    expect(screen.queryByRole('dialog')).not.toBeInTheDocument()
    expect(requests.sendRequest).not.toHaveBeenCalled()
    expect(useRequestBuilderStore.getState().tabs[TAB]?.missingSecrets).toBeNull()
  })

  it('sends anyway with the acknowledgement', async () => {
    refuse(['api_key'], [])
    vi.mocked(requests.sendRequest).mockReturnValue({ result: new Promise(() => undefined), cancel: vi.fn() })
    render(<MissingSecretsDialog tabId={TAB} />)
    expect(screen.queryByText('In a locked vault')).not.toBeInTheDocument()
    expect(screen.queryByRole('button', { name: 'Unlock Vault' })).not.toBeInTheDocument()
    await userEvent.click(screen.getByRole('button', { name: 'Send Anyway' }))
    expect(requests.sendRequest).toHaveBeenCalledWith('ws', TAB, expect.any(Object), { confirmedEnvironmentId: '', missingSecrets: true })
    expect(screen.queryByRole('dialog')).not.toBeInTheDocument()
  })

  it('offers to unlock the vault for locked secrets', async () => {
    refuse([], ['token'])
    render(<MissingSecretsDialog tabId={TAB} />)
    await userEvent.click(screen.getByRole('button', { name: 'Unlock Vault' }))
    expect(useEnvironmentsStore.getState().passphrasePrompt).toBe('unlock')
    expect(requests.sendRequest).not.toHaveBeenCalled()
  })
})
