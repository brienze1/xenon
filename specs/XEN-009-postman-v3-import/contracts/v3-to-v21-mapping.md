# Contract: Postman v3 → Postman v2.1 document mapping

The v3 DTOs build a v2.1 document, which then runs through the existing v2.1 import (research R2). This table is the
whole translation. Keys not listed are copied into the v2.1 object at the same level under their v3 name, so the
existing passthrough keeps them for export (R6). `{…}` means "same object, translated recursively".

## Normalisation applied first (both JSON and YAML)

| v3 input | Normalised to |
|---|---|
| `headers`, `response.headers`, `auth.credentials` as a YAML map `{k: v}` | list `[{key: k, value: v}]` in map order as decoded (yaml.v3 `MappingNode` order, i.e. file order) |
| `auth.credentials` as list `[{key, value}]` | unchanged |
| `queryParams[].value: null` | `""` |
| empty values in `settings` (`{}`, `[]`, e.g. `disabledSystemHeaders: []`) | dropped; `settings` with nothing left emits no `protocolProfileBehavior` |
| empty `scripts: []`, `headers: []`, `queryParams: []`, `variables: []` | emitted as empty v2.1 lists (mapped fields, no passthrough effect) |
| collection/folder `auth` list `[a1, a2, …]` | `a1`. If the list has 2+ entries: one `unsupported_auth` warning (R8), and the whole list is kept under `postmanV3.auth` |
| folder/collection order (YAML) | R10: by `order` asc, then name |

## Collection

| v3 | v2.1 |
|---|---|
| `id` | `info._postman_id` |
| `name` (or folder name, R11) | `info.name` |
| `description` | `info.description` |
| — | `info.schema = "https://schema.getpostman.com/json/collection/v2.1.0/collection.json"` |
| `variables[] {key, value, disabled, description, secret}` | `variable[] {key, value, disabled, description, type: "secret" if secret}`; a secret whose value is `********` gets `value: ""` |
| `auth` | `auth` (see Auth) |
| `scripts[]` | `event[]` (see Scripts) |
| `items[]` (JSON) / subfolders + `*.request.yaml` (YAML) | `item[]` |

## Folder (`$kind: collection` inside `items`, or a subfolder)

| v3 | v2.1 item |
|---|---|
| `id`, `name`, `description` | `id`, `name`, `description` |
| `auth`, `scripts` | `auth`, `event` |
| `variables` | `variable`. The v2.1 `folder_variables` warning fires. |
| children | `item[]`. Always present, even when empty, so v2.1 sees a folder. |

## Request (`$kind: http-request`)

| v3 | v2.1 item / item.request |
|---|---|
| `id`, `name` (R11) | `id`, `name` |
| `method` | `request.method` |
| `url` | `request.url.raw` |
| `queryParams[] {key, value, disabled, description}` | `request.url.query[]` (same keys) |
| `pathVariables[] {key, value, description}` | `request.url.variable[]` |
| `headers[] {key, value, disabled, description}` | `request.header[]` |
| `body` | `request.body` (see Body) |
| `auth` absent | no `auth` key (→ Inherit) |
| `auth {type: noauth}` | `request.auth {type: "noauth"}` |
| `auth {type, credentials}` | `request.auth` (see Auth) |
| `description` | `request.description` |
| `scripts[]` | `event[]` |
| `settings {followRedirects, strictSSL, maxRedirects, protocolVersion, disabledSystemHeaders, …}` | `protocolProfileBehavior {…}` (same keys, empty values dropped; omitted when empty) |
| `examples[]` (JSON inline) / example files (YAML) | `response[]` (see Example). A YAML `examples` path is relative to the request file and uses `/`; `./` and a trailing `/` are optional. |
| `variables` (if present) | `variable`. The v2.1 `request_variables` warning fires. |

## Request of another protocol

| v3 `$kind` | Result |
|---|---|
| `graphql-request` | v2.1 request `method: "POST"`, `url.raw = url`, `header = headers`, `body {mode: "graphql", graphql {query, variables}}`, `auth`. Scripts `beforeQuery` → `prerequest` and `afterResponse` → `test`. `settings` → `protocolProfileBehavior`. |
| `grpc-request`, `websocket-request`, `socket.io-request`, `mqtt-request`, `mcp-request`, `llm-request`, and any other `$kind` ending in `-request` | Not emitted. One `unsupported_request_type` warning: `<name> is a <kind> request, which Xenon does not support; it was skipped` |

## Body

| v3 `body.type` | v2.1 `body` |
|---|---|
| `json` / `text` / `xml` / `html` / `javascript` | `{mode: "raw", raw: content, options: {raw: {language: json|text|xml|html|javascript}}}` |
| `formdata` | `{mode: "formdata", formdata: [{key, type: text|file, value | src, disabled, contentType, description}]}` |
| `urlencoded` | `{mode: "urlencoded", urlencoded: [{key, value, disabled, description}]}` |
| `graphql` | `{mode: "graphql", graphql: {query, variables}}` |
| `file` | `{mode: "file", file: {src: content.src}}` |
| `none` / absent | no `body` key |
| other keys on `body` (e.g. `disabled`) | copied onto the v2.1 `body` |

## Auth

| v3 | v2.1 |
|---|---|
| `{type: T, credentials: [{key: k, value: v}…]}` | `{type: T, T: [{key: k, value: v, type: "string"}…]}`. Boolean/number values keep their JSON type and get `type: "boolean"|"number"`. |
| `{type: noauth}` | `{type: "noauth"}` |
| `id`, `name` of a list entry | dropped; the multi-auth list itself is kept under `postmanV3.auth` when it has 2+ entries |

The type mapping (bearer, basic, apikey, digest, oauth2, awsv4 → Xenon; others → unsupported and kept) is the
existing v2.1 one.

## Scripts

| v3 `scripts[].type` | v2.1 `event[]` |
|---|---|
| `beforeRequest`, `http:beforeRequest` | `{listen: "prerequest", script: {type: "text/javascript", exec: code split on "\n"}}` |
| `afterResponse`, `http:afterResponse` | `{listen: "test", script: {…}}` |
| `beforeQuery` (on a `graphql-request`) | `prerequest` |
| `graphql:*`, `grpc:*` (collection/folder) | not emitted. Kept under `postmanV3.scripts`, with one `protocol_script` warning per script |

## Example

| v3 `http-example` | v2.1 `response[]` |
|---|---|
| `id`, `name` (R11) | `id`, `name` |
| `request {url, method, headers, queryParams, body}` | `originalRequest {method, header, url {raw, query}, body}` |
| `response.statusCode` | `code` |
| `response.statusText` | `status` |
| `response.headers` | `header` |
| `response.body.content` | `body` (string) |
| `response.body.type` | `_postman_previewlanguage` (`json`, `text`, `xml`, `html`, `javascript`) |

## Environment / globals

| v3 | v2.1 environment JSON |
|---|---|
| `name` (R11) | `name` |
| `values[] {key, value, enabled}` | `values[] {key, value, enabled}` |
| `type: secret` or `secret: true` | `type: "secret"` |
| secret with `value == "********"` | `value: ""` |
| `source {…}` (vault link) | `value: ""`, `type: "secret"`, `source` kept. One `linked_secret` warning. |
| other `type` (`default`, `any`) | same `type` |
| — | `_postman_variable_scope: "environment" | "globals"` (R5) |
| `id`, `owner`, `color`, `createdAt`, `updatedAt`, `lastRevision`, `lastUpdatedBy`, `team`, `workspace`, `path` | copied (passthrough) |
