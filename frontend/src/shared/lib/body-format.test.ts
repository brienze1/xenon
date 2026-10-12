import { describe, expect, it } from 'vitest'
import { RAW_FIRST_BYTES, initialBodyView, isHtml, isImage, languageFor, mediaType, prettyBody } from './body-format'

describe('body format', () => {
  it('opens large bodies in Raw first', () => {
    expect(initialBodyView(RAW_FIRST_BYTES)).toBe('pretty')
    expect(initialBodyView(RAW_FIRST_BYTES + 1)).toBe('raw')
  })

  it('reads the media type and the editor language from a content type', () => {
    expect(mediaType('Application/JSON; charset=utf-8')).toBe('application/json')
    expect(mediaType('')).toBe('')
    expect(languageFor('application/problem+json')).toBe('json')
    expect(languageFor('text/html')).toBe('html')
    expect(languageFor('application/xml')).toBe('xml')
    expect(languageFor('text/javascript')).toBe('javascript')
    expect(languageFor('text/csv')).toBe('text')
  })

  it('pretty prints JSON only and keeps invalid JSON as it is', () => {
    expect(prettyBody('{"a":1}', 'json')).toBe('{\n  "a": 1\n}')
    expect(prettyBody('{"a":', 'json')).toBe('{"a":')
    expect(prettyBody('<a><b>1</b></a>', 'xml')).toBe('<a>\n  <b>1</b>\n</a>')
    expect(prettyBody('{"big":12345678901234567890}', 'json')).toBe('{\n  "big": 12345678901234567890\n}')
    expect(prettyBody('plain  text', 'text')).toBe('plain  text')
  })

  it('recognises previewable content', () => {
    expect(isImage('image/png')).toBe(true)
    expect(isImage('text/plain')).toBe(false)
    expect(isHtml('text/html; charset=utf-8')).toBe(true)
    expect(isHtml('application/xhtml+xml')).toBe(false)
  })
})
