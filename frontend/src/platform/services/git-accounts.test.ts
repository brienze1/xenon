import { beforeEach, describe, expect, it, vi } from 'vitest'

vi.mock('@wailsio/runtime', () => ({ CancelError: class CancelError extends Error {}, Clipboard: { SetText: vi.fn() } }))
vi.mock('../../../bindings/github.com/brienze1/xenon/internal/infra/server/binding/gitaccountservice', () => ({
  List: vi.fn(),
  AddToken: vi.fn(),
  AddSSH: vi.fn(),
  Update: vi.fn(),
  Remove: vi.fn(),
  StartGitHubSignIn: vi.fn(),
  PollGitHubSignIn: vi.fn(),
  UploadPublicKey: vi.fn(),
  TestConnection: vi.fn(),
  ConfirmHostKey: vi.fn(),
}))

const binding = await import('../../../bindings/github.com/brienze1/xenon/internal/infra/server/binding/gitaccountservice')
const runtime = await import('@wailsio/runtime')
const service = await import('./git-accounts')
const { AppError } = await import('../errors')

const accountDto = {
  id: 'a1',
  kind: 'token',
  provider: 'gitlab',
  ssh_mode: '',
  display_name: 'luis @ gitlab.com',
  host: 'gitlab.com',
  username: 'luis',
  commit_name: 'Luis',
  commit_email: 'luis@acme.test',
  avatar_url: '',
  key_path: '',
  public_key: '',
  created_at: '2026-10-11T12:00:00.000Z',
  has_passphrase: false,
}

const account = {
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
}

describe('git accounts service', () => {
  beforeEach(() => vi.resetAllMocks())

  it('lists accounts and treats an empty answer as none', async () => {
    vi.mocked(binding.List).mockResolvedValueOnce([accountDto] as never)
    await expect(service.listGitAccounts()).resolves.toEqual([account])
    vi.mocked(binding.List).mockResolvedValueOnce(null as never)
    await expect(service.listGitAccounts()).resolves.toEqual([])
  })

  it('adds token and SSH accounts with the snake_case contract', async () => {
    vi.mocked(binding.AddToken).mockResolvedValueOnce(accountDto as never)
    const input = { provider: 'gitlab' as const, host: 'gitlab.com', username: 'luis', token: 't', repository: 'https://gitlab.com/a/b.git', commitName: 'Luis', commitEmail: 'l@x.test' }
    await expect(service.addTokenAccount(input)).resolves.toEqual(account)
    expect(binding.AddToken).toHaveBeenCalledWith({
      provider: 'gitlab',
      host: 'gitlab.com',
      username: 'luis',
      token: 't',
      repository: 'https://gitlab.com/a/b.git',
      display_name: '',
      commit_name: 'Luis',
      commit_email: 'l@x.test',
    })
    vi.mocked(binding.AddSSH).mockResolvedValueOnce({ account: { ...accountDto, kind: 'ssh' }, public_key: 'ssh-ed25519 AAAA xenon' } as never)
    const added = await service.addSshAccount({ mode: 'generate', host: 'github.com', keyPath: '', passphrase: '', commitName: '', commitEmail: '' })
    expect(added).toEqual({ account: { ...account, kind: 'ssh' }, publicKey: 'ssh-ed25519 AAAA xenon' })
    expect(binding.AddSSH).toHaveBeenCalledWith({ mode: 'generate', host: 'github.com', key_path: '', passphrase: '', display_name: '', commit_name: '', commit_email: '' })
  })

  it('updates and removes accounts', async () => {
    vi.mocked(binding.Update).mockResolvedValueOnce({ ...accountDto, display_name: 'Work' } as never)
    await expect(service.updateGitAccount('a1', 'Work', 'L', 'l@x.test')).resolves.toMatchObject({ displayName: 'Work' })
    expect(binding.Update).toHaveBeenCalledWith({ id: 'a1', display_name: 'Work', commit_name: 'L', commit_email: 'l@x.test' })
    vi.mocked(binding.Remove).mockResolvedValueOnce(undefined as never)
    await service.removeGitAccount('a1')
    expect(binding.Remove).toHaveBeenCalledWith({ id: 'a1' })
  })

  it('runs the GitHub device flow', async () => {
    vi.mocked(binding.StartGitHubSignIn).mockResolvedValueOnce({ session_id: 's', user_code: 'ABCD-1234', verification_uri: 'https://github.com/login/device', expires_in: 900, interval: 5 } as never)
    await expect(service.startGitHubSignIn()).resolves.toEqual({ sessionId: 's', userCode: 'ABCD-1234', verificationUri: 'https://github.com/login/device', expiresIn: 900, interval: 5 })
    vi.mocked(binding.PollGitHubSignIn).mockResolvedValueOnce({ status: 'pending', interval: 10 } as never)
    await expect(service.pollGitHubSignIn('s')).resolves.toEqual({ status: 'pending', interval: 10, account: null })
    expect(binding.PollGitHubSignIn).toHaveBeenCalledWith({ session_id: 's' })
    vi.mocked(binding.PollGitHubSignIn).mockResolvedValueOnce({ status: 'done', interval: 5, account: accountDto } as never)
    await expect(service.pollGitHubSignIn('s')).resolves.toEqual({ status: 'done', interval: 5, account })
  })

  it('uploads keys, tests connections, confirms host keys and copies public keys', async () => {
    vi.mocked(binding.UploadPublicKey).mockResolvedValueOnce(undefined as never)
    await service.uploadPublicKey('s1', 'g1', 'Xenon')
    expect(binding.UploadPublicKey).toHaveBeenCalledWith({ ssh_account_id: 's1', github_account_id: 'g1', title: 'Xenon' })
    vi.mocked(binding.TestConnection).mockResolvedValueOnce({ default_branch: 'main', branches: null } as never)
    await expect(service.testConnection('a1', 'https://h/r.git')).resolves.toEqual({ defaultBranch: 'main', branches: [] })
    expect(binding.TestConnection).toHaveBeenCalledWith({ account_id: 'a1', repository: 'https://h/r.git' })
    vi.mocked(binding.ConfirmHostKey).mockResolvedValueOnce(undefined as never)
    await service.confirmHostKey('ssh://h/r.git', 'SHA256:x', true)
    expect(binding.ConfirmHostKey).toHaveBeenCalledWith({ repository: 'ssh://h/r.git', fingerprint: 'SHA256:x', accept: true })
    vi.mocked(runtime.Clipboard.SetText).mockResolvedValueOnce(undefined as never)
    await service.copyPublicKey('ssh-ed25519 AAAA')
    expect(runtime.Clipboard.SetText).toHaveBeenCalledWith('ssh-ed25519 AAAA')
  })

  it('maps errors and refuses empty answers', async () => {
    vi.mocked(binding.StartGitHubSignIn).mockRejectedValueOnce(new AppError('ACC-01503', 'not configured'))
    await expect(service.startGitHubSignIn()).rejects.toMatchObject({ code: service.GITHUB_SIGN_IN_NOT_CONFIGURED })
    vi.mocked(binding.TestConnection).mockResolvedValueOnce(null as never)
    await expect(service.testConnection('a', 'r')).rejects.toMatchObject({ code: 'UNK-00500' })
  })
})
