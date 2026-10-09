import { describe, expect, it } from 'vitest'
import { emptyDraft } from '@/platform/mappers/request'
import { sameDraftContent } from './draft'

const ROW = { key: 'a', value: '1', description: '', enabled: true }

describe('sameDraftContent', () => {
  it('matches two empty drafts', () => {
    expect(sameDraftContent(emptyDraft(), emptyDraft())).toBe(true)
  })

  it('ignores the UI-only ids of table rows', () => {
    const left = { ...emptyDraft(), params: [{ ...ROW, id: 'row-1' }] }
    const right = { ...emptyDraft(), params: [{ ...ROW, id: 'row-9' }] }
    expect(sameDraftContent(left, right)).toBe(true)
  })

  it('still compares the request id', () => {
    expect(sameDraftContent({ ...emptyDraft(), id: 'r1' }, emptyDraft())).toBe(false)
  })

  it('detects a changed field, a changed row and an extra row', () => {
    expect(sameDraftContent({ ...emptyDraft(), url: 'https://x' }, emptyDraft())).toBe(false)
    expect(sameDraftContent({ ...emptyDraft(), params: [{ ...ROW, id: 'a' }] }, { ...emptyDraft(), params: [{ ...ROW, id: 'a', value: '2' }] })).toBe(false)
    expect(sameDraftContent({ ...emptyDraft(), headers: [{ ...ROW, id: 'a' }] }, emptyDraft())).toBe(false)
  })

  it('compares nested settings, auth and body', () => {
    const base = emptyDraft()
    expect(sameDraftContent({ ...base, settings: { ...base.settings, verifySsl: false } }, emptyDraft())).toBe(false)
    expect(sameDraftContent({ ...base, auth: { ...base.auth, type: 'bearer' } }, emptyDraft())).toBe(false)
    expect(sameDraftContent({ ...base, body: { ...base.body, raw: { ...base.body.raw, content: '{}' } } }, emptyDraft())).toBe(false)
    expect(sameDraftContent({ ...base, disabledAutoHeaders: ['User-Agent'] }, emptyDraft())).toBe(false)
  })
})
