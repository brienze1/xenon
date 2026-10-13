import { render, screen, within } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest'

vi.mock('@/platform/services/requests', async (importOriginal) => {
  const { confirmationOf, expiredTokenOwnerOf, missingSecretsOf } = await importOriginal<typeof import('@/platform/services/requests')>()
  return {
    confirmationOf,
    expiredTokenOwnerOf,
    missingSecretsOf,
    sendRequest: vi.fn(),
    autoHeaders: vi.fn(),
  }
})
vi.mock('@/platform/services/variables', () => ({ inspectVariables: vi.fn(async () => ({ references: [], variables: [], scopes: {} })) }))

const requests = await import('@/platform/services/requests')
const { useRequestBuilderStore } = await import('@/features/request-builder/store')
const { useWorkspacesStore } = await import('@/features/workspaces/store')
const { ConfirmSendDialog } = await import('./ConfirmSendDialog')

const TAB = 'tab-1'

function askToConfirm(color: string, missingSecrets = false) {
  useRequestBuilderStore.getState().ensureTab(TAB)
  useRequestBuilderStore.setState((state) => ({
    tabs: {
      ...state.tabs,
      [TAB]: {
        ...state.tabs[TAB]!,
        confirmation: { environmentId: '01P', name: 'prod', color },
        acknowledgement: { confirmedEnvironmentId: '', missingSecrets, expiredToken: false },
      },
    },
  }))
}

describe('ConfirmSendDialog', () => {
  beforeEach(() => {
    vi.resetAllMocks()
    useRequestBuilderStore.getState().reset()
    useWorkspacesStore.setState({ activeLocalId: 'ws' })
  })

  afterEach(() => {
    useRequestBuilderStore.getState().reset()
    useWorkspacesStore.setState({ activeLocalId: null })
  })

  it('stays closed until Go asks for a confirmation', () => {
    useRequestBuilderStore.getState().ensureTab(TAB)
    render(<ConfirmSendDialog tabId={TAB} />)
    expect(screen.queryByRole('dialog')).not.toBeInTheDocument()
  })

  it('names the environment and cancels without sending', async () => {
    askToConfirm('red')
    render(<ConfirmSendDialog tabId={TAB} />)
    const dialog = screen.getByRole('dialog', { name: 'Send to prod?' })
    expect(within(dialog).getByText('prod', { selector: 'strong' })).toBeInTheDocument()
    await userEvent.click(within(dialog).getByRole('button', { name: 'Cancel' }))
    expect(screen.queryByRole('dialog')).not.toBeInTheDocument()
    expect(requests.sendRequest).not.toHaveBeenCalled()
    expect(useRequestBuilderStore.getState().tabs[TAB]?.confirmation).toBeNull()
  })

  it('sends with the environment confirmed, keeping an earlier secrets acknowledgement', async () => {
    askToConfirm('none', true)
    vi.mocked(requests.sendRequest).mockReturnValue({ result: new Promise(() => undefined), cancel: vi.fn() })
    render(<ConfirmSendDialog tabId={TAB} />)
    await userEvent.click(screen.getByRole('button', { name: 'Send to prod' }))
    expect(requests.sendRequest).toHaveBeenCalledWith('ws', TAB, expect.any(Object), { confirmedEnvironmentId: '01P', missingSecrets: true, expiredToken: false })
    expect(screen.queryByRole('dialog')).not.toBeInTheDocument()
  })
})
