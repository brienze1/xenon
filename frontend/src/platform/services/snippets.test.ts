import { beforeEach, describe, expect, it, vi } from 'vitest'
import { emptyDraft } from '../mappers/request'

vi.mock('@wailsio/runtime', () => ({ CancelError: class CancelError extends Error {} }))
vi.mock('../../../bindings/github.com/brienze1/xenon/internal/infra/server/binding/snippetservice', () => ({
  Languages: vi.fn(),
  Generate: vi.fn(),
  Copy: vi.fn(),
}))

const SnippetService = await import('../../../bindings/github.com/brienze1/xenon/internal/infra/server/binding/snippetservice')
const snippets = await import('./snippets')
const { toRequestDto } = await import('../mappers/request')

const input = { localId: 'ws', tabId: 't1', draft: { ...emptyDraft(), url: 'https://x' }, language: 'curl', options: { multiLine: 'false' }, includeSecrets: true }

describe('snippet service', () => {
  beforeEach(() => vi.resetAllMocks())

  it('lists the languages with their options as text', async () => {
    vi.mocked(SnippetService.Languages).mockResolvedValue([
      { id: 'curl', label: 'cURL', options: [{ id: 'multiLine', label: 'Multiline', description: 'Split across lines', type: 'boolean', default: true }, { id: 'quoteType', label: 'Quote', type: 'enum', default: 'single', values: ['single', 'double'] }, { id: 'requestTimeoutInSeconds', label: 'Timeout', type: 'positiveInteger', default: 0 }] },
      { id: 'go_native', label: 'Go', options: null },
    ] as never)
    await expect(snippets.listSnippetLanguages()).resolves.toEqual([
      {
        id: 'curl',
        label: 'cURL',
        options: [
          { id: 'multiLine', label: 'Multiline', description: 'Split across lines', type: 'boolean', defaultValue: 'true', values: [] },
          { id: 'quoteType', label: 'Quote', description: '', type: 'enum', defaultValue: 'single', values: ['single', 'double'] },
          { id: 'requestTimeoutInSeconds', label: 'Timeout', description: '', type: 'positiveInteger', defaultValue: '0', values: [] },
        ],
      },
      { id: 'go_native', label: 'Go', options: [] },
    ])
    vi.mocked(SnippetService.Languages).mockResolvedValue(null as never)
    await expect(snippets.listSnippetLanguages()).resolves.toEqual([])
  })

  it('generates and copies the draft as code', async () => {
    vi.mocked(SnippetService.Generate).mockResolvedValue({ text: 'curl https://x', notices: null } as never)
    await expect(snippets.generateSnippet(input)).resolves.toEqual({ text: 'curl https://x', notices: [] })
    const expected = { local_id: 'ws', tab_id: 't1', language: 'curl', options: { multiLine: 'false' }, include_secrets: true, request: toRequestDto(input.draft) }
    expect(SnippetService.Generate).toHaveBeenCalledWith(expected)
    vi.mocked(SnippetService.Copy).mockResolvedValue(undefined as never)
    await snippets.copySnippet(input)
    expect(SnippetService.Copy).toHaveBeenCalledWith(expected)
  })

  it('maps failures and empty answers to app errors', async () => {
    vi.mocked(SnippetService.Generate).mockResolvedValue(null as never)
    await expect(snippets.generateSnippet(input)).rejects.toMatchObject({ code: 'UNK-00204' })
    vi.mocked(SnippetService.Copy).mockRejectedValue(Object.assign(new Error('x'), { cause: { error: { code: 'SNP-01400', description: 'bad' } } }))
    await expect(snippets.copySnippet(input)).rejects.toMatchObject({ code: 'SNP-01400' })
  })
})
