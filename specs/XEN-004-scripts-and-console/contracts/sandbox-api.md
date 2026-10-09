# Contract: script sandbox surface

The exact surface a script sees. Parity reference: postman-sandbox 6.7.4. Every row is covered by
the parity corpus (research R7).

## 1. `pm`
| Member | Behaviour |
|---|---|
| `pm.variables` | local scope for the current send: `has, get, set, replaceIn, toObject`; `get` resolves through all scopes (local → environment → collection → globals) |
| `pm.environment` | active environment: `name, has, get, set, unset, clear, replaceIn, toObject`; writes = current values (XEN-002); no active env → no-op + console warning |
| `pm.collectionVariables` | request's collection: `has, get, set, unset, clear, replaceIn, toObject` |
| `pm.globals` | workspace globals: same methods |
| `pm.iterationData` | always empty: `has, get, unset, toObject, toJSON` |
| `pm.request` | Request model: `url` (Url: `toString, getHost, getPath, getPathWithQuery, getQueryString, getRaw, getRemote, update, addQueryParams, removeQueryParams, query`), `method`, `headers` (HeaderList/PropertyList: `add, upsert, remove, get, has, one, all, each, filter, find, map, reduce, count, clear, toObject, insert, prepend, append, populate, repopulate, idx, indexOf`), `body` (`mode, raw, urlencoded, formdata, file, graphql, update, isEmpty, toString, toJSON`), `addHeader, upsertHeader, removeHeader, addQueryParams, removeQueryParams, update, clone, size, toJSON`, `auth`; `pm.request.to` → `chai.expect(pm.request).to`. Mutable in pre-request only (post-response: read-only snapshot of what was sent) |
| `pm.response` | post-response only: `code, status, headers, responseTime, responseSize, cookies, json(), jsonp(), text(), reason(), details(), size(), mime(), contentInfo(), dataURI(), encoding()`, `to` |
| `pm.info` | `eventName` (`prerequest`\|`test`), `iteration` 0, `iterationCount` 1, `requestName`, `requestId` |
| `pm.cookies` | response cookies: `has, get, toObject` (PropertyList); `jar()` → stage 5 jar (`get, getAll, set, unset, clear`, callback style) |
| `pm.sendRequest(req[, cb])` | `req`: URL string or request definition (Postman shape); `{{vars}}` resolved; cb `(err, res, history)` or Promise<Response>; console-logged; not in History |
| `pm.execution` | `skipRequest()` (pre-request), `runRequest(id[, opts])`, `location` (+ `.current`), `setNextRequest()` no-op |
| `pm.test(name, fn)` | sync; `fn.length > 0` → async with `done(err?)`; `pm.test.skip(name)`; `pm.test.index()`; returns `pm` |
| `pm.expect` | `chai.expect` (chai 4.4.1) |
| `pm.vault` | `get/set/unset` (promises) on workspace globals' secret rows; `{{vault:name}}` resolves the same |
| `pm.visualizer` | `set/clear` no-op (info entry in console once per script) |
| `pm.require`, `pm.datasets`, `pm.mock`, `pm.state`, `pm.message` | throw `"<api> is not supported in Xenon"` |

## 2. Response assertions (`pm.response.to…`, ported from chai-postman 2.0.1)
Properties: `ok` (code 200 and reason OK), `error` (4xx/5xx), `json`, `withBody`, `info`,
`success`, `redirection`, `clientError`, `serverError`, `accepted`, `withoutContent`,
`badRequest`, `unauthorized`/`unauthorised`, `forbidden`, `notFound`, `notAcceptable`,
`rateLimited`, `postmanRequest`, `postmanResponse`, `postmanRequestOrResponse`.
Methods: `status(code|reason)`, `statusCode`, `statusCodeClass`, `statusReason`,
`header(key[, value])`, `body(val)`, `jsonBody([path][, val])`, `jsonSchema(schema[, ajvOpts])`.
Chainable: `responseTime`, `responseSize`. Failure messages identical to chai-postman.

## 3. Globals
`console` (`log, info, warn, error, debug, clear`), `setTimeout/setInterval/setImmediate` (+
clear), `atob`, `btoa`, `Buffer`, `URL`, `URLSearchParams`, `TextEncoder`, `TextDecoder`,
`structuredClone`, `queueMicrotask`, `crypto` (`getRandomValues`, `randomUUID`,
`subtle.digest`, `subtle.importKey`+`sign`/`verify` for HMAC), `AbortController`,
`AbortSignal`, `Event`, `EventTarget`, `DOMException`, `Blob`, `Intl` (en-US, research R6),
`require`. Legacy globals in §5.

## 4. `require()` modules
`ajv`, `atob`, `btoa`, `chai`, `cheerio`, `crypto-js`, `csv-parse/lib/sync`, `lodash`, `moment`,
`postman-collection` (Url, Header, PropertyList, Request, Response, Variable, VariableScope),
`tv4`, `uuid`, `xml2js`, `backbone`, `assert`, `buffer`, `events`, `path`, `punycode`,
`querystring`, `string_decoder`, `stream`, `timers`, `url`, `util`. Anything else → `Cannot find
module '<name>'` (Postman's message).

## 5. Legacy API
`tests` (object; collected at script end as test results), `globals`, `environment`, `data`
(empty), `request` (`url, method, headers, data`), `responseCookies`, `responseHeaders`,
`responseTime`, `responseCode` (`{code, name, detail}`), `responseBody` (text), `iteration` (0),
`postman` (`setEnvironmentVariable, getEnvironmentVariable, clearEnvironmentVariable,
clearEnvironmentVariables, setGlobalVariable, getGlobalVariable, clearGlobalVariable,
clearGlobalVariables, setNextRequest` no-op, `getResponseCookie, getResponseHeader`),
`CryptoJS`, `tv4`, `xml2Json(xml)`, `cheerio`, `_` (lodash 3), `Backbone`. First legacy use per
script logs Postman's deprecation notice as a console info entry.

## 6. Error kinds
`script_syntax`, `script_runtime`, `script_timeout`, `script_memory`, `script_stack_overflow`,
`script_unhandled_rejection`, `script_unsupported_api` — each with message, script (owner kind,
name, phase) and line/column when available.
