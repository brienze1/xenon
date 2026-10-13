import { describe, expect, it } from 'vitest'
import { beautify, canBeautify } from './beautify'

describe('beautify', () => {
  it('formats JSON keeping variables', () => {
    expect(beautify('{"id":{{userId}}}', 'json')).toBe('{\n  "id": {{userId}}\n}')
  })

  it('puts every XML element on its own line', () => {
    expect(beautify('<root><a id="1">x</a><b><c/></b><d>{{v}}</d></root>', 'xml')).toBe(
      '<root>\n  <a id="1">x</a>\n  <b>\n    <c />\n  </b>\n  <d>{{v}}</d>\n</root>',
    )
  })

  it('formats HTML and JavaScript keeping variables as typed', () => {
    expect(beautify('<div><p>{{name}}</p></div>', 'html')).toBe('<div>\n  <p>{{name}}</p>\n</div>')
    expect(beautify('const a={b:{{value}},c:[1,2]}', 'javascript')).toBe('const a = {\n  b: {{value}},\n  c: [1, 2]\n}')
  })

  it('offers nothing for plain text', () => {
    expect(canBeautify('text')).toBe(false)
    expect(canBeautify('xml')).toBe(true)
    expect(beautify('a  b', 'text')).toBeNull()
  })
})
