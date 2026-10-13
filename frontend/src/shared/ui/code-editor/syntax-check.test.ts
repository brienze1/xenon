import { describe, expect, it } from 'vitest'
import { findSyntaxErrors } from './syntax-check'

describe('findSyntaxErrors', () => {
  it('points at a trailing comma in JSON', async () => {
    expect(await findSyntaxErrors('{\n  "ids": [\n    500947,\n  ]\n}', 'json')).toEqual([{ line: 3, column: 11, message: 'Trailing comma before "]"' }])
  })

  it('points at a missing comma and at an unclosed body', async () => {
    expect(await findSyntaxErrors('{ "a": 1 "b": 2 }', 'json')).toEqual([{ line: 1, column: 10, message: 'Missing comma before this value' }])
    expect(await findSyntaxErrors('{ "a": 1,', 'json')).toEqual([{ line: 1, column: 9, message: 'Unexpected end of body: something is not closed' }])
  })

  it('accepts variables anywhere and valid bodies', async () => {
    expect(await findSyntaxErrors('{ "id": {{userId}}, "name": "{{name}}" }', 'json')).toEqual([])
    expect(await findSyntaxErrors('<a><b>{{v}}</b></a>', 'xml')).toEqual([])
    expect(await findSyntaxErrors('const a = {{value}}', 'javascript')).toEqual([])
    expect(await findSyntaxErrors('', 'json')).toEqual([])
  })

  it('reports mismatched closing tags in XML and HTML', async () => {
    expect(await findSyntaxErrors('<a>\n  <b>1</c>\n</a>', 'xml')).toEqual([
      { line: 2, column: 7, message: 'Closing tag </c> does not match the open element' },
    ])
    expect(await findSyntaxErrors('<div><p>x</span></div>', 'html')).toEqual([
      { line: 1, column: 10, message: 'Closing tag </span> does not match the open element' },
    ])
  })

  it('reports broken JavaScript and nothing for plain text', async () => {
    expect(await findSyntaxErrors('foo(', 'javascript')).toEqual([{ line: 1, column: 4, message: 'Unexpected end of body: something is not closed' }])
    expect(await findSyntaxErrors('{{{ not checked', 'text')).toEqual([])
  })
})

describe('findSyntaxErrors in YAML', () => {
  it('accepts documents with quoted variables', async () => {
    expect(await findSyntaxErrors("openapi: 3.1.0\nservers:\n  - url: '{{baseUrl}}'\npaths: {}\n", 'yaml')).toEqual([])
  })

  it('points at an unclosed flow sequence', async () => {
    expect(await findSyntaxErrors('info:\n  title: [unclosed\n', 'yaml')).toHaveLength(1)
  })
})
