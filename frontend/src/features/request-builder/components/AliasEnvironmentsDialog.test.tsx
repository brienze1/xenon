import { render, screen, within } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest'
import { AppError } from '@/platform/errors'
import { emptyDraft, type SavedAlias, type SavedRequest } from '@/platform/mappers/request'

vi.mock('@/platform/services/requests', async (importOriginal) => {
  const { confirmationOf, expiredTokenOwnerOf, missingSecretsOf } = await importOriginal<typeof import('@/platform/services/requests')>()
  return { confirmationOf, expiredTokenOwnerOf, missingSecretsOf, getRequest: vi.fn(), autoHeaders: vi.fn(), sendRequest: vi.fn() }
})
vi.mock('@/platform/services/aliases', () => ({ createAlias: vi.fn(), saveAlias: vi.fn(), deleteAlias: vi.fn(), setAliasEnvironments: vi.fn() }))
vi.mock('@/platform/services/variables', () => ({ inspectVariables: vi.fn(async () => ({ references: [], unresolved: [] })), suggestVariables: vi.fn() }))

const requests = await import('@/platform/services/requests')
const aliasService = await import('@/platform/services/aliases')
const { useRequestBuilderStore } = await import('@/features/request-builder/store')
const { useWorkspacesStore } = await import('@/features/workspaces/store')
const { useEnvironmentsStore } = await import('@/features/environments/store')
const { AliasEnvironmentsDialog } = await import('./AliasEnvironmentsDialog')

const TAB = 'tab-1'
const environments = [
  { id: 'hml', name: 'HML', color: 'green' as const, confirmBeforeSend: false, error: null },
  { id: 'dev', name: 'DEV', color: 'cyan' as const, confirmBeforeSend: false, error: null },
  { id: 'prod', name: 'Prod', color: 'red' as const, confirmBeforeSend: true, error: null },
  { id: 'error:broken.yaml', name: 'broken.yaml', color: 'none' as const, confirmBeforeSend: false, error: 'unreadable' },
]

function aliasOf(environmentIds: string[]): SavedAlias {
  return { id: 'a1', name: 'mx-test-user', environmentIds, draft: { ...emptyDraft(), id: 'r1', name: 'mx-test-user' } }
}

const SAVED: SavedRequest = { draft: { ...emptyDraft(), id: 'r1', name: 'Login' }, revision: 'rev-1', aliases: [aliasOf([])] }

function renderDialog(alias: SavedAlias, onClose = vi.fn()) {
  render(<AliasEnvironmentsDialog tabId={TAB} alias={alias} onClose={onClose} />)
  return { dialog: screen.getByRole('dialog', { name: 'Environments for mx-test-user' }), onClose }
}

describe('AliasEnvironmentsDialog', () => {
  beforeEach(async () => {
    vi.resetAllMocks()
    useRequestBuilderStore.getState().reset()
    useWorkspacesStore.setState({ activeLocalId: 'ws' })
    useEnvironmentsStore.setState({ environments, activeId: 'hml' })
    vi.mocked(requests.autoHeaders).mockResolvedValue([])
    vi.mocked(requests.getRequest).mockResolvedValue(SAVED)
    await useRequestBuilderStore.getState().openSaved(TAB, 'r1')
  })

  afterEach(() => useEnvironmentsStore.setState({ environments: [], activeId: null }))

  it('opens on All environments for an alias without a link', () => {
    const { dialog } = renderDialog(aliasOf([]))
    expect(within(dialog).getByRole('radio', { name: 'All environments' })).toBeChecked()
    expect(within(dialog).queryByRole('checkbox')).not.toBeInTheDocument()
  })

  it('opens on the alias’s environments, lists readable environments in order and saves the ticked ones', async () => {
    vi.mocked(aliasService.setAliasEnvironments).mockResolvedValue({ ...SAVED, revision: 'rev-2', aliases: [aliasOf(['hml', 'prod'])] })
    const user = userEvent.setup()
    const { dialog, onClose } = renderDialog(aliasOf(['hml']))
    expect(within(dialog).getByRole('radio', { name: 'Only these environments' })).toBeChecked()
    expect(within(dialog).getAllByRole('checkbox').map((box) => box.closest('label')?.textContent)).toEqual(['HML', 'DEV', 'Prod'])
    expect(within(dialog).getByRole('checkbox', { name: 'HML' })).toBeChecked()
    await user.click(within(dialog).getByRole('checkbox', { name: 'Prod' }))
    await user.click(within(dialog).getByRole('button', { name: 'Save' }))
    expect(aliasService.setAliasEnvironments).toHaveBeenCalledWith('ws', 'r1', 'a1', ['hml', 'prod'], 'rev-1')
    expect(onClose).toHaveBeenCalled()
  })

  it('needs at least one environment under Only these environments', async () => {
    const user = userEvent.setup()
    const { dialog } = renderDialog(aliasOf(['hml']))
    await user.click(within(dialog).getByRole('checkbox', { name: 'HML' }))
    expect(within(dialog).getByText('Pick at least one environment.')).toBeInTheDocument()
    expect(within(dialog).getByRole('button', { name: 'Save' })).toBeDisabled()
    await user.click(within(dialog).getByRole('radio', { name: 'All environments' }))
    expect(within(dialog).getByRole('button', { name: 'Save' })).toBeEnabled()
  })

  it('flags a deleted environment as removed on save and never submits it', async () => {
    vi.mocked(aliasService.setAliasEnvironments).mockResolvedValue({ ...SAVED, revision: 'rev-2', aliases: [aliasOf(['dev'])] })
    const user = userEvent.setup()
    const { dialog } = renderDialog(aliasOf(['gone', 'dev']))
    expect(within(dialog).getByText('Deleted environment · Removed when you save')).toBeInTheDocument()
    await user.click(within(dialog).getByRole('button', { name: 'Save' }))
    expect(aliasService.setAliasEnvironments).toHaveBeenCalledWith('ws', 'r1', 'a1', ['dev'], 'rev-1')
  })

  it('saves All environments as no environments', async () => {
    vi.mocked(aliasService.setAliasEnvironments).mockResolvedValue(SAVED)
    const user = userEvent.setup()
    const { dialog } = renderDialog(aliasOf(['hml']))
    await user.click(within(dialog).getByRole('radio', { name: 'All environments' }))
    await user.click(within(dialog).getByRole('button', { name: 'Save' }))
    expect(aliasService.setAliasEnvironments).toHaveBeenCalledWith('ws', 'r1', 'a1', [], 'rev-1')
  })

  it('shows Go’s refusal inline and stays open; Cancel writes nothing', async () => {
    vi.mocked(aliasService.setAliasEnvironments).mockRejectedValue(new AppError('ALS-10400', 'the environment does not exist in this workspace'))
    const user = userEvent.setup()
    const { dialog, onClose } = renderDialog(aliasOf(['hml']))
    await user.click(within(dialog).getByRole('button', { name: 'Save' }))
    expect(await within(dialog).findByText('the environment does not exist in this workspace')).toBeInTheDocument()
    expect(onClose).not.toHaveBeenCalled()
    vi.mocked(aliasService.setAliasEnvironments).mockClear()
    await user.click(within(dialog).getByRole('button', { name: 'Cancel' }))
    expect(aliasService.setAliasEnvironments).not.toHaveBeenCalled()
    expect(onClose).toHaveBeenCalled()
  })
})
