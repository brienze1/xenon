import { act, render, screen } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest'
import { AppError } from '@/platform/errors'

vi.mock('@/platform/services/git-accounts', () => ({
  GITHUB_SIGN_IN_NOT_CONFIGURED: 'ACC-01503',
  listGitAccounts: vi.fn(),
  startGitHubSignIn: vi.fn(),
  pollGitHubSignIn: vi.fn(),
}))

const service = await import('@/platform/services/git-accounts')
const { useGitAccountsStore } = await import('@/features/settings/store')
const { GitHubSignIn } = await import('./GitHubSignIn')

const code = { sessionId: 's1', userCode: 'WDJB-MJHT', verificationUri: 'https://github.com/login/device', expiresIn: 900, interval: 5 }

describe('GitHubSignIn', () => {
  beforeEach(() => {
    vi.resetAllMocks()
    vi.useFakeTimers({ shouldAdvanceTime: true })
    vi.mocked(service.listGitAccounts).mockResolvedValue([])
  })

  afterEach(() => {
    vi.useRealTimers()
  })

  it('shows the code and polls at the interval until GitHub approves', async () => {
    vi.mocked(service.startGitHubSignIn).mockResolvedValueOnce(code)
    vi.mocked(service.pollGitHubSignIn)
      .mockResolvedValueOnce({ status: 'pending', interval: 10, account: null })
      .mockResolvedValueOnce({ status: 'done', interval: 10, account: null })
    const onClose = vi.fn()
    render(<GitHubSignIn onClose={onClose} />)
    expect(await screen.findByText('WDJB-MJHT')).toBeInTheDocument()
    expect(screen.getByText(/https:\/\/github.com\/login\/device/)).toBeInTheDocument()
    await act(() => vi.advanceTimersByTimeAsync(4_900))
    expect(service.pollGitHubSignIn).not.toHaveBeenCalled()
    await act(() => vi.advanceTimersByTimeAsync(200))
    expect(service.pollGitHubSignIn).toHaveBeenCalledWith('s1')
    await act(() => vi.advanceTimersByTimeAsync(9_000))
    expect(service.pollGitHubSignIn).toHaveBeenCalledTimes(1)
    await act(() => vi.advanceTimersByTimeAsync(1_100))
    expect(await screen.findByText('Signed in. The account is listed above.')).toBeInTheDocument()
    expect(service.listGitAccounts).toHaveBeenCalled()
    await userEvent.click(screen.getByRole('button', { name: 'Done' }))
    expect(onClose).toHaveBeenCalled()
  })

  it('explains that sign-in is not configured', async () => {
    vi.mocked(service.startGitHubSignIn).mockRejectedValueOnce(new AppError('ACC-01503', 'not configured'))
    render(<GitHubSignIn onClose={vi.fn()} />)
    expect(await screen.findByRole('alert')).toHaveTextContent('GitHub sign-in is not configured in this build of Xenon')
    expect(screen.getByRole('button', { name: 'Cancel' })).toBeInTheDocument()
  })

  it.each([
    ['expired', 'The code expired.'],
    ['denied', 'The sign-in was denied on GitHub.'],
  ] as const)('stops when the sign-in is %s', async (status, message) => {
    vi.mocked(service.startGitHubSignIn).mockResolvedValueOnce({ ...code, interval: 1 })
    vi.mocked(service.pollGitHubSignIn).mockResolvedValueOnce({ status, interval: 1, account: null })
    render(<GitHubSignIn onClose={vi.fn()} />)
    await screen.findByText('WDJB-MJHT')
    await act(() => vi.advanceTimersByTimeAsync(1_100))
    expect(await screen.findByText(new RegExp(message))).toBeInTheDocument()
    await act(() => vi.advanceTimersByTimeAsync(5_000))
    expect(service.pollGitHubSignIn).toHaveBeenCalledTimes(1)
  })

  it('shows failures with a next step and resets on close', async () => {
    vi.mocked(service.startGitHubSignIn).mockResolvedValueOnce({ ...code, interval: 0 })
    vi.mocked(service.pollGitHubSignIn).mockRejectedValueOnce(new AppError('ACC-46503', 'GitHub could not be reached'))
    const view = render(<GitHubSignIn onClose={vi.fn()} />)
    await screen.findByText('WDJB-MJHT')
    await act(() => vi.advanceTimersByTimeAsync(1_100))
    expect(await screen.findByRole('alert')).toHaveTextContent('GitHub could not be reached. Try again.')
    view.unmount()
    expect(useGitAccountsStore.getState().signIn.phase).toBe('idle')
  })

  it('shows a start failure', async () => {
    vi.mocked(service.startGitHubSignIn).mockRejectedValueOnce(new AppError('ACC-44503', 'GitHub could not be reached'))
    render(<GitHubSignIn onClose={vi.fn()} />)
    expect(screen.getByText('Asking GitHub for a sign-in code…')).toBeInTheDocument()
    expect(await screen.findByRole('alert')).toHaveTextContent('GitHub could not be reached. Try again.')
  })
})
