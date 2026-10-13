# Contract: Bruno → Xenon mapping (.bru and OpenCollection YAML)

Bruno-only fields with no Xenon place go to a `bruno:` passthrough block.

## 1. Structure
| Bruno (.bru / YAML) | Xenon |
|---|---|
| `bruno.json` `name` / `opencollection.yml` name | collection `name` |
| `collection.bru` / root collection file: `headers`, `auth`, `script:*`, `tests`, `vars:*`, `docs` | collection `headers`, `auth`, `scripts`, description; `vars:pre-request` → collection variables (initial values) |
| directory + `folder.bru` / `folder.yml` (`meta.name`, `meta.seq`) | folder (`name`, order by seq) |
| `*.bru` / `*.yml` request (`meta`/`info`: name, seq, type http\|graphql) | request; order by seq then file name |
| `environments/*.bru|yml` | environments |
| `meta.tags`, `settings` keys without Xenon equivalent | `bruno:` passthrough |

## 2. Request fields
| Bruno | Xenon |
|---|---|
| method block (`get`…`trace`, `http` custom) / `http.method` | `method` |
| `url` | `url` |
| `params:query` / `params[type=query]` | `params` (`~` / `disabled` → enabled false) |
| `params:path` / `params[type=path]` | `pathVariables` |
| `headers` | `headers` |
| `body:json` / `text` / `xml` | `body.raw` language json/text/xml |
| `body:sparql` | `body.raw` text (+ passthrough note) |
| `body:form-urlencoded` | `body.urlencoded` |
| `body:multipart-form` (`@file(path)`, `@contentType(...)`) | `body.formdata` (file rows, `src` with `srcState: repick`; content type → passthrough) |
| `body:file` | `body.binary` (`srcState: repick`) |
| `body:graphql` + `body:graphql:vars` | `body.graphql` |
| `auth: inherit\|none` | `inherit` / `none` |
| `auth:basic`, `bearer`, `apikey`, `digest`, `oauth2`, `awsv4` | same Xenon auth types |
| `auth:ntlm`, `oauth1`, `wsse`, `akamai-edgegrid` | `unsupported` (source bruno) |
| `docs` | `description` |
| `settings` `encodeUrl`, `timeout`, `followRedirects` | `settings.encodeUrl`, `timeoutMs`, `followRedirects` (`maxRedirects` → passthrough) |
| `script:pre-request`, `script:post-response` / `runtime.scripts[before-request|after-response]` | `scripts.preRequest` / `postResponse` (translated, §3) |
| `tests` / `runtime.scripts[tests]` | appended to `scripts.postResponse` (translated) |
| `vars:pre-request` | prepended to pre-request script as `pm.variables.set("k", "v");` |
| `vars:post-response` (expressions) | appended to post-response script as `pm.variables.set("k", <translated expr>);` |
| `assert` / `runtime.assertions` | appended to post-response script as `pm.test(...)` (§4) |
| `body:grpc`, `body:ws`, request `type` other than http/graphql | request skipped, listed in summary |

## 3. Script translation (applied token-wise, longest match first)
| Bruno | pm.* |
|---|---|
| `bru.getEnvVar(k)` / `bru.setEnvVar(k, v)` | `pm.environment.get(k)` / `pm.environment.set(k, v)` |
| `bru.getVar(k)` / `bru.setVar(k, v)` | `pm.variables.get(k)` / `pm.variables.set(k, v)` |
| `bru.getRequestVar(k)` | `pm.variables.get(k)` |
| `bru.getCollectionVar(k)` | `pm.collectionVariables.get(k)` |
| `bru.getGlobalEnvVar(k)` / `bru.setGlobalEnvVar(k, v)` | `pm.globals.get(k)` / `pm.globals.set(k, v)` |
| `bru.getEnvName()` | `pm.environment.name` |
| `bru.interpolate(s)` | `pm.variables.replaceIn(s)` |
| `req.getUrl()` / `req.url` | `pm.request.url.toString()` |
| `req.getMethod()` / `req.method` | `pm.request.method` |
| `req.getHeader(n)` | `pm.request.headers.get(n)` |
| `req.setHeader(n, v)` | `pm.request.headers.upsert({ key: n, value: v })` |
| `req.getHeaders()` / `req.headers` | `pm.request.headers.toObject()` |
| `req.getBody()` / `req.body` | `JSON.parse(pm.request.body.raw)` |
| `req.setBody(b)` | `pm.request.body.update(JSON.stringify(b))` |
| `res.getStatus()` / `res.status` | `pm.response.code` |
| `res.getBody()` / `res.body` | `pm.response.json()` |
| `res.getHeader(n)` | `pm.response.headers.get(n)` |
| `res.getHeaders()` / `res.headers` | `pm.response.headers.toObject()` |
| `res.getResponseTime()` / `res.responseTime` | `pm.response.responseTime` |
| `test(name, fn)` | `pm.test(name, fn)` |
| `expect(x)` | `pm.expect(x)` |
Not translated (flagged): `bru.setNextRequest`, `bru.runner.*`, `bru.sleep`, `bru.sendRequest`
(callback signature differs), `bru.cookies.*`, `req.setUrl`, `req.setMethod`, `req.setTimeout`,
`req.setMaxRedirects`, `res.setBody`, `require(...)` of npm modules, and any line still containing
`bru.`, `req.` or `res.` after translation.

## 4. Assertions → `pm.test`
`expr op value` becomes `pm.test("<expr> <op> <value>", () => pm.expect(<translated expr>)<chai>);`
| op | chai | op | chai |
|---|---|---|---|
| `eq` | `.to.eql(v)` | `neq` | `.to.not.eql(v)` |
| `gt`/`gte`/`lt`/`lte` | `.to.be.above/at.least/below/at.most(v)` | `in` / `notIn` | `.to.be.oneOf(v)` / `.to.not.be.oneOf(v)` |
| `contains` / `notContains` | `.to.include(v)` / `.to.not.include(v)` | `length` | `.to.have.lengthOf(v)` |
| `matches` / `notMatches` | `.to.match(/v/)` / `.to.not.match(/v/)` | `startsWith` / `endsWith` | `.to.match(/^v/)` / `.to.match(/v$/)` (escaped) |
| `between` | `.to.be.within(a, b)` | `isEmpty` / `isNotEmpty` | `.to.be.empty` / `.to.not.be.empty` |
| `isNull` / `isUndefined` / `isDefined` | `.to.be.null` / `.to.be.undefined` / `.to.not.be.undefined` | `isTruthy` / `isFalsy` | `.to.be.ok` / `.to.not.be.ok` |
| `isNumber` / `isString` / `isBoolean` / `isArray` | `.to.be.a("number"/"string"/"boolean"/"array")` | `isJson` | `.to.be.an("object")` |
Expressions (`res.status`, `res.body.id`) use §3. Values are literals as written by Bruno
(numbers/booleans unquoted, strings quoted).

## 5. Environments
| Bruno | Xenon |
|---|---|
| `vars { k: v }` / `~k: v` | plain rows (enabled false for `~`) |
| `vars:secret [ k, ~k2 ]` | secret rows, missing (values are not in the collection) |
| `color` | environment colour mapped to nearest neon (green/cyan/magenta/yellow/red) or none |
