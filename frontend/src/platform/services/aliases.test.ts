import { beforeEach, describe, expect, it, vi } from 'vitest'
import { emptyDraft, toRequestDto } from '../mappers/request'

vi.mock('@wailsio/runtime', () => ({ CancelError: class CancelError extends Error {} }))
vi.mock('../../../bindings/github.com/brienze1/xenon/internal/infra/server/binding/aliasservice', () => ({
  Create: vi.fn(),
  Save: vi.fn(),
  Delete: vi.fn(),
  SetEnvironments: vi.fn(),
}))

const AliasService = await import('../../../bindings/github.com/brienze1/xenon/internal/infra/server/binding/aliasservice')
const aliases = await import('./aliases')

const requestDto = {
  ...toRequestDto({ ...emptyDraft(), id: 'r1', name: 'Login', url: 'https://x/v1' }),
  revision: 'rev-2',
  aliases: [{ id: 'a1', name: 'mx', environment_ids: ['hml'], request: toRequestDto({ ...emptyDraft(), id: '', name: '', url: 'https://x/v2' }) }],
}

describe('alias service', () => {
  beforeEach(() => vi.resetAllMocks())

  it('creates an alias and maps every alias as a draft of the request named after the alias', async () => {
    vi.mocked(AliasService.Create).mockResolvedValue(requestDto as never)
    const saved = await aliases.createAlias('ws', 'r1', 'mx', ['hml'], 'rev-1')
    expect(AliasService.Create).toHaveBeenCalledWith({ local_id: 'ws', request_id: 'r1', base_revision: 'rev-1', name: 'mx', environment_ids: ['hml'] })
    expect(saved.revision).toBe('rev-2')
    expect(saved.aliases).toEqual([{ id: 'a1', name: 'mx', environmentIds: ['hml'], draft: expect.objectContaining({ id: 'r1', name: 'mx', url: 'https://x/v2' }) }])
  })

  it('saves one alias with its name and content', async () => {
    vi.mocked(AliasService.Save).mockResolvedValue(requestDto as never)
    const draft = { ...emptyDraft(), id: 'r1', name: 'mx-renamed', url: 'https://x/v2?edited' }
    await aliases.saveAlias('ws', 'a1', draft, 'rev-1', true)
    expect(AliasService.Save).toHaveBeenCalledWith({
      local_id: 'ws',
      request_id: 'r1',
      base_revision: 'rev-1',
      force: true,
      alias: { id: 'a1', name: 'mx-renamed', environment_ids: [], request: toRequestDto(draft) },
    })
  })

  it('sets the environments of an alias and reads them back', async () => {
    vi.mocked(AliasService.SetEnvironments).mockResolvedValue(requestDto as never)
    const saved = await aliases.setAliasEnvironments('ws', 'r1', 'a1', ['hml'], 'rev-1')
    expect(AliasService.SetEnvironments).toHaveBeenCalledWith({ local_id: 'ws', request_id: 'r1', alias_id: 'a1', environment_ids: ['hml'], base_revision: 'rev-1' })
    expect(saved.aliases[0]?.environmentIds).toEqual(['hml'])
    vi.mocked(AliasService.SetEnvironments).mockRejectedValue({ cause: { error: { code: 'ALS-10400', description: 'the environment does not exist in this workspace' } } })
    await expect(aliases.setAliasEnvironments('ws', 'r1', 'a1', ['gone'], 'rev-1')).rejects.toMatchObject({ code: 'ALS-10400' })
  })

  it('reads an alias without environments as one for all environments', async () => {
    vi.mocked(AliasService.Create).mockResolvedValue({ ...requestDto, aliases: [{ ...requestDto.aliases[0], environment_ids: null }] } as never)
    const saved = await aliases.createAlias('ws', 'r1', 'mx', [], 'rev-1')
    expect(saved.aliases[0]?.environmentIds).toEqual([])
  })

  it('deletes an alias and turns failures into app errors', async () => {
    vi.mocked(AliasService.Delete).mockResolvedValue(requestDto as never)
    await aliases.deleteAlias('ws', 'r1', 'a1', 'rev-1')
    expect(AliasService.Delete).toHaveBeenCalledWith({ local_id: 'ws', request_id: 'r1', alias_id: 'a1', base_revision: 'rev-1' })
    vi.mocked(AliasService.Delete).mockRejectedValue({ cause: { error: { code: 'ALS-01404', description: 'the alias does not exist' } } })
    await expect(aliases.deleteAlias('ws', 'r1', 'a1', 'rev-1')).rejects.toMatchObject({ code: 'ALS-01404' })
  })
})
