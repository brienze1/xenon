import { render, screen, within } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { beforeEach, describe, expect, it, vi } from 'vitest'
import { AppError } from '@/platform/errors'
import { environmentConflict, fieldConflict, requestConflict, syncSession, syncStatus } from '@/features/sync/test-fixtures'

vi.mock('@/platform/services/sync', () => ({
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
}))

const service = await import('@/platform/services/sync')
const { useSyncStore } = await import('@/features/sync/store')
const { ConflictView } = await import('./ConflictView')

function openWith(conflicts = [requestConflict(), fieldConflict, environmentConflict]) {
  useSyncStore.setState({ session: syncSession(conflicts), selectedConflictId: conflicts[0]?.itemId ?? null, conflictsOpen: true })
}

describe('ConflictView', () => {
  beforeEach(() => {
    vi.resetAllMocks()
    vi.mocked(service.listChanges).mockResolvedValue([])
    useSyncStore.getState().reset('L')
  })

  it('renders nothing without a session', () => {
    render(<ConflictView />)
    expect(screen.queryByRole('dialog')).toBeNull()
  })

  it('lists the conflicts and shows mine and theirs side by side as requests', () => {
    openWith()
    render(<ConflictView />)
    const list = screen.getByRole('list', { name: 'Conflicts' })
    expect(within(list).getAllByRole('button')).toHaveLength(3)
    expect(within(list).getByRole('button', { name: /Users/ })).toHaveAttribute('aria-current', 'true')
    expect(screen.getByText('0 of 3 resolved', { exact: false })).toBeInTheDocument()
    const mine = screen.getByRole('region', { name: 'Mine' })
    const theirs = screen.getByRole('region', { name: 'Theirs' })
    expect(within(mine).getByText('PATCH')).toBeInTheDocument()
    expect(within(theirs).getByText('POST')).toBeInTheDocument()
    expect(within(theirs).getByText('https://a.test/users?v=2')).toBeInTheDocument()
    expect(within(mine).getByRole('table', { name: 'Headers' })).toBeInTheDocument()
    expect(screen.getByText('Common base')).toBeInTheDocument()
    const choices = screen.getByRole('group', { name: 'Resolution' })
    expect(within(choices).getAllByRole('button').map((button) => button.textContent)).toEqual(['Keep mine', 'Keep theirs', 'Keep both'])
    expect(screen.getByRole('button', { name: 'Finish' })).toBeDisabled()
  })

  it('resolves each conflict in any order and finishes once all are resolved', async () => {
    openWith([requestConflict(), fieldConflict])
    render(<ConflictView />)
    vi.mocked(service.resolveConflict).mockResolvedValueOnce(requestConflict({ resolution: 'both' }))
    await userEvent.click(screen.getByRole('button', { name: 'Keep both' }))
    expect(service.resolveConflict).toHaveBeenCalledWith('s1', 'r1', 'both', '')
    expect(await screen.findByText(/saved next to yours as “Users \(theirs\)”/)).toBeInTheDocument()
    expect(screen.getByRole('button', { name: 'Keep both' })).toHaveAttribute('aria-pressed', 'true')
    await userEvent.click(screen.getByRole('button', { name: /API/ }))
    const fields = screen.getByRole('list', { name: 'Conflicting fields' })
    expect(within(fields).getByText('description')).toBeInTheDocument()
    expect(within(fields).getByText('(removed)')).toBeInTheDocument()
    vi.mocked(service.resolveConflict).mockResolvedValueOnce({ ...fieldConflict, fields: [{ ...fieldConflict.fields[0]!, resolution: 'theirs' }, fieldConflict.fields[1]!] })
    await userEvent.click(within(screen.getByRole('group', { name: 'Resolution for description' })).getByRole('button', { name: 'Keep theirs' }))
    expect(service.resolveConflict).toHaveBeenLastCalledWith('s1', 'c1', 'theirs', 'description')
    expect(screen.getByRole('button', { name: 'Finish' })).toBeDisabled()
    vi.mocked(service.resolveConflict).mockResolvedValueOnce({ ...fieldConflict, resolution: 'merged' })
    await userEvent.click(within(screen.getByRole('group', { name: 'Resolution for auth.type' })).getByRole('button', { name: 'Keep mine' }))
    expect(await screen.findByText('2 of 2 resolved', { exact: false })).toBeInTheDocument()
    vi.mocked(service.finishPull).mockResolvedValue(syncStatus({ state: 'ahead', ahead: 1 }))
    await userEvent.click(screen.getByRole('button', { name: 'Finish' }))
    expect(service.finishPull).toHaveBeenCalledWith('s1')
    expect(screen.queryByRole('dialog')).toBeNull()
    expect(useSyncStore.getState()).toMatchObject({ panelOpen: true, status: { state: 'ahead' } })
  })

  it('shows deletions and environments without secret values, and Keep or Delete', async () => {
    openWith([environmentConflict])
    render(<ConflictView />)
    expect(within(screen.getByRole('region', { name: 'Mine' })).getByText('Deleted on this side')).toBeInTheDocument()
    const theirs = screen.getByRole('region', { name: 'Theirs' })
    expect(within(theirs).getByText('https://theirs.test')).toBeInTheDocument()
    expect(within(theirs).getByText(/secret - value stays on each machine/)).toBeInTheDocument()
    expect(screen.getByText(/mine is your uncommitted edit/)).toBeInTheDocument()
    expect(screen.queryByText('Common base')).toBeNull()
    expect(within(screen.getByRole('group', { name: 'Resolution' })).getAllByRole('button').map((button) => button.textContent)).toEqual(['Keep', 'Delete'])
    vi.mocked(service.resolveConflict).mockRejectedValue(new AppError('GIT-16404', 'the pull is not open any more'))
    await userEvent.click(screen.getByRole('button', { name: 'Delete' }))
    expect(await screen.findByText(/the pull is not open any more/)).toBeInTheDocument()
  })

  it('shows structural files as text and cancels back to the panel', async () => {
    openWith([requestConflict({ itemKind: 'collection', itemName: 'API', base: null, mine: { text: 'name: Mine', request: null, environment: null } })])
    render(<ConflictView />)
    expect(within(screen.getByRole('region', { name: 'Mine' })).getByText('name: Mine')).toBeInTheDocument()
    expect(within(screen.getByRole('group', { name: 'Resolution' })).getAllByRole('button').map((button) => button.textContent)).toEqual(['Keep mine', 'Keep theirs'])
    vi.mocked(service.cancelPull).mockResolvedValue(syncStatus({ state: 'diverged' }))
    await userEvent.click(screen.getByRole('button', { name: 'Cancel pull' }))
    expect(service.cancelPull).toHaveBeenCalledWith('s1')
    expect(screen.queryByRole('dialog')).toBeNull()
    expect(useSyncStore.getState()).toMatchObject({ session: null, panelOpen: true })
  })

  it('closes with Escape and keeps the session for later', async () => {
    openWith()
    render(<ConflictView />)
    await userEvent.keyboard('{Escape}')
    expect(screen.queryByRole('dialog')).toBeNull()
    expect(useSyncStore.getState().session?.id).toBe('s1')
  })
})
