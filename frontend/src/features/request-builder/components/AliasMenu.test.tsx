import { render, screen, within } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest'
import { AppError } from '@/platform/errors'
import { emptyDraft, type RequestDraft, type SavedRequest } from '@/platform/mappers/request'

vi.mock('@/platform/services/requests', async (importOriginal) => {
  const { confirmationOf, expiredTokenOwnerOf, missingSecretsOf } = await importOriginal<typeof import('@/platform/services/requests')>()
  return { confirmationOf, expiredTokenOwnerOf, missingSecretsOf, getRequest: vi.fn(), autoHeaders: vi.fn(), sendRequest: vi.fn() }
})
vi.mock('@/platform/services/aliases', () => ({ createAlias: vi.fn(), saveAlias: vi.fn(), deleteAlias: vi.fn() }))
vi.mock('@/platform/services/variables', () => ({ inspectVariables: vi.fn(async () => ({ references: [], unresolved: [] })), suggestVariables: vi.fn() }))

const requests = await import('@/platform/services/requests')
const aliasService = await import('@/platform/services/aliases')
const { useRequestBuilderStore } = await import('@/features/request-builder/store')
const { useTabsStore } = await import('@/features/tabs/store')
const { useWorkspacesStore } = await import('@/features/workspaces/store')
const { AliasMenu } = await import('./AliasMenu')

const TAB = 'tab-1'

function draftOf(name: string, url: string): RequestDraft {
  return { ...emptyDraft(), id: 'r1', name, url }
}

const SAVED: SavedRequest = {
  draft: draftOf('Login', 'https://x/v1'),
  revision: 'rev-1',
  aliases: [
    { id: 'a1', name: 'test-user', draft: draftOf('test-user', 'https://x/v2') },
    { id: 'a2', name: 'mx-test-user', draft: draftOf('mx-test-user', 'https://x/v3') },
  ],
}

function tab() {
  return useRequestBuilderStore.getState().tabs[TAB]!
}

async function openMenu(user: ReturnType<typeof userEvent.setup>) {
  await user.click(screen.getByRole('button', { name: /^Aliases:/ }))
  return screen.getByRole('menu', { name: 'Aliases' })
}

describe('AliasMenu', () => {
  beforeEach(async () => {
    vi.resetAllMocks()
    useRequestBuilderStore.getState().reset()
    useWorkspacesStore.setState({ activeLocalId: 'ws' })
    vi.mocked(requests.autoHeaders).mockResolvedValue([])
    vi.mocked(requests.getRequest).mockResolvedValue(SAVED)
    useTabsStore.setState({ tabs: [{ id: TAB, requestId: 'r1', pinned: false }], previewTabId: TAB, activeTabId: TAB })
    await useRequestBuilderStore.getState().openSaved(TAB, 'r1')
  })

  afterEach(() => useWorkspacesStore.setState({ activeLocalId: null }))

  it('lists Default, the aliases in file order, then Create new alias…', async () => {
    const user = userEvent.setup()
    render(<AliasMenu tabId={TAB} />)
    expect(screen.getByRole('button', { name: 'Aliases: Default' })).toBeInTheDocument()
    const menu = await openMenu(user)
    const radios = within(menu).getAllByRole('menuitemradio')
    expect(radios.map((item) => item.textContent)).toEqual(['Default', 'test-user', 'mx-test-user'])
    expect(radios[0]).toHaveAttribute('aria-checked', 'true')
    const items = within(menu).getAllByRole('menuitem')
    expect(items.at(-1)).toHaveTextContent('Create new alias…')
    expect(within(menu).getByRole('menuitem', { name: 'Delete alias test-user…' })).toBeInTheDocument()
    expect(within(menu).queryByRole('menuitem', { name: /Delete alias Default/ })).not.toBeInTheDocument()
  })

  it('switches the tab to the picked alias and pins a preview tab', async () => {
    const user = userEvent.setup()
    render(<AliasMenu tabId={TAB} />)
    await user.click(within(await openMenu(user)).getByRole('menuitemradio', { name: 'mx-test-user' }))
    expect(tab().aliasId).toBe('a2')
    expect(tab().draft.url).toBe('https://x/v3')
    expect(screen.getByRole('button', { name: 'Aliases: mx-test-user' })).toBeInTheDocument()
    expect(useTabsStore.getState().previewTabId).toBeNull()
  })

  it('creates an alias by name and shows Go’s refusal inline without closing', async () => {
    const user = userEvent.setup()
    render(<AliasMenu tabId={TAB} />)
    await user.click(within(await openMenu(user)).getByRole('menuitem', { name: 'Create new alias…' }))
    const dialog = screen.getByRole('dialog', { name: 'Create New Alias' })
    await user.click(within(dialog).getByRole('button', { name: 'Create Alias' }))
    expect(within(dialog).getByRole('alert')).toHaveTextContent('Enter a name for the alias.')
    vi.mocked(aliasService.createAlias).mockRejectedValueOnce(new AppError('ALS-04400', '"Default" is reserved; choose another alias name'))
    await user.type(within(dialog).getByRole('textbox', { name: 'Alias name' }), 'Default')
    await user.click(within(dialog).getByRole('button', { name: 'Create Alias' }))
    expect(await within(dialog).findByRole('alert')).toHaveTextContent('"Default" is reserved')
    expect(screen.getByRole('dialog', { name: 'Create New Alias' })).toBeInTheDocument()
    vi.mocked(aliasService.createAlias).mockResolvedValueOnce({ ...SAVED, revision: 'rev-2', aliases: [...SAVED.aliases, { id: 'a3', name: 'canary', draft: draftOf('canary', 'https://x/v1') }] })
    await user.clear(within(dialog).getByRole('textbox', { name: 'Alias name' }))
    await user.type(within(dialog).getByRole('textbox', { name: 'Alias name' }), 'canary')
    await user.click(within(dialog).getByRole('button', { name: 'Create Alias' }))
    expect(aliasService.createAlias).toHaveBeenLastCalledWith('ws', 'r1', 'canary', 'rev-1')
    expect(await screen.findByRole('button', { name: 'Aliases: canary' })).toBeInTheDocument()
    expect(screen.queryByRole('dialog')).not.toBeInTheDocument()
  })

  it('deletes an alias after confirming and goes back to the Default when it was shown', async () => {
    const user = userEvent.setup()
    useRequestBuilderStore.getState().selectVersion(TAB, 'a1')
    render(<AliasMenu tabId={TAB} />)
    await user.click(within(await openMenu(user)).getByRole('menuitem', { name: 'Delete alias test-user…' }))
    const dialog = screen.getByRole('dialog', { name: 'Delete alias?' })
    expect(dialog).toHaveTextContent('“test-user” will be removed from Login.')
    await user.click(within(dialog).getByRole('button', { name: 'Cancel' }))
    expect(aliasService.deleteAlias).not.toHaveBeenCalled()
    vi.mocked(aliasService.deleteAlias).mockResolvedValueOnce({ ...SAVED, revision: 'rev-2', aliases: [SAVED.aliases[1]!] })
    await user.click(within(await openMenu(user)).getByRole('menuitem', { name: 'Delete alias test-user…' }))
    await user.click(within(screen.getByRole('dialog', { name: 'Delete alias?' })).getByRole('button', { name: 'Delete' }))
    expect(aliasService.deleteAlias).toHaveBeenCalledWith('ws', 'r1', 'a1', 'rev-1')
    expect(await screen.findByRole('button', { name: 'Aliases: Default' })).toBeInTheDocument()
    expect(tab().aliasId).toBeNull()
  })
})
