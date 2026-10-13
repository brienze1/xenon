import { describe, expect, it } from 'vitest'
import type { Request as RequestDto } from '../../../bindings/github.com/brienze1/xenon/internal/integration/entrypoint/dto/models'
import { cancelledResult, emptyAuth, emptyDraft, isLegacyGrant, noAcknowledgement, toAuth, toAuthDto, toRequestDraft, toRequestDto, toSendResult, usesBrowser } from './request'

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
  settings: { timeout_ms: 10, follow_redirects: false, verify_ssl: true, encode_url: null, disable_cookie_jar: false },
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
    expect(back).toEqual({ ...dto, auth: { ...toAuthDto(emptyAuth()), type: 'bearer', bearer: { token: 't' }, basic: { username: 'u', password: 'p' } }, revision: '', scripts: { pre_request: 'pm.test()', post_response: '', disabled: ['postResponse'] }, inherit_scripts: false, headers: [{ key: 'Accept', value: '*/*', description: '', enabled: true }], settings: { ...dto.settings, encode_url: true } })
  })

  it('keeps unsupported auth details read-only and the file re-pick state', () => {
    const draft = toRequestDraft({ ...dto, auth: { ...dto.auth, type: 'unsupported', unsupported: { source: 'postman', type: 'hawk' } } })
    expect(draft.auth).toMatchObject({ type: 'unsupported', unsupported: { source: 'postman', type: 'hawk' } })
    expect(draft.body.binary).toEqual({ src: '/b.bin', srcState: 'ok' })
    expect(toRequestDto(draft).auth).toMatchObject({ type: 'unsupported', bearer: { token: 't' }, basic: { username: 'u', password: 'p' } })
    expect(toRequestDto(draft).auth).not.toHaveProperty('unsupported')
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
    expect(emptyDraft().settings).toEqual({ timeoutMs: 0, followRedirects: true, verifySsl: true, encodeUrl: true, disableCookieJar: false })
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
      cookies_set: [{ name: 'session', value: 'abc', domain: 'api.test', path: '/', expires: '', same_site: 'Lax', secure: true, http_only: true, host_only: true, raw: 'session=abc; Path=/' }],
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
    expect(ok.cookiesSet).toEqual([{ name: 'session', value: 'abc', domain: 'api.test', path: '/', expires: null, sameSite: 'Lax', secure: true, httpOnly: true, hostOnly: true, raw: 'session=abc; Path=/' }])
    const failed = toSendResult({ ...ok, tab_id: 't', send_id: 's', outcome: 'unknown', tests: null, test_summary: { total: 0, passed: 0, failed: 0, skipped: 0 }, script_errors: null, ok: false, status_code: 0, status_text: '', proto: '', headers: null, cookies_set: null, size_bytes: 0, content_type: '', is_binary: false, body_url: '', timings, error: { kind: 'nope', message: 'm', hint: 'h' } })
    expect(failed).toMatchObject({ outcome: 'sent', tests: [], scriptErrors: [], inlineBody: null, headers: [], cookiesSet: [], error: { kind: 'other', message: 'm', hint: 'h' } })
    expect(cancelledResult('t')).toMatchObject({ ok: false, error: { kind: 'cancelled', message: 'Request cancelled' } })
  })

  it('starts with nothing acknowledged', () => {
    expect(noAcknowledgement()).toEqual({ confirmedEnvironmentId: '', missingSecrets: false, expiredToken: false })
  })

  it('maps the API key, OAuth 2.0, AWS and Digest settings both ways', () => {
    const full = {
      type: 'oauth2',
      bearer: { token: '' },
      basic: { username: '', password: '' },
      apikey: { key: 'X-Key', value: 'v', in: 'query' },
      oauth2: {
        grant_type: 'authorization_code_with_pkce',
        token_name: 'n',
        redirect_uri: 'http://127.0.0.1:1/cb',
        auth_url: 'https://a',
        access_token_url: 'https://t',
        refresh_token_url: 'https://r',
        client_id: 'c',
        client_secret: '{{s}}',
        username: 'u',
        password: 'p',
        scope: 'openid',
        state: 'st',
        challenge_algorithm: 'plain',
        code_verifier: 'v',
        client_authentication: 'body',
        add_token_to: 'queryParams',
        header_prefix: 'Token',
        use_token: 'id_token',
        audience: 'au',
        resource: 're',
        auto_refresh: false,
        auth_request_params: [{ key: 'k', value: 'v', enabled: true, send_as: 'request_url' }],
        token_request_params: [{ key: 'h', value: '1', enabled: false, send_as: 'request_header' }],
        refresh_request_params: [{ key: 'b', value: '2', enabled: true, send_as: 'weird' }],
      },
      awsv4: { access_key: 'a', secret_key: 's', session_token: 't', region: 'r', service: 'v', add_auth_data_to_query: true },
      digest: { username: 'u', password: 'p', algorithm: 'SHA-256-sess', qop: 'auth-int', realm: 'r', nonce: 'n', opaque: 'o', nonce_count: '1', client_nonce: 'c', disable_retry_request: true },
    }
    const auth = toAuth({ ...full, oauth2: { ...full.oauth2, uses_postman_callback: true, suggested_callback: 'http://127.0.0.1:53682/callback', legacy: false } })
    expect(auth.apikey).toEqual({ key: 'X-Key', value: 'v', in: 'query' })
    expect(auth.oauth2).toMatchObject({ grantType: 'authorization_code_with_pkce', challengeAlgorithm: 'plain', clientAuthentication: 'body', addTokenTo: 'queryParams', useToken: 'id_token', autoRefresh: false, usesPostmanCallback: true, suggestedCallback: 'http://127.0.0.1:53682/callback' })
    expect(auth.oauth2.refreshRequestParams).toEqual([{ key: 'b', value: '2', enabled: true, sendAs: 'request_body' }])
    expect(auth.awsv4).toEqual({ accessKey: 'a', secretKey: 's', sessionToken: 't', region: 'r', service: 'v', addAuthDataToQuery: true })
    expect(auth.digest).toMatchObject({ algorithm: 'SHA-256-sess', qop: 'auth-int', nonceCount: '1', clientNonce: 'c', disableRetryRequest: true })
    expect(toAuthDto(auth)).toEqual({ ...full, oauth2: { ...full.oauth2, refresh_request_params: [{ key: 'b', value: '2', enabled: true, send_as: 'request_body' }] } })
    const odd = toAuth({
      type: 'digest',
      bearer: { token: '' },
      basic: { username: '', password: '' },
      apikey: { key: '', value: '', in: 'cookie' },
      oauth2: { ...full.oauth2, grant_type: 'device', redirect_uri: '', challenge_algorithm: 'S512', client_authentication: 'x', add_token_to: 'x', use_token: 'x', auth_request_params: null, token_request_params: null, refresh_request_params: null },
      awsv4: null,
      digest: { ...full.digest, algorithm: 'SHA-1', qop: 'x' },
    })
    expect(odd.apikey.in).toBe('header')
    expect(odd.oauth2).toMatchObject({ grantType: 'authorization_code', redirectUri: 'http://127.0.0.1:53682/callback', challengeAlgorithm: 'S256', clientAuthentication: 'header', addTokenTo: 'header', useToken: 'access_token', authRequestParams: [], usesPostmanCallback: false, suggestedCallback: '' })
    expect(odd.awsv4).toEqual(emptyAuth().awsv4)
    expect(odd.digest).toMatchObject({ algorithm: 'MD5', qop: '' })
    expect(toAuth({ type: 'none', bearer: { token: '' }, basic: { username: '', password: '' } })).toEqual({ ...emptyAuth(), type: 'none' })
  })

  it('tells legacy and browser grants apart', () => {
    expect(['implicit', 'password_credentials'].every((grant) => isLegacyGrant(grant as never))).toBe(true)
    expect(isLegacyGrant('client_credentials')).toBe(false)
    expect(usesBrowser('authorization_code') && usesBrowser('authorization_code_with_pkce') && usesBrowser('implicit')).toBe(true)
    expect(usesBrowser('password_credentials') || usesBrowser('client_credentials')).toBe(false)
  })
})
