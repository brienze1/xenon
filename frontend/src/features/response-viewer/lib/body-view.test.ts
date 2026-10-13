import { describe, expect, it } from 'vitest'
import { initialBodyView, isHtml, isImage, languageFor, prettyBody, suggestedFileName } from './body-view'

describe('body view helpers', () => {
  it('opens bodies above 5 MB in Raw first', () => {
    expect(initialBodyView(5 * 1024 * 1024)).toBe('pretty')
    expect(initialBodyView(5 * 1024 * 1024 + 1)).toBe('raw')
  })

  it.each([
    ['application/json; charset=utf-8', 'json'],
    ['application/problem+json', 'json'],
    ['text/html', 'html'],
    ['application/xml', 'xml'],
    ['text/javascript', 'javascript'],
    ['text/plain', 'text'],
    ['', 'text'],
  ])('%s is highlighted as %s', (contentType, language) => {
    expect(languageFor(contentType)).toBe(language)
  })

  it('pretty prints every language except plain text, and keeps broken JSON as received', () => {
    expect(prettyBody('{"a":[1]}', 'json')).toBe('{\n  "a": [\n    1\n  ]\n}')
    expect(prettyBody('{broken', 'json')).toBe('{broken')
    expect(prettyBody('<a><b/></a>', 'xml')).toBe('<a>\n  <b />\n</a>')
    expect(prettyBody('a  b', 'text')).toBe('a  b')
  })

  it('classifies previews', () => {
    expect(isImage('image/png')).toBe(true)
    expect(isImage('text/plain')).toBe(false)
    expect(isHtml('text/html; charset=utf-8')).toBe(true)
    expect(isHtml('application/xhtml+xml')).toBe(false)
  })

  it.each([
    ['application/json', 'response.json'],
    ['text/html', 'response.html'],
    ['text/plain', 'response.txt'],
    ['image/svg+xml', 'response.svg'],
    ['image/png', 'response.png'],
    ['application/octet-stream', 'response.bin'],
    ['application/javascript', 'response.js'],
  ])('suggests a file name for %s', (contentType, name) => {
    expect(suggestedFileName(contentType)).toBe(name)
  })
})
