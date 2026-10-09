# Contract: Postman → Xenon mapping (collection v2.0/v2.1, environment, globals)

"Mapped" keys become Xenon fields; **every other source key goes to the item's `postman:`
passthrough** ([format-additions.md](format-additions.md) §2). Stage 6 export inverts this table.

## Collection (`info` + root)
| Postman | Xenon |
|---|---|
| `info.name` | collection `name` |
| `info.description` (string or `{content}`) | `description` (content; `type`/`version` → passthrough) |
| `info._postman_id`, `schema`, `version`, `_exporter_id`, `_collection_link` | passthrough |
| `item[]` order | `order` + child files |
| `auth` | `auth` (see Auth) |
| `event[]` | `scripts.preRequest` / `scripts.postResponse` (+ metadata → passthrough) |
| `variable[]` | collection `variables` (key, value→initial value as string, `disabled`→`enabled:false`, description); `id`/`type`/`name`/`system` → row passthrough |
| `protocolProfileBehavior` | passthrough (collection has no Xenon settings) |

## Folder (item-group)
| Postman | Xenon |
|---|---|
| `name`, `description`, `item[]` | `name`, `description`, `order` + children |
| `auth`, `event[]` | `auth`, `scripts` |
| `variable[]` | passthrough (folders have no variables) + summary warning |
| `protocolProfileBehavior`, `id` | passthrough |

## Request item
| Postman | Xenon |
|---|---|
| `name`, `description` | `name`, `description` |
| `request` as string | GET with that URL |
| `request.method` | `method` (upper-cased; any token) |
| `request.url.raw` (or rebuilt) | `url` |
| `request.url.query[]` | `params` (key, value, `disabled`→`enabled`, description); null key/value → "" |
| `request.url.variable[]` | `pathVariables` (key, value, description) |
| `request.url.hash`, `port`, `protocol`/`host`/`path` parts | passthrough (raw is authoritative) |
| `request.header[]` / raw string | `headers` (raw string parsed `Name: value` per line, flagged in passthrough) |
| `request.body.mode` raw | `body.raw` (`options.raw.language` → `language`: json/javascript/xml/html/text; unknown → text) |
| `… urlencoded[]` | `body.urlencoded` |
| `… formdata[]` | `body.formdata` (`type` text/file, `value`, `src` first path → `src`; extra `src` paths + `contentType` → passthrough) |
| `… file` | `body.binary.src` (`content` → passthrough) |
| `… graphql` | `body.graphql` (`variables` string kept as text) |
| `body.disabled: true` | `body.mode: none` + original body kept in passthrough |
| `request.auth` | `auth` |
| `event[]` | `scripts` (`listen: prerequest` → preRequest, `test` → postResponse; `exec[]` joined with `\n`; `disabled` → script kept with passthrough flag, not run) |
| `response[]` | `examples` (below) |
| `protocolProfileBehavior.followRedirects` / `strictSSL` / `disableUrlEncoding` | `settings.followRedirects` / `verifySsl` / `encodeUrl` = !disableUrlEncoding (all keys also kept in passthrough) |
| `variable[]`, `proxy`, `certificate`, `id` | passthrough (+ warning for variables) |

## Auth
`auth` absent/null → `{type: inherit}`; `noauth` → `{type: none}`.
| Postman type | Xenon |
|---|---|
| `bearer` (`token`) | `bearer.token` |
| `basic` (`username`, `password`) | `basic` |
| `apikey` (`key`, `value`, `in: header\|query`) | `apikey` |
| `oauth2` (all attributes) | `oauth2` sub-object (attributes by key; stage 5 defines semantics) |
| `awsv4` (`accessKey`, `secretKey`, `region`, `service`, `sessionToken`) | `awsv4` |
| `digest` (`username`, `password`, `realm`, `nonce`, `algorithm`, `qop`, `nc`, `cnonce`, `opaque`) | `digest` |
| `hawk`, `ntlm`, `edgegrid`, `oauth1`, `jwt`, `asap`, unknown | `{type: unsupported, unsupported: {source: postman, type, data}}` |
v2.0 attribute maps (`{"basic": {"username": …}}`) are normalised to v2.1 lists first. Attribute
`type` fields go to passthrough.

## Example (`response[]` item)
| Postman | Xenon `examples[]` |
|---|---|
| `name` | `name` |
| `originalRequest` | `request` (same mapping as a request) |
| `code`, `status` | `response.code`, `response.status` |
| `header[]` | `response.headers` |
| `body` | `response.body` (text; binary not representable in v2.1 JSON) |
| `_postman_previewlanguage`, `cookie[]`, `responseTime`, `timings`, `id` | example passthrough |
| Content type | from `Content-Type` header, else preview language |

## Environment / globals
| Postman | Xenon |
|---|---|
| `name` | environment `name` (globals: merged into `workspace.yaml` `variables`) |
| `values[]` `key`, `enabled` | row `key`, `enabled` |
| `type: default\|text\|any` + `value` | plain row, initial value |
| `type: secret` + non-empty `value` | secret row; value → vault (current value) |
| `type: secret` + empty `value` | secret row, missing |
| `id`, `_postman_variable_scope`, `_postman_exported_at`, `_postman_exported_using` | environment passthrough |
New environments get XEN-002 defaults for colour and confirm-before-send (prod rule).
