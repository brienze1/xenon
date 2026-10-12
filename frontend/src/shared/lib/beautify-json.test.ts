import { describe, expect, it } from 'vitest'
import { beautifyJson } from './beautify-json'

describe('beautifyJson', () => {
  it('indents objects and arrays with two spaces', () => {
    expect(beautifyJson('{\n\n  "test":"lkalala", "list":[1,2]}')).toBe('{\n  "test": "lkalala",\n  "list": [\n    1,\n    2\n  ]\n}')
  })

  it('keeps empty objects and arrays on one line', () => {
    expect(beautifyJson('{"a":{ },"b":[ ]}')).toBe('{\n  "a": {},\n  "b": []\n}')
  })

  it('keeps variables, big numbers and duplicate keys as typed', () => {
    expect(beautifyJson('{"id":{{userId}},"n":12345678901234567890,"id":"{{x}}"}')).toBe(
      '{\n  "id": {{userId}},\n  "n": 12345678901234567890,\n  "id": "{{x}}"\n}',
    )
  })

  it('leaves whitespace, brackets and escaped quotes inside strings alone', () => {
    expect(beautifyJson('{"s":"a, b: {c} [d] \\"e\\""}')).toBe('{\n  "s": "a, b: {c} [d] \\"e\\""\n}')
  })

  it('returns null when brackets or strings do not balance', () => {
    expect(beautifyJson('{"a":1')).toBeNull()
    expect(beautifyJson('{"a":1]')).toBeNull()
    expect(beautifyJson('{"a":"1}')).toBeNull()
  })
})
