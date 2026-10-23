import type {
  Auth as AuthDto,
  AuthOAuth2 as AuthOAuth2Dto,
  OAuth2Param as OAuth2ParamDto,
  AutoHeader as AutoHeaderDto,
  FormDataRow as FormDataRowDto,
  KeyValue as KeyValueDto,
  PathVariable as PathVariableDto,
  Request as RequestDto,
  RequestAlias as RequestAliasDto,
  ScriptError as ScriptErrorDto,
  Scripts as ScriptsDto,
  SendResult as SendResultDto,
  TestResult as TestResultDto,
} from '../../../bindings/github.com/brienze1/xenon/internal/integration/entrypoint/dto/models'
import { newRowId } from '@/shared/lib/rows'
import { toCookie, type Cookie } from './cookie'

export const STANDARD_METHODS = ['GET', 'POST', 'PUT', 'PATCH', 'DELETE', 'HEAD', 'OPTIONS'] as const
export const AUTH_TYPES = ['inherit', 'none', 'bearer', 'basic', 'apikey', 'oauth2', 'awsv4', 'digest', 'unsupported'] as const
export const BODY_MODES = ['none', 'formdata', 'urlencoded', 'raw', 'binary', 'graphql'] as const
export const RAW_LANGUAGES = ['json', 'xml', 'text', 'html', 'javascript'] as const
export const SEND_OUTCOMES = ['sent', 'skipped_by_script', 'stopped_by_script_error', 'cancelled', 'error'] as const
export const SEND_ERROR_KINDS = ['dns', 'connection_refused', 'timeout', 'tls', 'cancelled', 'missing_file', 'invalid_url', 'other'] as const

const AUTH_TYPE_SET: ReadonlySet<string> = new Set(AUTH_TYPES)
const BODY_MODE_SET: ReadonlySet<string> = new Set(BODY_MODES)
const RAW_LANGUAGE_SET: ReadonlySet<string> = new Set(RAW_LANGUAGES)
export const SEND_ERROR_KIND_SET: ReadonlySet<string> = new Set(SEND_ERROR_KINDS)
export const SEND_OUTCOME_SET: ReadonlySet<string> = new Set(SEND_OUTCOMES)

export type AuthType = (typeof AUTH_TYPES)[number]
/** `repick` marks a file reference from another machine: it is never read or sent until the file is chosen again. */
export type FileState = 'ok' | 'repick'
export type BodyMode = (typeof BODY_MODES)[number]
export type RawLanguage = (typeof RAW_LANGUAGES)[number]
export type SendErrorKind = (typeof SEND_ERROR_KINDS)[number]
/** How a send ended: sent, stopped or skipped by a pre-request script, cancelled, or failed. */
export type SendOutcome = (typeof SEND_OUTCOMES)[number]

export interface KeyValueRow {
  id: string
  key: string
  value: string
  description: string
  enabled: boolean
}

export interface PathVariableRow {
  id: string
  key: string
  value: string
  description: string
}

export interface FormDataRow extends KeyValueRow {
  type: 'text' | 'file'
  src: string
  srcState: FileState
}

/** An imported auth type Xenon cannot send; kept unchanged for export and shown read-only. */
export interface UnsupportedAuth {
  source: string
  type: string
}

export type ApiKeyLocation = 'header' | 'query'

export interface ApiKeyAuth {
  key: string
  value: string
  in: ApiKeyLocation
}

export const OAUTH2_GRANTS = ['authorization_code', 'authorization_code_with_pkce', 'implicit', 'password_credentials', 'client_credentials'] as const
export type OAuth2Grant = (typeof OAUTH2_GRANTS)[number]
export type OAuth2ParamTarget = 'request_url' | 'request_body' | 'request_header'
export type TokenUse = 'access_token' | 'id_token'

/** A custom auth, token or refresh request parameter and where it is sent. */
export interface OAuth2Param {
  key: string
  value: string
  enabled: boolean
  sendAs: OAuth2ParamTarget
}

/** OAuth 2.0 settings, with Postman's field meaning; tokens never live here (they are in the vault). */
export interface OAuth2Auth {
  grantType: OAuth2Grant
  tokenName: string
  redirectUri: string
  authUrl: string
  accessTokenUrl: string
  refreshTokenUrl: string
  clientId: string
  clientSecret: string
  username: string
  password: string
  scope: string
  state: string
  challengeAlgorithm: 'S256' | 'plain'
  codeVerifier: string
  clientAuthentication: 'header' | 'body'
  addTokenTo: 'header' | 'queryParams'
  headerPrefix: string
  useToken: TokenUse
  audience: string
  resource: string
  autoRefresh: boolean
  authRequestParams: OAuth2Param[]
  tokenRequestParams: OAuth2Param[]
  refreshRequestParams: OAuth2Param[]
  /** Read-only: Postman's web callback URL is in use; `suggestedCallback` is the loopback URL to register. */
  usesPostmanCallback: boolean
  suggestedCallback: string
}

export interface AwsV4Auth {
  accessKey: string
  secretKey: string
  sessionToken: string
  region: string
  service: string
  addAuthDataToQuery: boolean
}

export const DIGEST_ALGORITHMS = ['MD5', 'MD5-sess', 'SHA-256', 'SHA-256-sess', 'SHA-512-256', 'SHA-512-256-sess'] as const
export type DigestAlgorithm = (typeof DIGEST_ALGORITHMS)[number]
export type DigestQop = '' | 'auth' | 'auth-int'

export interface DigestAuth {
  username: string
  password: string
  algorithm: DigestAlgorithm
  qop: DigestQop
  realm: string
  nonce: string
  opaque: string
  nonceCount: string
  clientNonce: string
  disableRetryRequest: boolean
}

export interface RequestAuth {
  type: AuthType
  bearer: { token: string }
  basic: { username: string; password: string }
  apikey: ApiKeyAuth
  oauth2: OAuth2Auth
  awsv4: AwsV4Auth
  digest: DigestAuth
  unsupported: UnsupportedAuth | null
}

export const DEFAULT_OAUTH2_CALLBACK = 'http://127.0.0.1:53682/callback'

export function emptyAuth(): RequestAuth {
  return {
    type: 'inherit',
    bearer: { token: '' },
    basic: { username: '', password: '' },
    apikey: { key: '', value: '', in: 'header' },
    oauth2: {
      grantType: 'authorization_code',
      tokenName: '',
      redirectUri: DEFAULT_OAUTH2_CALLBACK,
      authUrl: '',
      accessTokenUrl: '',
      refreshTokenUrl: '',
      clientId: '',
      clientSecret: '',
      username: '',
      password: '',
      scope: '',
      state: '',
      challengeAlgorithm: 'S256',
      codeVerifier: '',
      clientAuthentication: 'header',
      addTokenTo: 'header',
      headerPrefix: 'Bearer',
      useToken: 'access_token',
      audience: '',
      resource: '',
      autoRefresh: true,
      authRequestParams: [],
      tokenRequestParams: [],
      refreshRequestParams: [],
      usesPostmanCallback: false,
      suggestedCallback: '',
    },
    awsv4: { accessKey: '', secretKey: '', sessionToken: '', region: '', service: '', addAuthDataToQuery: false },
    digest: { username: '', password: '', algorithm: 'MD5', qop: '', realm: '', nonce: '', opaque: '', nonceCount: '', clientNonce: '', disableRetryRequest: false },
    unsupported: null,
  }
}

/** Whether a grant is kept for compatibility only (RFC 9700). */
export function isLegacyGrant(grant: OAuth2Grant): boolean {
  return grant === 'implicit' || grant === 'password_credentials'
}

/** Whether a grant signs in through the system browser. */
export function usesBrowser(grant: OAuth2Grant): boolean {
  return grant === 'authorization_code' || grant === 'authorization_code_with_pkce' || grant === 'implicit'
}

export interface RequestBody {
  mode: BodyMode
  raw: { language: RawLanguage; content: string }
  formdata: FormDataRow[]
  urlencoded: KeyValueRow[]
  binary: { src: string; srcState: FileState }
  graphql: { query: string; variables: string }
}

export interface RequestSettings {
  timeoutMs: number
  followRedirects: boolean
  verifySsl: boolean
  encodeUrl: boolean
  disableCookieJar: boolean
}

export type ScriptPhase = 'preRequest' | 'postResponse'

/** A request's, folder's or collection's scripts; `disabled` lists the phases imported as disabled. */
export interface ScriptSet {
  preRequest: string
  postResponse: string
  disabled: ScriptPhase[]
}

export interface RequestDraft {
  id: string
  name: string
  description: string
  method: string
  url: string
  params: KeyValueRow[]
  pathVariables: PathVariableRow[]
  headers: KeyValueRow[]
  disabledAutoHeaders: string[]
  auth: RequestAuth
  body: RequestBody
  settings: RequestSettings
  scripts: ScriptSet
  inheritScripts: boolean
}

export interface AutoHeader {
  key: string
  value: string
  disabled: boolean
}

export interface SendTimings {
  dnsMs: number
  connectMs: number
  tlsMs: number
  firstByteMs: number
  downloadMs: number
  totalMs: number
}

export interface SendFailure {
  kind: SendErrorKind
  message: string
  hint: string
}

export type TestStatus = 'passed' | 'failed' | 'skipped'

export interface TestResult {
  index: number
  name: string
  status: TestStatus
  errorMessage: string
  /** The script the test came from, e.g. "Billing · Post-response". */
  scriptLabel: string
  async: boolean
}

export interface TestSummary {
  total: number
  passed: number
  failed: number
  skipped: number
}

export interface ScriptFailure {
  kind: string
  message: string
  line: number | null
  column: number | null
  scriptLabel: string
}

export interface SendResult {
  tabId: string
  sendId: string
  outcome: SendOutcome
  tests: TestResult[]
  testSummary: TestSummary
  scriptErrors: ScriptFailure[]
  ok: boolean
  statusCode: number
  statusText: string
  proto: string
  headers: Array<{ key: string; value: string }>
  /** The cookies the final response set (the response Cookies tab). */
  cookiesSet: Cookie[]
  sizeBytes: number
  contentType: string
  isBinary: boolean
  bodyUrl: string
  inlineBody: string | null
  timings: SendTimings
  error: SendFailure | null
}

export function emptyDraft(): RequestDraft {
  return {
    id: '',
    name: 'Untitled Request',
    description: '',
    method: 'GET',
    url: '',
    params: [],
    pathVariables: [],
    headers: [],
    disabledAutoHeaders: [],
    auth: emptyAuth(),
    body: {
      mode: 'none',
      raw: { language: 'json', content: '' },
      formdata: [],
      urlencoded: [],
      binary: { src: '', srcState: 'ok' },
      graphql: { query: '', variables: '' },
    },
    settings: { timeoutMs: 0, followRedirects: true, verifySsl: true, encodeUrl: true, disableCookieJar: false },
    scripts: emptyScripts(),
    inheritScripts: true,
  }
}

export function emptyScripts(): ScriptSet {
  return { preRequest: '', postResponse: '', disabled: [] }
}

const SCRIPT_PHASES: ReadonlySet<string> = new Set<ScriptPhase>(['preRequest', 'postResponse'])
const TEST_STATUSES: ReadonlySet<string> = new Set<TestStatus>(['passed', 'failed', 'skipped'])

export function toScriptSet(dto: ScriptsDto | null | undefined): ScriptSet {
  if (!dto) {
    return emptyScripts()
  }
  return {
    preRequest: dto.pre_request,
    postResponse: dto.post_response,
    disabled: (dto.disabled ?? []).filter((phase): phase is ScriptPhase => SCRIPT_PHASES.has(phase)),
  }
}

export function toScriptsDto(scripts: ScriptSet): ScriptsDto {
  return { pre_request: scripts.preRequest, post_response: scripts.postResponse, disabled: [...scripts.disabled] }
}

function toTestResult(dto: TestResultDto): TestResult {
  return {
    index: dto.index,
    name: dto.name,
    status: oneOf<TestStatus>(TEST_STATUSES, dto.status, 'failed'),
    errorMessage: dto.error_message ?? '',
    scriptLabel: dto.script.label,
    async: dto.async,
  }
}

export function toScriptFailure(dto: ScriptErrorDto): ScriptFailure {
  return { kind: dto.kind, message: dto.message, line: dto.line ?? null, column: dto.column ?? null, scriptLabel: dto.script.label }
}

function oneOf<T extends string>(values: ReadonlySet<string>, value: string | undefined, fallback: T): T {
  return values.has(value ?? '') ? (value as T) : fallback
}

export function toKeyValueRow(dto: KeyValueDto): KeyValueRow {
  return { id: newRowId(), key: dto.key, value: dto.value, description: dto.description, enabled: dto.enabled ?? true }
}

export function toKeyValueDto(row: KeyValueRow): KeyValueDto {
  return { key: row.key, value: row.value, description: row.description, enabled: row.enabled }
}

function toPathVariableRow(dto: PathVariableDto): PathVariableRow {
  return { id: newRowId(), key: dto.key, value: dto.value, description: dto.description }
}

function toFileState(state: string | undefined): FileState {
  return state === 'repick' ? 'repick' : 'ok'
}

function toFormDataRow(dto: FormDataRowDto): FormDataRow {
  return { ...toKeyValueRow(dto), type: dto.type === 'file' ? 'file' : 'text', src: dto.src, srcState: toFileState(dto.src_state) }
}

/** A named version of a saved request: its draft carries the request's id and the alias's name. */
export interface SavedAlias {
  id: string
  name: string
  draft: RequestDraft
}

export interface SavedRequest {
  draft: RequestDraft
  revision: string
  aliases: SavedAlias[]
}

const GRANT_SET: ReadonlySet<string> = new Set(OAUTH2_GRANTS)
const DIGEST_ALGORITHM_SET: ReadonlySet<string> = new Set(DIGEST_ALGORITHMS)
const PARAM_TARGETS: ReadonlySet<string> = new Set<OAuth2ParamTarget>(['request_url', 'request_body', 'request_header'])

function toOAuth2Params(dtos: OAuth2ParamDto[] | null | undefined, fallback: OAuth2ParamTarget): OAuth2Param[] {
  return (dtos ?? []).map((dto) => ({ key: dto.key, value: dto.value, enabled: dto.enabled, sendAs: oneOf<OAuth2ParamTarget>(PARAM_TARGETS, dto.send_as, fallback) }))
}

function toOAuth2ParamDtos(params: OAuth2Param[]): OAuth2ParamDto[] {
  return params.map((param) => ({ key: param.key, value: param.value, enabled: param.enabled, send_as: param.sendAs }))
}

function toOAuth2(dto: AuthOAuth2Dto | null | undefined): OAuth2Auth {
  const defaults = emptyAuth().oauth2
  if (!dto) {
    return defaults
  }
  return {
    grantType: oneOf<OAuth2Grant>(GRANT_SET, dto.grant_type, defaults.grantType),
    tokenName: dto.token_name,
    redirectUri: dto.redirect_uri || defaults.redirectUri,
    authUrl: dto.auth_url,
    accessTokenUrl: dto.access_token_url,
    refreshTokenUrl: dto.refresh_token_url,
    clientId: dto.client_id,
    clientSecret: dto.client_secret,
    username: dto.username,
    password: dto.password,
    scope: dto.scope,
    state: dto.state,
    challengeAlgorithm: dto.challenge_algorithm === 'plain' ? 'plain' : 'S256',
    codeVerifier: dto.code_verifier,
    clientAuthentication: dto.client_authentication === 'body' ? 'body' : 'header',
    addTokenTo: dto.add_token_to === 'queryParams' ? 'queryParams' : 'header',
    headerPrefix: dto.header_prefix,
    useToken: dto.use_token === 'id_token' ? 'id_token' : 'access_token',
    audience: dto.audience,
    resource: dto.resource,
    autoRefresh: dto.auto_refresh,
    authRequestParams: toOAuth2Params(dto.auth_request_params, 'request_url'),
    tokenRequestParams: toOAuth2Params(dto.token_request_params, 'request_body'),
    refreshRequestParams: toOAuth2Params(dto.refresh_request_params, 'request_body'),
    usesPostmanCallback: dto.uses_postman_callback ?? false,
    suggestedCallback: dto.suggested_callback ?? '',
  }
}

export function toOAuth2Dto(oauth2: OAuth2Auth): AuthOAuth2Dto {
  return {
    grant_type: oauth2.grantType,
    token_name: oauth2.tokenName,
    redirect_uri: oauth2.redirectUri,
    auth_url: oauth2.authUrl,
    access_token_url: oauth2.accessTokenUrl,
    refresh_token_url: oauth2.refreshTokenUrl,
    client_id: oauth2.clientId,
    client_secret: oauth2.clientSecret,
    username: oauth2.username,
    password: oauth2.password,
    scope: oauth2.scope,
    state: oauth2.state,
    challenge_algorithm: oauth2.challengeAlgorithm,
    code_verifier: oauth2.codeVerifier,
    client_authentication: oauth2.clientAuthentication,
    add_token_to: oauth2.addTokenTo,
    header_prefix: oauth2.headerPrefix,
    use_token: oauth2.useToken,
    audience: oauth2.audience,
    resource: oauth2.resource,
    auto_refresh: oauth2.autoRefresh,
    auth_request_params: toOAuth2ParamDtos(oauth2.authRequestParams),
    token_request_params: toOAuth2ParamDtos(oauth2.tokenRequestParams),
    refresh_request_params: toOAuth2ParamDtos(oauth2.refreshRequestParams),
  }
}

export function toAuth(dto: AuthDto): RequestAuth {
  const defaults = emptyAuth()
  return {
    type: oneOf<AuthType>(AUTH_TYPE_SET, dto.type, 'inherit'),
    bearer: { token: dto.bearer.token },
    basic: { username: dto.basic.username, password: dto.basic.password },
    apikey: dto.apikey ? { key: dto.apikey.key, value: dto.apikey.value, in: dto.apikey.in === 'query' ? 'query' : 'header' } : defaults.apikey,
    oauth2: toOAuth2(dto.oauth2),
    awsv4: dto.awsv4
      ? {
          accessKey: dto.awsv4.access_key,
          secretKey: dto.awsv4.secret_key,
          sessionToken: dto.awsv4.session_token,
          region: dto.awsv4.region,
          service: dto.awsv4.service,
          addAuthDataToQuery: dto.awsv4.add_auth_data_to_query,
        }
      : defaults.awsv4,
    digest: dto.digest
      ? {
          username: dto.digest.username,
          password: dto.digest.password,
          algorithm: oneOf<DigestAlgorithm>(DIGEST_ALGORITHM_SET, dto.digest.algorithm, 'MD5'),
          qop: dto.digest.qop === 'auth' || dto.digest.qop === 'auth-int' ? dto.digest.qop : '',
          realm: dto.digest.realm,
          nonce: dto.digest.nonce,
          opaque: dto.digest.opaque,
          nonceCount: dto.digest.nonce_count,
          clientNonce: dto.digest.client_nonce,
          disableRetryRequest: dto.digest.disable_retry_request,
        }
      : defaults.digest,
    unsupported: dto.unsupported ? { source: dto.unsupported.source, type: dto.unsupported.type } : null,
  }
}

export function toAuthDto(auth: RequestAuth): AuthDto {
  return {
    type: auth.type,
    bearer: { token: auth.bearer.token },
    basic: { username: auth.basic.username, password: auth.basic.password },
    apikey: { key: auth.apikey.key, value: auth.apikey.value, in: auth.apikey.in },
    oauth2: toOAuth2Dto(auth.oauth2),
    awsv4: {
      access_key: auth.awsv4.accessKey,
      secret_key: auth.awsv4.secretKey,
      session_token: auth.awsv4.sessionToken,
      region: auth.awsv4.region,
      service: auth.awsv4.service,
      add_auth_data_to_query: auth.awsv4.addAuthDataToQuery,
    },
    digest: {
      username: auth.digest.username,
      password: auth.digest.password,
      algorithm: auth.digest.algorithm,
      qop: auth.digest.qop,
      realm: auth.digest.realm,
      nonce: auth.digest.nonce,
      opaque: auth.digest.opaque,
      nonce_count: auth.digest.nonceCount,
      client_nonce: auth.digest.clientNonce,
      disable_retry_request: auth.digest.disableRetryRequest,
    },
  }
}

export function toSavedRequest(dto: RequestDto): SavedRequest {
  return { draft: toRequestDraft(dto), revision: dto.revision, aliases: (dto.aliases ?? []).map((alias) => toSavedAlias(dto.id, alias)) }
}

function toSavedAlias(requestId: string, dto: RequestAliasDto): SavedAlias {
  return { id: dto.id, name: dto.name, draft: { ...toRequestDraft(dto.request), id: requestId, name: dto.name } }
}

/** The owner of an alias's own auth (its OAuth 2.0 token), as Go names it: `<request id>:<alias id>`. */
export function aliasAuthOwner(requestId: string, aliasId: string): string {
  return `${requestId}:${aliasId}`
}

export function toRequestAliasDto(aliasId: string, draft: RequestDraft): RequestAliasDto {
  return { id: aliasId, name: draft.name, request: toRequestDto(draft) }
}

export function toRequestDraft(dto: RequestDto): RequestDraft {
  const fallback = emptyDraft()
  return {
    id: dto.id,
    name: dto.name || fallback.name,
    description: dto.description,
    method: dto.method || fallback.method,
    url: dto.url,
    params: (dto.params ?? []).map(toKeyValueRow),
    pathVariables: (dto.path_variables ?? []).map(toPathVariableRow),
    headers: (dto.headers ?? []).map(toKeyValueRow),
    disabledAutoHeaders: dto.auto_headers.disabled ?? [],
    auth: toAuth(dto.auth),
    body: {
      mode: oneOf<BodyMode>(BODY_MODE_SET, dto.body.mode, 'none'),
      raw: { language: oneOf<RawLanguage>(RAW_LANGUAGE_SET, dto.body.raw.language, 'json'), content: dto.body.raw.content },
      formdata: (dto.body.formdata ?? []).map(toFormDataRow),
      urlencoded: (dto.body.urlencoded ?? []).map(toKeyValueRow),
      binary: { src: dto.body.binary.src, srcState: toFileState(dto.body.binary.src_state) },
      graphql: { query: dto.body.graphql.query, variables: dto.body.graphql.variables },
    },
    settings: {
      timeoutMs: dto.settings.timeout_ms,
      followRedirects: dto.settings.follow_redirects ?? true,
      verifySsl: dto.settings.verify_ssl ?? true,
      encodeUrl: dto.settings.encode_url ?? true,
      disableCookieJar: dto.settings.disable_cookie_jar ?? false,
    },
    scripts: toScriptSet(dto.scripts),
    inheritScripts: dto.inherit_scripts ?? true,
  }
}

export function toRequestDto(draft: RequestDraft): RequestDto {
  return {
    id: draft.id,
    name: draft.name,
    description: draft.description,
    method: draft.method,
    url: draft.url,
    revision: '',
    aliases: [],
    params: draft.params.map(toKeyValueDto),
    path_variables: draft.pathVariables.map(({ key, value, description }) => ({ key, value, description })),
    headers: draft.headers.map(toKeyValueDto),
    auto_headers: { disabled: draft.disabledAutoHeaders },
    auth: toAuthDto(draft.auth),
    body: {
      mode: draft.body.mode,
      raw: { language: draft.body.raw.language, content: draft.body.raw.content },
      formdata: draft.body.formdata.map((row) => ({ ...toKeyValueDto(row), type: row.type, src: row.src, src_state: row.srcState })),
      urlencoded: draft.body.urlencoded.map(toKeyValueDto),
      binary: { src: draft.body.binary.src, src_state: draft.body.binary.srcState },
      graphql: { query: draft.body.graphql.query, variables: draft.body.graphql.variables },
    },
    settings: {
      timeout_ms: draft.settings.timeoutMs,
      follow_redirects: draft.settings.followRedirects,
      verify_ssl: draft.settings.verifySsl,
      encode_url: draft.settings.encodeUrl,
      disable_cookie_jar: draft.settings.disableCookieJar,
    },
    scripts: toScriptsDto(draft.scripts),
    inherit_scripts: draft.inheritScripts,
  }
}

export function toAutoHeader(dto: AutoHeaderDto): AutoHeader {
  return { key: dto.key, value: dto.value, disabled: dto.disabled }
}

export function toSendResult(dto: SendResultDto): SendResult {
  return {
    tabId: dto.tab_id,
    sendId: dto.send_id,
    outcome: oneOf<SendOutcome>(SEND_OUTCOME_SET, dto.outcome, 'sent'),
    tests: (dto.tests ?? []).map(toTestResult),
    testSummary: { total: dto.test_summary.total, passed: dto.test_summary.passed, failed: dto.test_summary.failed, skipped: dto.test_summary.skipped },
    scriptErrors: (dto.script_errors ?? []).map(toScriptFailure),
    ok: dto.ok,
    statusCode: dto.status_code,
    statusText: dto.status_text,
    proto: dto.proto,
    headers: (dto.headers ?? []).map(({ key, value }) => ({ key, value })),
    cookiesSet: (dto.cookies_set ?? []).map(toCookie),
    sizeBytes: dto.size_bytes,
    contentType: dto.content_type,
    isBinary: dto.is_binary,
    bodyUrl: dto.body_url,
    inlineBody: dto.inline_body ?? null,
    timings: {
      dnsMs: dto.timings.dns_ms,
      connectMs: dto.timings.connect_ms,
      tlsMs: dto.timings.tls_ms,
      firstByteMs: dto.timings.first_byte_ms,
      downloadMs: dto.timings.download_ms,
      totalMs: dto.timings.total_ms,
    },
    error: dto.error ? { kind: oneOf<SendErrorKind>(SEND_ERROR_KIND_SET, dto.error.kind, 'other'), message: dto.error.message, hint: dto.error.hint } : null,
  }
}

export function cancelledResult(tabId: string): SendResult {
  return {
    tabId,
    sendId: '',
    outcome: 'cancelled',
    tests: [],
    testSummary: { total: 0, passed: 0, failed: 0, skipped: 0 },
    scriptErrors: [],
    ok: false,
    statusCode: 0,
    statusText: '',
    proto: '',
    headers: [],
    cookiesSet: [],
    sizeBytes: 0,
    contentType: '',
    isBinary: false,
    bodyUrl: '',
    inlineBody: null,
    timings: { dnsMs: 0, connectMs: 0, tlsMs: 0, firstByteMs: 0, downloadMs: 0, totalMs: 0 },
    error: { kind: 'cancelled', message: 'Request cancelled', hint: '' },
  }
}

/** What the user already agreed to for this send. */
export interface SendAcknowledgement {
  confirmedEnvironmentId: string
  missingSecrets: boolean
  /** Send with an OAuth 2.0 token that expired and could not be refreshed ("Send anyway"). */
  expiredToken: boolean
}

export function noAcknowledgement(): SendAcknowledgement {
  return { confirmedEnvironmentId: '', missingSecrets: false, expiredToken: false }
}
