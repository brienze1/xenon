import { describe, expect, it } from 'vitest'
import { emptyDraft, type RequestDraft } from '@/platform/mappers/request'
import { fieldText, withFieldText } from './draft-fields'

function row(key: string, value: string) {
  return { id: key, key, value, description: '', enabled: true }
}

function draft(): RequestDraft {
  const base = emptyDraft()
  return {
    ...base,
    url: 'https://x/:id',
    params: [row('page', '2')],
    headers: [row('Authorization', 'Bearer abc')],
    pathVariables: [{ id: 'p', key: 'id', value: '42', description: '' }],
    body: { ...base.body, urlencoded: [row('grant', 'code')], formdata: [{ ...row('file', 'x'), type: 'text', src: '', srcState: 'ok' }] },
  }
}

describe('draft fields', () => {
  it('reads the text of every field Inspect names', () => {
    const request = draft()
    expect(fieldText(request, 'url')).toBe('https://x/:id')
    expect(fieldText(request, 'params.0.key')).toBe('page')
    expect(fieldText(request, 'headers.0.value')).toBe('Bearer abc')
    expect(fieldText(request, 'path_variables.0.value')).toBe('42')
    expect(fieldText(request, 'body.urlencoded.0.value')).toBe('code')
    expect(fieldText(request, 'body.formdata.0.key')).toBe('file')
    expect(fieldText(request, 'params.3.key')).toBeNull()
    expect(fieldText(request, 'path_variables.3.value')).toBeNull()
    expect(fieldText(request, 'auth.bearer.token')).toBeNull()
  })

  it('replaces a field and leaves the rest untouched', () => {
    const request = draft()
    expect(withFieldText(request, 'url', 'u').url).toBe('u')
    expect(withFieldText(request, 'headers.0.value', 'Bearer {{t}}').headers[0]).toEqual(row('Authorization', 'Bearer {{t}}'))
    expect(withFieldText(request, 'params.0.key', 'p').params[0]?.key).toBe('p')
    expect(withFieldText(request, 'path_variables.0.value', '7').pathVariables[0]?.value).toBe('7')
    expect(withFieldText(request, 'body.urlencoded.0.key', 'g').body.urlencoded[0]?.key).toBe('g')
    const formdata = withFieldText(request, 'body.formdata.0.value', 'y').body.formdata[0]
    expect(formdata).toMatchObject({ value: 'y', type: 'text' })
    expect(withFieldText(request, 'auth.bearer.token', 'x')).toBe(request)
    expect(withFieldText(request, 'headers.0.value', 'v').params).toBe(request.params)
  })
})
