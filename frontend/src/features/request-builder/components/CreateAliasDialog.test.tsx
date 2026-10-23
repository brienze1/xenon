import { render, screen, within } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest'
import { emptyDraft, type SavedRequest } from '@/platform/mappers/request'

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
const { useTabsStore } = await import('@/features/tabs/store')
const { CreateAliasDialog } = await import('./CreateAliasDialog')
const { UnavailableAliasBanner } = await import('./UnavailableAliasBanner')

const TAB = 'tab-1'
const environments = [
  { id: 'hml', name: 'HML', color: 'green' as const, confirmBeforeSend: false, error: null },
  { id: 'prod', name: 'Prod', color: 'red' as const, confirmBeforeSend: true, error: null },
]
const SAVED: SavedRequest = { draft: { ...emptyDraft(), id: 'r1', name: 'Login' }, revision: 'rev-1', aliases: [] }

describe('CreateAliasDialog environments', () => {
  beforeEach(async () => {
    vi.resetAllMocks()
    useRequestBuilderStore.getState().reset()
    useWorkspacesStore.setState({ activeLocalId: 'ws' })
    useEnvironmentsStore.setState({ environments, activeId: 'prod' })
    useTabsStore.setState({ tabs: [{ id: TAB, requestId: 'r1', pinned: false }], previewTabId: TAB, activeTabId: TAB })
    vi.mocked(requests.autoHeaders).mockResolvedValue([])
    vi.mocked(requests.getRequest).mockResolvedValue(SAVED)
    await useRequestBuilderStore.getState().openSaved(TAB, 'r1')
  })

  afterEach(() => useEnvironmentsStore.setState({ environments: [], activeId: null }))

  it('starts on All environments and creates the alias for all environments by default', async () => {
    vi.mocked(aliasService.createAlias).mockResolvedValue({ ...SAVED, revision: 'rev-2', aliases: [{ id: 'a1', name: 'v2', environmentIds: [], draft: SAVED.draft }] })
    const user = userEvent.setup()
    render(<CreateAliasDialog tabId={TAB} open onClose={vi.fn()} />)
    const dialog = screen.getByRole('dialog', { name: 'Create New Alias' })
    expect(within(dialog).getByRole('radio', { name: 'All environments' })).toBeChecked()
    await user.type(within(dialog).getByRole('textbox', { name: 'Alias name' }), 'v2')
    await user.click(within(dialog).getByRole('button', { name: 'Create Alias' }))
    expect(aliasService.createAlias).toHaveBeenCalledWith('ws', 'r1', 'v2', [], 'rev-1')
  })

  it('needs an environment ticked under Only these environments', async () => {
    const user = userEvent.setup()
    render(<CreateAliasDialog tabId={TAB} open onClose={vi.fn()} />)
    const dialog = screen.getByRole('dialog', { name: 'Create New Alias' })
    await user.type(within(dialog).getByRole('textbox', { name: 'Alias name' }), 'canary')
    await user.click(within(dialog).getByRole('radio', { name: 'Only these environments' }))
    expect(within(dialog).getByRole('button', { name: 'Create Alias' })).toBeDisabled()
  })

  it('creates an alias for HML while Prod is active and shows it with Send blocked', async () => {
    vi.mocked(aliasService.createAlias).mockResolvedValue({ ...SAVED, revision: 'rev-2', aliases: [{ id: 'a1', name: 'canary', environmentIds: ['hml'], draft: { ...SAVED.draft, name: 'canary' } }] })
    const user = userEvent.setup()
    render(
      <>
        <UnavailableAliasBanner tabId={TAB} />
        <CreateAliasDialog tabId={TAB} open onClose={vi.fn()} />
      </>,
    )
    const dialog = screen.getByRole('dialog', { name: 'Create New Alias' })
    await user.type(within(dialog).getByRole('textbox', { name: 'Alias name' }), 'canary')
    await user.click(within(dialog).getByRole('radio', { name: 'Only these environments' }))
    await user.click(within(dialog).getByRole('checkbox', { name: 'HML' }))
    await user.click(within(dialog).getByRole('button', { name: 'Create Alias' }))
    expect(aliasService.createAlias).toHaveBeenCalledWith('ws', 'r1', 'canary', ['hml'], 'rev-1')
    expect(useRequestBuilderStore.getState().tabs[TAB]?.aliasId).toBe('a1')
    expect(await screen.findByRole('status')).toHaveTextContent('canary is only for HML.')
  })
})
