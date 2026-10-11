import { act, render, screen } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest'
import { AppError } from '@/platform/errors'

vi.mock('@/platform/services/git-accounts', () => ({ GITHUB_SIGN_IN_NOT_CONFIGURED: 'ACC-01503', confirmHostKey: vi.fn() }))

const service = await import('@/platform/services/git-accounts')
const { useGitAccountsStore } = await import('@/features/settings/store')
const { HostKeyPrompt } = await import('./HostKeyPrompt')

const prompt = { repository: 'ssh://git@git.acme.test/a/b.git', host: 'git.acme.test', fingerprint: 'SHA256:abc', keyType: 'ssh-ed25519' }

describe('HostKeyPrompt', () => {
  beforeEach(() => vi.resetAllMocks())

  afterEach(() => {
    act(() => useGitAccountsStore.setState({ hostKeyPrompt: null }))
  })

  it('stays closed without an unknown host', () => {
    render(<HostKeyPrompt />)
    expect(screen.queryByRole('dialog')).not.toBeInTheDocument()
  })

  it('shows the fingerprint and remembers the key on confirmation', async () => {
    act(() => useGitAccountsStore.getState().promptHostKey(prompt))
    render(<HostKeyPrompt />)
    const dialog = screen.getByRole('dialog', { name: 'Trust git.acme.test?' })
    expect(dialog).toHaveTextContent('SHA256:abc')
    expect(dialog).toHaveTextContent('ssh-ed25519')
    vi.mocked(service.confirmHostKey).mockRejectedValueOnce(new AppError('ACC-15400', 'the host now presents another fingerprint'))
    await userEvent.click(screen.getByRole('button', { name: 'Trust and remember' }))
    expect(await screen.findByText('the host now presents another fingerprint. Close and try the connection again.')).toBeInTheDocument()
    vi.mocked(service.confirmHostKey).mockResolvedValueOnce()
    await userEvent.click(screen.getByRole('button', { name: 'Trust and remember' }))
    expect(service.confirmHostKey).toHaveBeenLastCalledWith(prompt.repository, 'SHA256:abc', true)
    expect(screen.queryByRole('dialog')).not.toBeInTheDocument()
  })

  it('declines without remembering', async () => {
    act(() => useGitAccountsStore.getState().promptHostKey(prompt))
    render(<HostKeyPrompt />)
    vi.mocked(service.confirmHostKey).mockResolvedValueOnce()
    await userEvent.click(screen.getByRole('button', { name: 'Cancel' }))
    expect(service.confirmHostKey).toHaveBeenCalledWith(prompt.repository, 'SHA256:abc', false)
    expect(screen.queryByRole('dialog')).not.toBeInTheDocument()
  })

  it('answers nothing when no host is waiting', async () => {
    await expect(useGitAccountsStore.getState().answerHostKey(true)).resolves.toBeNull()
    expect(service.confirmHostKey).not.toHaveBeenCalled()
  })
})
