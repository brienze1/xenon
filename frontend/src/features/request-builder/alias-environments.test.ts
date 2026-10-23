import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest'
import { AppError } from '@/platform/errors'
import { emptyDraft, type RequestDraft, type SavedRequest } from '@/platform/mappers/request'

vi.mock('@/platform/services/requests', async (importOriginal) => {
  const { confirmationOf, expiredTokenOwnerOf, missingSecretsOf } = await importOriginal<typeof import('@/platform/services/requests')>()
  return { confirmationOf, expiredTokenOwnerOf, missingSecretsOf, getRequest: vi.fn(), autoHeaders: vi.fn(), sendRequest: vi.fn() }
})
vi.mock('@/platform/services/aliases', () => ({ createAlias: vi.fn(), saveAlias: vi.fn(), deleteAlias: vi.fn(), setAliasEnvironments: vi.fn() }))
vi.mock('@/platform/services/variables', () => ({ inspectVariables: vi.fn(async () => ({ references: [], unresolved: [] })), suggestVariables: vi.fn() }))

const requests = await import('@/platform/services/requests')
const aliasService = await import('@/platform/services/aliases')
const { useWorkspacesStore } = await import('@/features/workspaces/store')
const { useEnvironmentsStore } = await import('@/features/environments/store')
const { useRequestBuilderStore } = await import('./store')

const TAB = 'tab-1'

function draftOf(name: string, url: string): RequestDraft {
  return { ...emptyDraft(), id: 'r1', name, url }
}

function saved(revision: string, mxEnvironments: string[]): SavedRequest {
  return { draft: draftOf('Login', 'https://x/v1'), revision, aliases: [{ id: 'a1', name: 'mx-test-user', environmentIds: mxEnvironments, draft: draftOf('mx-test-user', 'https://x/v2') }] }
}

function tab() {
  return useRequestBuilderStore.getState().tabs[TAB]!
}

describe('request builder - alias environments', () => {
  beforeEach(async () => {
    vi.resetAllMocks()
    useRequestBuilderStore.getState().reset()
    useWorkspacesStore.setState({ activeLocalId: 'ws' })
    useEnvironmentsStore.setState({ activeId: 'hml' })
    vi.mocked(requests.autoHeaders).mockResolvedValue([])
    vi.mocked(requests.getRequest).mockResolvedValue(saved('rev-1', ['hml']))
    vi.mocked(requests.sendRequest).mockReturnValue({ sendId: 's1', result: new Promise(() => undefined) } as never)
    await useRequestBuilderStore.getState().openSaved(TAB, 'r1', 'a1')
  })

  afterEach(() => {
    useWorkspacesStore.setState({ activeLocalId: null })
    useEnvironmentsStore.setState({ activeId: null })
  })

  it('does not send an alias the active environment is not one of, and sends it once it is', async () => {
    useEnvironmentsStore.setState({ activeId: 'prod' })
    await useRequestBuilderStore.getState().send(TAB)
    expect(requests.sendRequest).not.toHaveBeenCalled()
    useEnvironmentsStore.setState({ activeId: null })
    await useRequestBuilderStore.getState().send(TAB)
    expect(requests.sendRequest).not.toHaveBeenCalled()
    useEnvironmentsStore.setState({ activeId: 'hml' })
    void useRequestBuilderStore.getState().send(TAB)
    expect(requests.sendRequest).toHaveBeenCalledTimes(1)
  })

  it('keeps the alias and its unsaved edits on screen when the active environment changes', () => {
    useRequestBuilderStore.getState().setUrl(TAB, 'https://x/v2?edited')
    useEnvironmentsStore.setState({ activeId: 'prod' })
    expect(tab().aliasId).toBe('a1')
    expect(tab().draft.url).toBe('https://x/v2?edited')
    expect(tab().dirty).toBe(true)
  })

  it('saves new environments with the tab’s revision and leaves the alias’s draft as it is', async () => {
    useRequestBuilderStore.getState().setUrl(TAB, 'https://x/v2?edited')
    vi.mocked(aliasService.setAliasEnvironments).mockResolvedValue(saved('rev-2', ['dev']))
    expect(await useRequestBuilderStore.getState().setAliasEnvironments(TAB, 'a1', ['dev'])).toBeNull()
    expect(aliasService.setAliasEnvironments).toHaveBeenCalledWith('ws', 'r1', 'a1', ['dev'], 'rev-1')
    expect(tab().baseRevision).toBe('rev-2')
    expect(tab().aliases[0]?.environmentIds).toEqual(['dev'])
    expect(tab().draft.url).toBe('https://x/v2?edited')
    await useRequestBuilderStore.getState().send(TAB)
    expect(requests.sendRequest).not.toHaveBeenCalled()
  })

  it('returns Go’s refusal and flags a conflict when the file changed on disk', async () => {
    vi.mocked(aliasService.setAliasEnvironments).mockRejectedValueOnce(new AppError('ALS-10400', 'the environment does not exist in this workspace'))
    expect(await useRequestBuilderStore.getState().setAliasEnvironments(TAB, 'a1', ['gone'])).toMatchObject({ code: 'ALS-10400' })
    expect(tab().conflict).toBe(false)
    vi.mocked(aliasService.setAliasEnvironments).mockRejectedValueOnce(new AppError('REQ-02409', 'the request changed on disk since it was opened'))
    await useRequestBuilderStore.getState().setAliasEnvironments(TAB, 'a1', ['hml'])
    expect(tab().conflict).toBe(true)
  })

  it('creates an alias limited to the chosen environments', async () => {
    vi.mocked(aliasService.createAlias).mockResolvedValue(saved('rev-2', ['hml']))
    await useRequestBuilderStore.getState().createAlias(TAB, 'canary', ['hml'])
    expect(aliasService.createAlias).toHaveBeenCalledWith('ws', 'r1', 'canary', ['hml'], 'rev-1')
  })
})
