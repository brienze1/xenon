# Research: Import and Saved Examples (Stage 3)

Sources verified 2026-10-09: Postman collection v2.1.0 JSON schema, Postman docs (export data,
collection schemas), `postmanlabs/curl-to-postman` (supportedOptions/unnecessaryOptions,
cmd transform), `usebruno/bruno` (`bruToJson.js`, `envToJson.js`, `collectionBruToJson.js`,
`postman-translations.js`, `bruno-to-postman-translator.js`), Bruno OpenCollection docs,
module proxy for Go libraries.

## R1. Postman input formats
- **Decision**: accept collection v2.1 and v2.0 JSON (schema URL in `info.schema`; v2.0 auth
  key/value maps normalised to v2.1 attribute lists), environment JSON
  (`_postman_variable_scope: environment`), globals JSON (`…: globals`), data dump (zip or folder
  with `archive.json` + per-type subfolders of JSON files). v1 and v3 YAML are detected and
  refused with a reason.
- **Data dump**: contains collections and environments from all of the user's workspaces with
  no workspace membership → per-item target assignment (spec US3).
- **URL**: `url.raw` is authoritative; rebuilt from `protocol/host/port/path/query/hash` only
  when `raw` is absent.

## R2. Passthrough model for lossless export
- **Decision**: every imported file gets a `postman:` block holding the source item's keys that
  Xenon does **not** map (the contract lists mapped keys per kind). Export (stage 6) generates
  Postman JSON from Xenon fields and overlays the passthrough's unmapped keys; arrays Xenon owns
  (headers, query, etc.) are always regenerated, so edits/deletions in Xenon win.
- **Rationale**: storing the full original JSON would duplicate bodies/scripts and resurrect
  deleted rows on export; storing only unmapped keys keeps files small and edits authoritative.
- **Alternatives**: full original snapshot (rejected above); dropping unknowns (violates
  constitution VI round-trip).

## R3. Unsupported auth types
- **Decision**: new auth type value `unsupported` with
  `unsupported: { source: postman|bruno, type: <original>, data: <original attributes> }`.
  It behaves as "no auth" when sending (never inherits), is shown read-only, and exports back to
  the original type.

## R4. Saved examples storage
- **Decision**: in the request file under `examples:` (format additions §3). Response body as a
  literal block for text; `bodyEncoding: base64` for binary. Keeps the PRD's one-file-per-request
  rule; conflicts stay per request.
- **Alternatives**: a sibling `<request>.examples/` directory (smaller diffs for huge bodies,
  but a second artifact per request and a new directory rule in the tree).

## R5. Bruno
- **Decision**: two parsers behind one adapter:
  - `.bru`: a hand-written line-based parser following `bruno-lang` v2 grammar (blocks
    `meta`, method blocks incl. `http`, `params:query|path`, `headers`, `auth:*`, `body:*`,
    `vars:pre-request|post-response`, `assert`, `script:pre-request|post-response`, `tests`,
    `docs`, `settings`; dictionary `key: value`, `~` disabled, `'''` multiline,
    `@file(...)`/`@contentType(...)` in multipart). Collection root `bruno.json`; `collection.bru`,
    `folder.bru`, `environments/*.bru` (`vars`, `vars:secret [ … ]`, `color`).
  - YAML (OpenCollection): root `opencollection.yml` (`opencollection: "1.0.0"`), `folder.yml`,
    request `*.yml` (`info`, `http`, `runtime.scripts`, `runtime.assertions`, `settings`, `docs`),
    `environments/*.yml`; parsed with `go.yaml.in/yaml/v3`. Field names follow Bruno's structure
    reference; fixtures are generated with Bruno 3.1 itself so the parser is tested against real
    output.
- Order: `meta.seq` / `info.seq`, ties by file name. A collection mixing both formats is refused.

## R6. Bruno → pm translation
- **Decision**: a regex/token table ported from Bruno's `bruno-to-postman-translator.js`
  ([contracts/bruno-mapping.md](contracts/bruno-mapping.md) §3), applied per line; lines with any
  remaining `bru.`/`req.`/`res.` token are prefixed `// XENON: untranslated` and reported.
  Assertions are converted with an operator table (§4). Request vars become
  `pm.variables.set(...)` lines (pre-request vars in the pre-request script head; post-response
  vars in the post-response script, evaluated as expressions over `res`, translated the same way).

## R7. cURL parsing
- **Decision**: normalise continuations first (bash `\⏎`, cmd `^⏎` + `^` escapes + `""`→`"` per
  curl-to-postman's `transformCmdToBash`, PowerShell `` `⏎ ``), then tokenize with
  `mvdan.cc/sh/v3/syntax` (handles `$'…'`, nested quotes) and apply Xenon's flag table
  (spec FR-024). Dependency: `mvdan.cc/sh/v3` (BSD-3, actively maintained).
- **Alternatives**: `go-shellwords` (less robust with `$'…'`), `k1LoW/curlreq` (too few flags).

## R8. pm usage scanner
- **Decision**: pure use case scanning script text for `pm.` member chains (tokenised, ignoring
  comments and strings) and comparing against the MVP subset list (PRD "What scripts can do");
  results grouped by API with locations; included in every import summary.

## R9. Detection order
1. Text starting with `curl ` (after trimming) → cURL.
2. Zip → unpack to a temp dir (size cap 1 GB, zip-slip protection) → treat as folder.
3. Folder: `archive.json` → Postman dump; `bruno.json` → Bruno .bru; `opencollection.yml` →
   Bruno YAML; otherwise refuse.
4. JSON file/text: `info.schema` contains `collection/v2.1` or `v2.0` → collection; `v1` or
   `item` absent → refuse; `_postman_variable_scope` → environment/globals.
5. `.yaml`/`.yml` file with `*.request.yaml` neighbours → Postman v3 → refuse (not in MVP).

## R10. Atomic per-item import
- **Decision**: each top-level item is written into a staging directory inside the workspace
  (`.xenon-import-<id>/`, ignored by the tree and watcher) and moved into place with one rename
  (directory) or atomic file writes (environment + vault + local values, using XEN-002's
  multi-store order). On failure the staging directory is removed. `.gitignore` entry
  `.xenon-import-*/` is added by Xenon's block.
