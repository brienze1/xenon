import { describe, expect, it } from 'vitest'
import { emptyDraft } from '@/platform/mappers/request'
import { isCurlCommand, withPastedRequest } from './curl'

describe('curl', () => {
  it('recognises a cURL command by its first word', () => {
    expect(isCurlCommand("curl 'https://api.test'")).toBe(true)
    expect(isCurlCommand('  CURL.exe -X POST https://api.test')).toBe(true)
    expect(isCurlCommand('https://api.test/curl me')).toBe(false)
    expect(isCurlCommand('curl')).toBe(false)
    expect(isCurlCommand('curly braces')).toBe(false)
  })

  it('replaces what the command describes and keeps the request identity', () => {
    const current = { ...emptyDraft(), id: 'r1', name: 'Kept', description: 'Notes', disabledAutoHeaders: ['User-Agent'], url: 'https://old.test' }
    const pasted = {
      ...emptyDraft(),
      id: 'other',
      name: 'Ignored',
      method: 'PUT',
      url: 'https://new.test',
      headers: [{ id: 'h', key: 'A', value: '1', description: '', enabled: true }],
      settings: { ...emptyDraft().settings, verifySsl: false },
    }
    expect(withPastedRequest(current, pasted)).toEqual({
      ...pasted,
      id: 'r1',
      name: 'Kept',
      description: 'Notes',
      disabledAutoHeaders: ['User-Agent'],
    })
  })
})
