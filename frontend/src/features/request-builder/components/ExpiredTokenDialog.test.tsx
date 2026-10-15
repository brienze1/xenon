import { render, screen } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest'
import { AppError } from '@/platform/errors'

vi.mock('@/platform/services/requests', async (importOriginal) => {
  const { confirmationOf, expiredTokenOwnerOf, missingSecretsOf } = await importOriginal<typeof import('@/platform/services/requests')>()
  return { confirmationOf, expiredTokenOwnerOf, missingSecretsOf, sendRequest: vi.fn(), autoHeaders: vi.fn() }
})

const requests = await import('@/platform/services/requests')
const { useRequestBuilderStore } = await import('@/features/request-builder/store')
const { useWorkspacesStore } = await import('@/features/workspaces/store')
const { ExpiredTokenDialog } = await import('./ExpiredTokenDialog')

const TAB = 'tab-1'

function refuseOnce() {
  vi.mocked(requests.sendRequest).mockReturnValueOnce({ result: Promise.reject(new AppError('SND-07409', 'expired', [{ attribute: 'owner_id', messages: ['c1'] }])), cancel: vi.fn() })
}

describe('ExpiredTokenDialog', () => {
  beforeEach(() => {
    vi.resetAllMocks()
    useRequestBuilderStore.getState().reset()
    useRequestBuilderStore.getState().ensureTab(TAB)
    useWorkspacesStore.setState({ activeLocalId: 'ws' })
  })

  afterEach(() => useWorkspacesStore.setState({ activeLocalId: null }))

  it('asks after a refused send and sends anyway with the acknowledgement', async () => {
    const user = userEvent.setup()
    refuseOnce()
    render(<ExpiredTokenDialog tabId={TAB} />)
    await useRequestBuilderStore.getState().send(TAB)
    expect(await screen.findByRole('dialog', { name: 'Access Token Expired' })).toBeInTheDocument()
    expect(useRequestBuilderStore.getState().tabs[TAB]!.expiredTokenOwner).toBe('c1')
    vi.mocked(requests.sendRequest).mockReturnValueOnce({ result: new Promise(() => undefined), cancel: vi.fn() })
    await user.click(screen.getByRole('button', { name: 'Send Anyway' }))
    expect(requests.sendRequest).toHaveBeenLastCalledWith('ws', TAB, expect.any(Object), { confirmedEnvironmentId: '', missingSecrets: false, expiredToken: true }, null)
    expect(screen.queryByRole('dialog')).not.toBeInTheDocument()
  })

  it('opens the Authorization tab to get a new token, or cancels', async () => {
    const user = userEvent.setup()
    refuseOnce()
    render(<ExpiredTokenDialog tabId={TAB} />)
    await useRequestBuilderStore.getState().send(TAB)
    await user.click(await screen.findByRole('button', { name: 'Get New Access Token' }))
    expect(useRequestBuilderStore.getState().tabs[TAB]!.subTab).toBe('auth')
    expect(screen.queryByRole('dialog')).not.toBeInTheDocument()
    refuseOnce()
    await useRequestBuilderStore.getState().send(TAB)
    await user.click(await screen.findByRole('button', { name: 'Cancel' }))
    expect(useRequestBuilderStore.getState().tabs[TAB]!.expiredTokenOwner).toBeNull()
    expect(requests.sendRequest).toHaveBeenCalledTimes(2)
  })
})
