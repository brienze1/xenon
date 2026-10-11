import { describe, expect, it } from 'vitest'
import { currentSnippetLanguage, editorLanguageOf, optionValue, rememberedOptions } from './snippet-options'

const languages = [
  { id: 'curl', label: 'cURL', options: [] },
  { id: 'go_native', label: 'Go', options: [] },
]

describe('snippet options', () => {
  it('picks the language last used, ignoring ones the generator no longer offers', () => {
    expect(currentSnippetLanguage([], languages)).toBe('curl')
    expect(currentSnippetLanguage([], [])).toBe('curl')
    expect(currentSnippetLanguage([{ language: 'go_native', options: {} }, { language: 'cobol', options: {} }], languages)).toBe('go_native')
    expect(currentSnippetLanguage([{ language: 'python_requests', options: {} }], [])).toBe('python_requests')
  })

  it('reads remembered options with the defaults as fallback', () => {
    const saved = [{ language: 'curl', options: { multiLine: 'false' } }]
    expect(rememberedOptions(saved, 'curl')).toEqual({ multiLine: 'false' })
    expect(rememberedOptions(saved, 'go_native')).toEqual({})
    expect(optionValue({ multiLine: 'false' }, 'multiLine', 'true')).toBe('false')
    expect(optionValue({}, 'longFormat', 'true')).toBe('true')
  })

  it('highlights JavaScript and shows the other languages as text', () => {
    expect(editorLanguageOf('js_fetch')).toBe('javascript')
    expect(editorLanguageOf('curl')).toBe('text')
  })
})
