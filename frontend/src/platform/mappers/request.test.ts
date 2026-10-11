import { describe, expect, it } from 'vitest'
import type { Request as RequestDto } from '../../../bindings/github.com/brienze1/xenon/internal/integration/entrypoint/dto/models'
import { cancelledResult, emptyDraft, noAcknowledgement, toRequestDraft, toRequestDto, toSendResult } from './request'

const dto: RequestDto = {
  scripts: { pre_request: 'pm.test()', post_response: '', disabled: ['postResponse', 'bogus'] },
  inherit_scripts: false,
  id: '01J',
  name: 'Create invoice',
  description: 'd',
  method: 'POST',
  url: 'https://h/:id?a=1',
  revision: 'r',
  params: [{ key: 'a', value: '1', description: '', enabled: true }, { key: 'b', value: '2', description: 'x', enabled: false }],
  path_variables: [{ key: 'id', value: '7', description: '' }],
  headers: [{ key: 'Accept', value: '*/*', description: '' }],
  auto_headers: { disabled: ['User-Agent'] },
  auth: { type: 'bearer', bearer: { token: 't' }, basic: { username: 'u', password: 'p' } },
  body: {
    mode: 'formdata',
    raw: { language: 'xml', content: '<a/>' },
    formdata: [{ key: 'f', type: 'file', value: '', src: '/a.bin', src_state: 'repick', description: '', enabled: true }],
    urlencoded: [{ key: 'u', value: '1', description: '', enabled: true }],
    binary: { src: '/b.bin', src_state: 'ok' },
    graphql: { query: 'q', variables: '{}' },
  },
  settings: { timeout_ms: 10, follow_redirects: false, verify_ssl: true, encode_url: null },
}

describe('request mapper', () => {
  it('maps a request DTO to a draft with row ids', () => {
    const draft = toRequestDraft(dto)
    expect(draft).toMatchObject({
      id: '01J',
      name: 'Create invoice',
      method: 'POST',
      url: 'https://h/:id?a=1',
      params: [{ key: 'a', enabled: true }, { key: 'b', enabled: false, description: 'x' }],
      pathVariables: [{ key: 'id', value: '7' }],
      headers: [{ key: 'Accept', enabled: true }],
      disabledAutoHeaders: ['User-Agent'],
      auth: { type: 'bearer', bearer: { token: 't' } },
      body: { mode: 'formdata', raw: { language: 'xml' }, formdata: [{ key: 'f', type: 'file', src: '/a.bin', srcState: 'repick' }] },
      settings: { timeoutMs: 10, followRedirects: false, verifySsl: true, encodeUrl: true },
    })
    const ids = [...draft.params, ...draft.headers].map((row) => row.id)
    expect(new Set(ids).size).toBe(ids.length)
  })

  it('round trips a draft to a DTO', () => {
    const back = toRequestDto(toRequestDraft(dto))
    expect(back).toEqual({ ...dto, revision: '', scripts: { pre_request: 'pm.test()', post_response: '', disabled: ['postResponse'] }, inherit_scripts: false, headers: [{ key: 'Accept', value: '*/*', description: '', enabled: true }], settings: { ...dto.settings, encode_url: true } })
  })

  it('keeps unsupported auth details read-only and the file re-pick state', () => {
    const draft = toRequestDraft({ ...dto, auth: { ...dto.auth, type: 'unsupported', unsupported: { source: 'postman', type: 'hawk' } } })
    expect(draft.auth).toMatchObject({ type: 'unsupported', unsupported: { source: 'postman', type: 'hawk' } })
    expect(draft.body.binary).toEqual({ src: '/b.bin', srcState: 'ok' })
    expect(toRequestDto(draft).auth).toEqual({ type: 'unsupported', bearer: { token: 't' }, basic: { username: 'u', password: 'p' } })
    expect(toRequestDraft(dto).auth.unsupported).toBeNull()
  })

  it('falls back to defaults for unknown or empty values', () => {
    const draft = toRequestDraft({
      ...dto,
      name: '',
      method: '',
      params: null,
      path_variables: null,
      headers: null,
      auto_headers: { disabled: null },
      auth: { ...dto.auth, type: 'weird' },
      body: { ...dto.body, mode: 'weird', raw: { language: 'weird', content: '' }, formdata: null, urlencoded: null },
    })
    expect(draft).toMatchObject({
      name: 'Untitled Request',
      method: 'GET',
      params: [],
      pathVariables: [],
      headers: [],
      disabledAutoHeaders: [],
      auth: { type: 'inherit' },
      body: { mode: 'none', raw: { language: 'json' }, formdata: [], urlencoded: [] },
    })
    expect(emptyDraft().settings).toEqual({ timeoutMs: 0, followRedirects: true, verifySsl: true, encodeUrl: true })
  })

  it('maps send results and failures', () => {
    const timings = { dns_ms: 1, connect_ms: 2, tls_ms: 3, first_byte_ms: 4, download_ms: 5, total_ms: 15 }
    const ok = toSendResult({
      tab_id: 't',
      send_id: 's',
      outcome: 'sent',
      tests: [{ index: 0, name: 'ok', status: 'passed', script: { owner_kind: 'request', owner_id: 'r', owner_name: 'R', phase: 'postResponse', label: 'R · Post-response' }, async: false }],
      test_summary: { total: 1, passed: 1, failed: 0, skipped: 0 },
      script_errors: [{ kind: 'script_runtime', message: 'Error: x', line: 2, column: 3, script: { owner_kind: 'request', owner_id: 'r', owner_name: 'R', phase: 'postResponse', label: 'R · Post-response' } }],
      ok: true,
      status_code: 200,
      status_text: 'OK',
      proto: 'HTTP/1.1',
      headers: [{ key: 'A', value: '1' }],
      size_bytes: 2,
      content_type: 'text/plain',
      is_binary: false,
      body_url: '/xenon/response/t',
      inline_body: 'hi',
      timings,
    })
    expect(ok).toMatchObject({
      sendId: 's',
      outcome: 'sent',
      tests: [{ index: 0, name: 'ok', status: 'passed', errorMessage: '', scriptLabel: 'R · Post-response', async: false }],
      testSummary: { total: 1, passed: 1 },
      scriptErrors: [{ kind: 'script_runtime', message: 'Error: x', line: 2, column: 3, scriptLabel: 'R · Post-response' }],
    })
    expect(ok).toMatchObject({ tabId: 't', inlineBody: 'hi', error: null, timings: { dnsMs: 1, totalMs: 15 }, headers: [{ key: 'A', value: '1' }] })
    const failed = toSendResult({ ...ok, tab_id: 't', send_id: 's', outcome: 'unknown', tests: null, test_summary: { total: 0, passed: 0, failed: 0, skipped: 0 }, script_errors: null, ok: false, status_code: 0, status_text: '', proto: '', headers: null, size_bytes: 0, content_type: '', is_binary: false, body_url: '', timings, error: { kind: 'nope', message: 'm', hint: 'h' } })
    expect(failed).toMatchObject({ outcome: 'sent', tests: [], scriptErrors: [], inlineBody: null, headers: [], error: { kind: 'other', message: 'm', hint: 'h' } })
    expect(cancelledResult('t')).toMatchObject({ ok: false, error: { kind: 'cancelled', message: 'Request cancelled' } })
  })

  it('starts with nothing acknowledged', () => {
    expect(noAcknowledgement()).toEqual({ confirmedEnvironmentId: '', missingSecrets: false })
  })
})
