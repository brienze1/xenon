import { render, screen } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { beforeEach, describe, expect, it, vi } from 'vitest'
import { AppError } from '@/platform/errors'

vi.mock('@/platform/services/sync', () => ({ cloneAsWorkspace: vi.fn() }))
vi.mock('@/platform/services/git-accounts', () => ({ listGitAccounts: vi.fn() }))
vi.mock('@/platform/services/dialogs', () => ({ pickFolder: vi.fn() }))

const service = await import('@/platform/services/sync')
const accounts = await import('@/platform/services/git-accounts')
const dialogs = await import('@/platform/services/dialogs')
const { useSyncStore } = await import('@/features/sync/store')
const { useWorkspacesStore } = await import('@/features/workspaces/store')
const { CloneWorkspaceDialog } = await import('./CloneWorkspaceDialog')

const work = { id: 'a1', kind: 'token', provider: 'gitlab', sshMode: '', displayName: 'Work GitLab', host: 'gitlab.test', username: '', commitName: 'Luis', commitEmail: 'l@a.test', avatarUrl: '', publicKey: '', hasPassphrase: false } as const
const backend = { localId: 'B', name: 'Backend API', path: '/clones/backend/api/xenon', missing: false }

describe('CloneWorkspaceDialog', () => {
  beforeEach(() => {
    vi.resetAllMocks()
    vi.mocked(accounts.listGitAccounts).mockResolvedValue([work])
    useSyncStore.getState().reset('L')
    useSyncStore.setState({ cloneOpen: true })
  })

  it('clones the repository as a new workspace and switches to it', async () => {
    const load = vi.spyOn(useWorkspacesStore.getState(), 'load').mockResolvedValue()
    const activate = vi.spyOn(useWorkspacesStore.getState(), 'activate').mockResolvedValue()
    vi.mocked(dialogs.pickFolder).mockResolvedValue('/clones/backend')
    vi.mocked(service.cloneAsWorkspace).mockResolvedValue(backend)
    render(<CloneWorkspaceDialog />)
    const clone = screen.getByRole('button', { name: 'Clone and Open' })
    expect(clone).toBeDisabled()
    await userEvent.type(screen.getByLabelText('Repository URL'), ' https://gitlab.test/acme/backend.git ')
    await userEvent.type(screen.getByLabelText('Subfolder (optional)'), 'api/xenon')
    await userEvent.selectOptions(await screen.findByLabelText('Account'), 'a1')
    await userEvent.click(screen.getByRole('button', { name: 'Choose…' }))
    expect(screen.getByLabelText('Branch (optional)')).toHaveAttribute('placeholder', 'Remote default branch')
    await userEvent.click(clone)
    expect(service.cloneAsWorkspace).toHaveBeenCalledWith({
      repository: 'https://gitlab.test/acme/backend.git',
      branch: '',
      subfolder: 'api/xenon',
      accountId: 'a1',
      clonePath: '/clones/backend',
      name: '',
    })
    expect(load).toHaveBeenCalled()
    expect(activate).toHaveBeenCalledWith('B')
    expect(useSyncStore.getState().cloneOpen).toBe(false)
    load.mockRestore()
    activate.mockRestore()
  })

  it('shows the failure with its next step and stays open', async () => {
    vi.mocked(service.cloneAsWorkspace).mockRejectedValue(new AppError('GIT-03404', 'the branch release does not exist on the remote', [{ attribute: 'next_step', messages: ['Pick another branch'] }]))
    render(<CloneWorkspaceDialog />)
    await userEvent.type(screen.getByLabelText('Repository URL'), 'https://gitlab.test/acme/backend.git')
    await userEvent.type(screen.getByLabelText('Branch (optional)'), 'release')
    await userEvent.type(screen.getByLabelText('Workspace name (optional)'), 'Backend')
    await userEvent.selectOptions(await screen.findByLabelText('Account'), 'a1')
    await userEvent.click(screen.getByRole('button', { name: 'Clone and Open' }))
    expect(await screen.findByText(/does not exist on the remote/)).toBeInTheDocument()
    expect(service.cloneAsWorkspace).toHaveBeenCalledWith(expect.objectContaining({ branch: 'release', name: 'Backend', clonePath: '' }))
    expect(useSyncStore.getState().cloneOpen).toBe(true)
    await userEvent.click(screen.getByRole('button', { name: 'Cancel' }))
    expect(useSyncStore.getState().cloneOpen).toBe(false)
  })
})
