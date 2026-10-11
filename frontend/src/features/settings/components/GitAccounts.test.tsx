import { act, render, screen, within } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest'
import { AppError } from '@/platform/errors'
import type { GitAccount } from '@/platform/mappers/git-accounts'

vi.mock('@/platform/services/git-accounts', () => ({
  GITHUB_SIGN_IN_NOT_CONFIGURED: 'ACC-01503',
  listGitAccounts: vi.fn(),
  addTokenAccount: vi.fn(),
  addSshAccount: vi.fn(),
  updateGitAccount: vi.fn(),
  removeGitAccount: vi.fn(),
  startGitHubSignIn: vi.fn(),
  pollGitHubSignIn: vi.fn(),
  uploadPublicKey: vi.fn(),
  confirmHostKey: vi.fn(),
  copyPublicKey: vi.fn(),
}))

const service = await import('@/platform/services/git-accounts')
const { useGitAccountsStore } = await import('@/features/settings/store')
const { GitAccounts } = await import('./GitAccounts')

function account(overrides: Partial<GitAccount>): GitAccount {
  return {
    id: 'a1',
    kind: 'token',
    provider: 'gitlab',
    sshMode: '',
    displayName: 'luis @ gitlab.com',
    host: 'gitlab.com',
    username: 'luis',
    commitName: 'Luis',
    commitEmail: 'luis@acme.test',
    avatarUrl: '',
    publicKey: '',
    hasPassphrase: false,
    ...overrides,
  }
}

const gitlab = account({})
const github = account({ id: 'g1', kind: 'github', provider: 'github', displayName: 'luis @ github.com', avatarUrl: 'https://a/7' })

describe('GitAccounts', () => {
  beforeEach(() => {
    vi.resetAllMocks()
    vi.mocked(service.listGitAccounts).mockResolvedValue([gitlab, github])
  })

  afterEach(() => {
    act(() => useGitAccountsStore.setState({ accounts: [], error: null, hostKeyPrompt: null }))
  })

  it('lists the accounts with their commit identity and removes one', async () => {
    render(<GitAccounts />)
    const list = await screen.findByRole('list', { name: 'Git accounts' })
    expect(within(list).getByText('luis @ gitlab.com')).toBeInTheDocument()
    expect(within(list).getAllByText(/commits as Luis <luis@acme.test>/)).toHaveLength(2)
    expect(within(list).getByText(/GitHub sign-in/)).toBeInTheDocument()
    vi.mocked(service.removeGitAccount).mockResolvedValueOnce()
    await userEvent.click(screen.getByRole('button', { name: 'Remove luis @ gitlab.com' }))
    expect(service.removeGitAccount).toHaveBeenCalledWith('a1')
    expect(screen.queryByText('luis @ gitlab.com')).not.toBeInTheDocument()
  })

  it('shows why an account could not be removed', async () => {
    render(<GitAccounts />)
    vi.mocked(service.removeGitAccount).mockRejectedValueOnce(new AppError('ACC-30503', 'the keychain failed'))
    await userEvent.click(await screen.findByRole('button', { name: 'Remove luis @ gitlab.com' }))
    expect(screen.getByText('the keychain failed. Try again.')).toBeInTheDocument()
  })

  it('explains an empty list and a failed load', async () => {
    vi.mocked(service.listGitAccounts).mockRejectedValueOnce(new AppError('VLT-25500', 'app.json unreadable'))
    render(<GitAccounts />)
    expect(await screen.findByText('app.json unreadable. Close and open Settings again.')).toBeInTheDocument()
    expect(screen.getByText(/No Git accounts yet/)).toBeInTheDocument()
  })

  it('adds a token account with the host presets and shows refusals', async () => {
    render(<GitAccounts />)
    await userEvent.click(await screen.findByRole('button', { name: 'Add token account' }))
    const form = screen.getByRole('region', { name: 'Add token account' })
    expect(within(form).getByRole('textbox', { name: 'Host name' })).toHaveValue('github.com')
    expect(within(form).queryByRole('textbox', { name: 'Username' })).not.toBeInTheDocument()
    await userEvent.selectOptions(within(form).getByRole('combobox', { name: 'Host' }), 'other')
    expect(within(form).getByRole('textbox', { name: 'Host name' })).toHaveValue('')
    await userEvent.selectOptions(within(form).getByRole('combobox', { name: 'Host' }), 'gitlab')
    expect(within(form).getByRole('textbox', { name: 'Host name' })).toHaveValue('gitlab.com')
    const save = within(form).getByRole('button', { name: 'Verify and add' })
    expect(save).toBeDisabled()
    await userEvent.type(within(form).getByLabelText('Token'), 'glpat-x')
    await userEvent.type(within(form).getByRole('textbox', { name: 'Repository to verify the token with' }), 'https://gitlab.com/a/b.git')
    await userEvent.type(within(form).getByRole('textbox', { name: 'Commit name' }), 'Luis')
    await userEvent.type(within(form).getByRole('textbox', { name: 'Commit email' }), 'l@acme.test')
    vi.mocked(service.addTokenAccount).mockRejectedValueOnce(new AppError('ACC-02401', 'the Git host refused the credentials'))
    await userEvent.click(save)
    expect(await within(form).findByRole('alert')).toHaveTextContent('the Git host refused the credentials. Check the token and try again.')
    vi.mocked(service.addTokenAccount).mockResolvedValueOnce(gitlab)
    await userEvent.click(save)
    expect(service.addTokenAccount).toHaveBeenLastCalledWith({
      provider: 'gitlab',
      host: 'gitlab.com',
      username: '',
      token: 'glpat-x',
      repository: 'https://gitlab.com/a/b.git',
      commitName: 'Luis',
      commitEmail: 'l@acme.test',
    })
    expect(screen.queryByRole('region', { name: 'Add token account' })).not.toBeInTheDocument()
    expect(screen.getByRole('button', { name: 'Add SSH key' })).toBeInTheDocument()
  })

  it('generates an SSH key, shows the public key to copy and uploads it to GitHub', async () => {
    render(<GitAccounts />)
    await userEvent.click(await screen.findByRole('button', { name: 'Add SSH key' }))
    const form = screen.getByRole('region', { name: 'Add SSH key' })
    expect(within(form).getByRole('radio', { name: 'Generate a new key' })).toBeChecked()
    expect(within(form).queryByRole('textbox', { name: 'Private key file' })).not.toBeInTheDocument()
    const ssh = account({ id: 's1', kind: 'ssh', sshMode: 'generated', displayName: 'SSH @ github.com', publicKey: 'ssh-ed25519 AAAA xenon' })
    vi.mocked(service.addSshAccount).mockResolvedValueOnce({ account: ssh, publicKey: 'ssh-ed25519 AAAA xenon' })
    await userEvent.click(within(form).getByRole('button', { name: 'Add SSH key' }))
    const panel = await screen.findByRole('region', { name: 'Public key' })
    expect(within(panel).getByText('ssh-ed25519 AAAA xenon')).toBeInTheDocument()
    await userEvent.click(within(panel).getByRole('button', { name: 'Copy public key' }))
    expect(service.copyPublicKey).toHaveBeenCalledWith('ssh-ed25519 AAAA xenon')
    expect(within(panel).getByRole('combobox', { name: 'GitHub account' })).toHaveValue('g1')
    vi.mocked(service.uploadPublicKey).mockRejectedValueOnce(new AppError('ACC-13422', 'GitHub refused the key'))
    await userEvent.click(within(panel).getByRole('button', { name: 'Add to GitHub' }))
    expect(await within(panel).findByRole('status')).toHaveTextContent('GitHub refused the key. Add the key on GitHub by hand.')
    vi.mocked(service.uploadPublicKey).mockResolvedValueOnce()
    await userEvent.click(within(panel).getByRole('button', { name: 'Add to GitHub' }))
    expect(service.uploadPublicKey).toHaveBeenLastCalledWith('s1', 'g1', 'Xenon')
    expect(within(panel).getByRole('status')).toHaveTextContent('The key was added to GitHub.')
    await userEvent.click(within(panel).getByRole('button', { name: 'Done' }))
    expect(screen.queryByRole('region', { name: 'Public key' })).not.toBeInTheDocument()
  })

  it('uses a key file with its passphrase and closes when there is nothing to show', async () => {
    render(<GitAccounts />)
    await userEvent.click(await screen.findByRole('button', { name: 'Add SSH key' }))
    const form = screen.getByRole('region', { name: 'Add SSH key' })
    await userEvent.click(within(form).getByRole('radio', { name: 'Use an existing key file' }))
    await userEvent.type(within(form).getByRole('textbox', { name: 'Private key file' }), '/home/l/.ssh/id')
    await userEvent.type(within(form).getByLabelText('Key passphrase (if any)'), 'pw')
    vi.mocked(service.addSshAccount).mockRejectedValueOnce(new AppError('ACC-42400', 'wrong passphrase'))
    await userEvent.click(within(form).getByRole('button', { name: 'Add SSH key' }))
    expect(await within(form).findByRole('alert')).toHaveTextContent('wrong passphrase. Check the key and try again.')
    await userEvent.click(within(form).getByRole('radio', { name: 'Use the SSH agent' }))
    vi.mocked(service.addSshAccount).mockResolvedValueOnce({ account: account({ kind: 'ssh' }), publicKey: '' })
    await userEvent.click(within(form).getByRole('button', { name: 'Add SSH key' }))
    expect(service.addSshAccount).toHaveBeenLastCalledWith({ mode: 'agent', host: 'github.com', keyPath: '/home/l/.ssh/id', passphrase: 'pw', commitName: '', commitEmail: '' })
    expect(screen.queryByRole('region', { name: 'Add SSH key' })).not.toBeInTheDocument()
  })
})
