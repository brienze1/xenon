import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest'
import { AppError } from '@/platform/errors'
import { emptyDraft, type RequestDraft, type SavedRequest } from '@/platform/mappers/request'

vi.mock('@/platform/services/requests', async (importOriginal) => {
  const { confirmationOf, expiredTokenOwnerOf, missingSecretsOf } = await importOriginal<typeof import('@/platform/services/requests')>()
  return {
    confirmationOf,
    expiredTokenOwnerOf,
    missingSecretsOf,
    applyUrl: vi.fn(),
    applyParams: vi.fn(),
    parseBulk: vi.fn(),
    parseCurl: vi.fn(),
    formatBulk: vi.fn(),
    autoHeaders: vi.fn(),
    sendRequest: vi.fn(),
    getRequest: vi.fn(),
    saveRequest: vi.fn(),
    saveRequestAs: vi.fn(),
    lastResult: vi.fn(),
    releaseResponse: vi.fn(),
    saveResponseBody: vi.fn(),
  }
})
vi.mock('@/platform/services/aliases', () => ({ createAlias: vi.fn(), saveAlias: vi.fn(), deleteAlias: vi.fn() }))
vi.mock('@/platform/services/dialogs', () => ({ pickFolder: vi.fn(), pickFile: vi.fn(), pickSaveFile: vi.fn() }))
vi.mock('@/platform/services/workspaces', () => ({ listWorkspaces: vi.fn(), createWorkspace: vi.fn(), activateWorkspace: vi.fn() }))
vi.mock('@/platform/services/tree', () => ({ getTree: vi.fn() }))
vi.mock('@/platform/services/variables', () => ({ inspectVariables: vi.fn(), suggestVariables: vi.fn() }))

const requests = await import('@/platform/services/requests')
const aliasService = await import('@/platform/services/aliases')
const variables = await import('@/platform/services/variables')
const { useWorkspacesStore } = await import('@/features/workspaces/store')
const { selectDirtyHiddenVersions, useRequestBuilderStore } = await import('./store')

const TAB = 'tab-1'

function draftOf(name: string, url: string): RequestDraft {
  return { ...emptyDraft(), id: 'r1', name, url }
}

function saved(revision: string, aliasUrl = 'https://x/v2', extra: SavedRequest['aliases'] = []): SavedRequest {
  return {
    draft: draftOf('Login', 'https://x/v1'),
    revision,
    aliases: [{ id: 'a1', name: 'mx', environmentIds: [], draft: draftOf('mx', aliasUrl) }, ...extra],
  }
}

function state() {
  return useRequestBuilderStore.getState()
}

function tab() {
  return state().tabs[TAB]!
}

describe('request builder - aliases', () => {
  beforeEach(() => {
    vi.resetAllMocks()
    state().reset()
    vi.mocked(requests.autoHeaders).mockResolvedValue([])
    vi.mocked(variables.inspectVariables).mockResolvedValue({ references: [], unresolved: [] } as never)
    useWorkspacesStore.setState({ activeLocalId: 'ws' })
    vi.mocked(requests.getRequest).mockResolvedValue(saved('rev-1'))
  })

  afterEach(() => state().reset())

  it('opens on the Default and lists the aliases; opening on an alias shows it', async () => {
    await state().openSaved(TAB, 'r1')
    expect(tab().aliasId).toBeNull()
    expect(tab().aliases.map((alias) => alias.name)).toEqual(['mx'])
    expect(tab().draft.url).toBe('https://x/v1')
    await state().openSaved('tab-2', 'r1', 'a1')
    expect(state().tabs['tab-2']!.aliasId).toBe('a1')
    expect(state().tabs['tab-2']!.draft.url).toBe('https://x/v2')
    expect(state().tabs['tab-2']!.savedName).toBe('mx')
  })

  it('switches versions without prompting and keeps each version’s unsaved edits', async () => {
    await state().openSaved(TAB, 'r1')
    state().setUrl(TAB, 'https://x/v1?edited')
    state().selectVersion(TAB, 'a1')
    expect(tab().draft.url).toBe('https://x/v2')
    expect(tab().dirty).toBe(false)
    expect(selectDirtyHiddenVersions(state(), TAB)).toEqual([null])
    state().setUrl(TAB, 'https://x/v2?edited')
    state().selectVersion(TAB, null)
    expect(tab().draft.url).toBe('https://x/v1?edited')
    expect(tab().dirty).toBe(true)
    expect(selectDirtyHiddenVersions(state(), TAB)).toEqual(['a1'])
  })

  it('saves only the alias on screen, with its name, and keeps the other versions aside', async () => {
    await state().openSaved(TAB, 'r1')
    state().setUrl(TAB, 'https://x/v1?edited')
    state().selectVersion(TAB, 'a1')
    state().setUrl(TAB, 'https://x/v2?edited')
    state().editName(TAB, 'mx-renamed')
    vi.mocked(aliasService.saveAlias).mockResolvedValue({
      ...saved('rev-2'),
      aliases: [{ id: 'a1', name: 'mx-renamed', environmentIds: [], draft: draftOf('mx-renamed', 'https://x/v2?edited') }],
    })
    const result = await state().save(TAB)
    expect(result).toEqual({ ok: true, value: true })
    expect(aliasService.saveAlias).toHaveBeenCalledWith('ws', 'a1', expect.objectContaining({ id: 'r1', name: 'mx-renamed', url: 'https://x/v2?edited' }), 'rev-1', false)
    expect(requests.saveRequest).not.toHaveBeenCalled()
    expect(tab().dirty).toBe(false)
    expect(tab().savedName).toBe('mx-renamed')
    expect(tab().baseRevision).toBe('rev-2')
    expect(selectDirtyHiddenVersions(state(), TAB)).toEqual([null])
  })

  it('reports a stale alias save as a conflict', async () => {
    await state().openSaved(TAB, 'r1', 'a1')
    vi.mocked(aliasService.saveAlias).mockRejectedValue(new AppError('REQ-02409', 'changed on disk'))
    const result = await state().save(TAB)
    expect(result.ok).toBe(false)
    expect(tab().conflict).toBe(true)
  })

  it('creates an alias from the saved Default and shows it', async () => {
    await state().openSaved(TAB, 'r1')
    state().setUrl(TAB, 'https://x/v1?unsaved')
    vi.mocked(aliasService.createAlias).mockResolvedValue(saved('rev-2', 'https://x/v2', [{ id: 'a2', name: 'canary', environmentIds: [], draft: draftOf('canary', 'https://x/v1') }]))
    const failure = await state().createAlias(TAB, 'canary')
    expect(failure).toBeNull()
    expect(aliasService.createAlias).toHaveBeenCalledWith('ws', 'r1', 'canary', [], 'rev-1')
    expect(tab().aliasId).toBe('a2')
    expect(tab().draft.url).toBe('https://x/v1')
    expect(tab().baseRevision).toBe('rev-2')
    expect(tab().versions['']!.draft.url).toBe('https://x/v1?unsaved')
  })

  it('returns why an alias could not be created and stays on the version shown', async () => {
    await state().openSaved(TAB, 'r1')
    vi.mocked(aliasService.createAlias).mockRejectedValue(new AppError('ALS-05409', 'taken'))
    const failure = await state().createAlias(TAB, 'mx')
    expect(failure?.code).toBe('ALS-05409')
    expect(tab().aliasId).toBeNull()
  })

  it('deleting the alias on screen drops its edits and goes back to the Default', async () => {
    await state().openSaved(TAB, 'r1')
    state().setUrl(TAB, 'https://x/v1?edited')
    state().selectVersion(TAB, 'a1')
    state().setUrl(TAB, 'https://x/v2?edited')
    vi.mocked(aliasService.deleteAlias).mockResolvedValue({ draft: draftOf('Login', 'https://x/v1'), revision: 'rev-2', aliases: [] })
    const failure = await state().deleteAlias(TAB, 'a1')
    expect(failure).toBeNull()
    expect(aliasService.deleteAlias).toHaveBeenCalledWith('ws', 'r1', 'a1', 'rev-1')
    expect(tab().aliasId).toBeNull()
    expect(tab().draft.url).toBe('https://x/v1?edited')
    expect(tab().aliases).toEqual([])
    expect(tab().versions).toEqual({})
  })

  it('sends the version on screen with its alias', async () => {
    vi.mocked(requests.sendRequest).mockReturnValue({ result: new Promise(() => {}), cancel: vi.fn() })
    await state().openSaved(TAB, 'r1', 'a1')
    void state().send(TAB)
    expect(requests.sendRequest).toHaveBeenCalledWith('ws', TAB, expect.objectContaining({ url: 'https://x/v2' }), expect.any(Object), 'a1')
  })

  it('reloading from disk replaces only the version on screen and keeps the others’ edits', async () => {
    await state().openSaved(TAB, 'r1')
    state().setUrl(TAB, 'https://x/v1?edited')
    state().selectVersion(TAB, 'a1')
    state().setUrl(TAB, 'https://x/v2?mine')
    vi.mocked(requests.getRequest).mockResolvedValue({ ...saved('rev-3', 'https://x/v2?theirs') })
    await state().openSaved(TAB, 'r1')
    expect(tab().aliasId).toBe('a1')
    expect(tab().draft.url).toBe('https://x/v2?theirs')
    expect(tab().baseRevision).toBe('rev-3')
    expect(tab().versions['']!.draft.url).toBe('https://x/v1?edited')
  })

  it('falls back to the Default when a restored alias no longer exists', async () => {
    state().loadDraft(TAB, draftOf('gone', 'https://x/gone'), 'rev-0', true)
    vi.mocked(requests.getRequest).mockResolvedValue({ draft: draftOf('Login', 'https://x/v1'), revision: 'rev-1', aliases: [] })
    await state().restoreVersions(TAB, 'missing', { missing: draftOf('gone', 'https://x/gone') })
    expect(tab().aliasId).toBeNull()
    expect(tab().draft.url).toBe('https://x/v1')
    expect(tab().versions).toEqual({})
  })

  it('a duplicated tab keeps the alias and the versions of the tab it copies', async () => {
    await state().openSaved(TAB, 'r1')
    state().setUrl(TAB, 'https://x/v1?edited')
    state().selectVersion(TAB, 'a1')
    state().loadDraft('tab-2', structuredClone(tab().draft), tab().baseRevision, tab().dirty, tab().savedName)
    state().copyVersions(TAB, 'tab-2')
    expect(state().tabs['tab-2']!.aliasId).toBe('a1')
    expect(state().tabs['tab-2']!.versions['']!.draft.url).toBe('https://x/v1?edited')
  })
})
