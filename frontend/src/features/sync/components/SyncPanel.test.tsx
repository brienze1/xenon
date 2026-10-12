import { render, screen, within } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { beforeEach, describe, expect, it, vi } from 'vitest'
import { AppError } from '@/platform/errors'
import { pingChange, syncSession, syncStatus, usersChange } from '@/features/sync/test-fixtures'

vi.mock('@/platform/services/sync', () => ({
  SUBFOLDER_HAS_WORKSPACE: 'GIT-01409',
  getSyncStatus: vi.fn(),
  linkRepository: vi.fn(),
  listChanges: vi.fn(),
  commitChanges: vi.fn(),
  pushWorkspace: vi.fn(),
  pullWorkspace: vi.fn(),
  saveSyncSettings: vi.fn(),
  getSyncSession: vi.fn(),
  resolveConflict: vi.fn(),
  finishPull: vi.fn(),
  cancelPull: vi.fn(),
  unlinkRepository: vi.fn(),
}))
vi.mock('@/platform/services/git-accounts', () => ({ listGitAccounts: vi.fn() }))
vi.mock('@/platform/services/dialogs', () => ({ pickFolder: vi.fn() }))

const service = await import('@/platform/services/sync')
const accounts = await import('@/platform/services/git-accounts')
const dialogs = await import('@/platform/services/dialogs')
const { useSyncStore } = await import('@/features/sync/store')
const { useWorkspacesStore } = await import('@/features/workspaces/store')
const { SyncPanel } = await import('./SyncPanel')
const { FooterSyncStatus } = await import('./FooterSyncStatus')
const { LinkRepoDialog } = await import('./LinkRepoDialog')

const work = { id: 'a1', kind: 'token', provider: 'gitlab', sshMode: '', displayName: 'Work GitLab', host: 'gitlab.test', username: '', commitName: 'Luis', commitEmail: 'l@a.test', avatarUrl: '', publicKey: '', hasPassphrase: false } as const

describe('sync UI', () => {
  beforeEach(() => {
    vi.resetAllMocks()
    vi.mocked(accounts.listGitAccounts).mockResolvedValue([work])
    vi.mocked(service.listChanges).mockResolvedValue([pingChange, usersChange])
    useSyncStore.getState().reset('L')
  })

  it('shows the footer status and opens the panel from it', async () => {
    const { rerender } = render(<FooterSyncStatus />)
    expect(screen.queryByRole('button')).toBeNull()
    useSyncStore.setState({ status: syncStatus({ state: 'behind', behind: 2 }) })
    rerender(<FooterSyncStatus />)
    expect(screen.getByRole('button', { name: /Git: Behind \(2 to pull\)/ })).toBeInTheDocument()
    useSyncStore.setState({ busy: 'pull' })
    rerender(<FooterSyncStatus />)
    expect(screen.getByRole('button', { name: /Git: Syncing…/ })).toBeInTheDocument()
    useSyncStore.setState({ busy: null, status: syncStatus({ state: 'not_linked' }) })
    rerender(<FooterSyncStatus />)
    await userEvent.click(screen.getByRole('button', { name: /Git: Not synced/ }))
    expect(useSyncStore.getState().panelOpen).toBe(true)
  })

  it('lists the changes with checkboxes and commits and pushes the checked ones', async () => {
    useSyncStore.setState({ status: syncStatus({ state: 'local_changes', localChanges: 2 }) })
    useSyncStore.getState().openPanel()
    render(<SyncPanel />)
    const list = await screen.findByRole('list', { name: 'Changed items' })
    expect(within(list).getByText('Ping')).toBeInTheDocument()
    expect(within(list).getByText('Added')).toBeInTheDocument()
    const commitAndPush = screen.getByRole('button', { name: 'Commit and Push' })
    expect(commitAndPush).toBeDisabled()
    await userEvent.click(within(list).getAllByRole('checkbox')[1]!)
    await userEvent.type(screen.getByLabelText('Commit message'), 'Ping with HEAD')
    vi.mocked(service.commitChanges).mockResolvedValue(syncStatus({ state: 'ahead', ahead: 1 }))
    vi.mocked(service.pushWorkspace).mockResolvedValue(syncStatus())
    await userEvent.click(commitAndPush)
    expect(service.commitChanges).toHaveBeenCalledWith('L', 'Ping with HEAD', ['r1'])
    expect(service.pushWorkspace).toHaveBeenCalled()
    expect(await screen.findByText('Up to date')).toBeInTheDocument()
    expect(screen.getByText(/does not run Git hooks/)).toBeInTheDocument()
  })

  it('shows errors with their next step, notices, conflicts and the missing account', async () => {
    useSyncStore.setState({
      status: syncStatus({ state: 'error', notice: 'The shared sync settings changed', settings: { ...syncStatus().settings, accountId: '' }, lastError: { code: 'GIT-07428', message: 'No account - choose one', nextStep: 'Choose a Git account.' } }),
    })
    useSyncStore.getState().openPanel()
    render(<SyncPanel />)
    expect(await screen.findByText('No account - choose one. Choose a Git account.')).toBeInTheDocument()
    expect(screen.getByText('The shared sync settings changed')).toBeInTheDocument()
    await screen.findByRole('option', { name: 'Work GitLab' })
    vi.mocked(service.getSyncStatus).mockResolvedValue(syncStatus())
    await userEvent.selectOptions(screen.getByLabelText('Account'), 'a1')
    expect(service.saveSyncSettings).toHaveBeenCalledWith('L', 'a1', false, false)
    vi.mocked(service.pullWorkspace).mockRejectedValue(new AppError('GIT-10503', 'the Git host could not be reached', [{ attribute: 'next_step', messages: ['Offline - check the connection and try again'] }]))
    await userEvent.click(screen.getByRole('button', { name: 'Pull' }))
    expect(await screen.findByText('the Git host could not be reached. Offline - check the connection and try again')).toBeInTheDocument()
  })

  it('offers to resolve the conflicts of the last pull, also from the footer', async () => {
    useSyncStore.setState({ status: syncStatus({ state: 'conflict', ahead: 1, behind: 2 }) })
    useSyncStore.getState().openPanel()
    render(<SyncPanel />)
    expect(screen.getByText(/found items changed on both sides/)).toBeInTheDocument()
    vi.mocked(service.getSyncSession).mockResolvedValue(syncSession())
    await userEvent.click(screen.getByRole('button', { name: 'Resolve conflicts…' }))
    expect(useSyncStore.getState()).toMatchObject({ conflictsOpen: true, panelOpen: false })
    useSyncStore.setState({ conflictsOpen: false })
    render(<FooterSyncStatus />)
    await userEvent.click(screen.getByRole('button', { name: /Git: Conflict/ }))
    expect(service.getSyncSession).toHaveBeenCalledTimes(2)
    expect(useSyncStore.getState().conflictsOpen).toBe(true)
  })

  it('turns the automatic pull and push on and off', async () => {
    useSyncStore.setState({ status: syncStatus() })
    useSyncStore.getState().openPanel()
    vi.mocked(service.getSyncStatus).mockResolvedValue(syncStatus({ settings: { ...syncStatus().settings, autoPull: true } }))
    render(<SyncPanel />)
    const autoPull = screen.getByRole('checkbox', { name: /Pull when the workspace is opened/ })
    const autoPush = screen.getByRole('checkbox', { name: /Push after each save/ })
    expect(autoPull).not.toBeChecked()
    expect(autoPush).not.toBeChecked()
    await userEvent.click(autoPull)
    expect(service.saveSyncSettings).toHaveBeenLastCalledWith('L', 'a1', true, false)
    expect(await screen.findByRole('checkbox', { name: /Pull when the workspace is opened/ })).toBeChecked()
    await userEvent.click(autoPush)
    expect(service.saveSyncSettings).toHaveBeenLastCalledWith('L', 'a1', true, true)
  })

  it('unlinks the repository into a chosen folder after confirmation', async () => {
    const load = vi.spyOn(useWorkspacesStore.getState(), 'load').mockResolvedValue()
    const activate = vi.spyOn(useWorkspacesStore.getState(), 'activate').mockResolvedValue()
    useSyncStore.setState({ status: syncStatus() })
    useSyncStore.getState().openPanel()
    render(<SyncPanel />)
    vi.mocked(dialogs.pickFolder).mockResolvedValueOnce(null)
    await userEvent.click(screen.getByRole('button', { name: 'Unlink Repository…' }))
    expect(screen.queryByRole('button', { name: 'Unlink and Delete Clone' })).toBeNull()
    vi.mocked(dialogs.pickFolder).mockResolvedValue('/w/acme-plain')
    await userEvent.click(screen.getByRole('button', { name: 'Unlink Repository…' }))
    expect(screen.getByText(/delete Xenon's clone at/)).toHaveTextContent(syncStatus().settings.clonePath)
    await userEvent.click(screen.getByRole('button', { name: 'Cancel' }))
    expect(service.unlinkRepository).not.toHaveBeenCalled()
    await userEvent.click(screen.getByRole('button', { name: 'Unlink Repository…' }))
    vi.mocked(service.unlinkRepository).mockResolvedValue({ localId: 'L', name: 'Acme', path: '/w/acme-plain', missing: false })
    vi.mocked(service.getSyncStatus).mockResolvedValue(syncStatus({ state: 'not_linked' }))
    await userEvent.click(screen.getByRole('button', { name: 'Unlink and Delete Clone' }))
    expect(service.unlinkRepository).toHaveBeenCalledWith('L', '/w/acme-plain')
    expect(activate).toHaveBeenCalledWith('L')
    expect(await screen.findByText(/not linked to a Git repository/)).toBeInTheDocument()
    load.mockRestore()
    activate.mockRestore()
  })

  it('offers to link a workspace without a repository', async () => {
    useSyncStore.setState({ status: syncStatus({ state: 'not_linked' }) })
    useSyncStore.getState().openPanel()
    render(<SyncPanel />)
    expect(screen.getByText(/not linked to a Git repository/)).toBeInTheDocument()
    await userEvent.click(screen.getByRole('button', { name: 'Link repository…' }))
    expect(useSyncStore.getState().linkOpen).toBe(true)
    expect(service.listChanges).not.toHaveBeenCalled()
  })

  it('links a repository and asks for a choice when the subfolder holds a workspace', async () => {
    vi.spyOn(useWorkspacesStore.getState(), 'load').mockResolvedValue()
    vi.spyOn(useWorkspacesStore.getState(), 'activate').mockResolvedValue()
    vi.mocked(dialogs.pickFolder).mockResolvedValue('/clones/acme')
    useSyncStore.getState().openLink()
    render(<LinkRepoDialog />)
    const link = screen.getByRole('button', { name: 'Link' })
    await userEvent.type(screen.getByLabelText('Repository URL'), 'https://git.test/acme.git')
    await userEvent.type(screen.getByLabelText('Subfolder (optional)'), 'api/xenon')
    await screen.findByRole('option', { name: 'Work GitLab' })
    expect(link).toBeDisabled()
    await userEvent.selectOptions(screen.getByLabelText('Account'), 'a1')
    await userEvent.click(screen.getByRole('button', { name: 'Choose…' }))
    expect(screen.getByLabelText('Clone location')).toHaveValue('/clones/acme')
    vi.mocked(service.linkRepository).mockRejectedValueOnce(new AppError('GIT-01409', 'the subfolder already holds a Xenon workspace'))
    await userEvent.click(link)
    expect(await screen.findByText('The subfolder already holds a Xenon workspace.')).toBeInTheDocument()
    expect(link).toBeDisabled()
    await userEvent.click(screen.getByLabelText(/Use the repo's workspace/))
    vi.mocked(service.linkRepository).mockResolvedValue(syncStatus())
    await userEvent.click(link)
    expect(service.linkRepository).toHaveBeenLastCalledWith({
      localId: 'L',
      repository: 'https://git.test/acme.git',
      branch: 'main',
      subfolder: 'api/xenon',
      accountId: 'a1',
      clonePath: '/clones/acme',
      mode: 'adopt_repo',
    })
    await vi.waitFor(() => expect(useSyncStore.getState().linkOpen).toBe(false))
  })

  it('shows other link failures with their next step', async () => {
    useSyncStore.getState().openLink()
    render(<LinkRepoDialog />)
    await userEvent.type(screen.getByLabelText('Repository URL'), 'https://git.test/acme.git')
    await screen.findByRole('option', { name: 'Work GitLab' })
    await userEvent.selectOptions(screen.getByLabelText('Account'), 'a1')
    vi.mocked(service.linkRepository).mockRejectedValue(new AppError('GIT-02415', 'Large File Storage', [{ attribute: 'next_step', messages: ['Use another subfolder.'] }]))
    await userEvent.click(screen.getByRole('button', { name: 'Link' }))
    expect(await screen.findByText('Large File Storage. Use another subfolder.')).toBeInTheDocument()
    await userEvent.click(screen.getByRole('button', { name: 'Cancel' }))
    expect(useSyncStore.getState().linkOpen).toBe(false)
  })
})
