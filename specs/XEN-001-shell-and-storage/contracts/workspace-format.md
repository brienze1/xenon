# Contract: Workspace on-disk format (format version 1)

This is **the** storage contract for every Xenon workspace. Stage 1 defines it in full; later
stages fill in reserved fields but never change the meaning of existing ones. Anything written
here is what lands in a company's git repository, so it must be readable, diff-friendly and
free of secret values (constitution I, II).

## 1. Folder layout

```
<workspace-folder>/                 ← anywhere on disk, chosen by the user
├── workspace.yaml
├── .gitignore                      ← written by Xenon (see §8)
├── environments/                   ← stage 2 (created empty in stage 1)
└── collections/
    ├── billing-api/                ← one directory per collection
    │   ├── collection.yaml
    │   ├── create-customer.yaml    ← one file per request
    │   └── invoices/               ← one directory per folder (nest freely)
    │       ├── folder.yaml
    │       ├── create-invoice.yaml
    │       └── drafts/
    │           ├── folder.yaml
    │           └── create-draft.yaml
    └── auth-service/
        └── collection.yaml
```

Rules:

- A directory under `collections/` is a collection iff it contains `collection.yaml`; a directory
  under a collection/folder is a folder iff it contains `folder.yaml`. Directories without the
  marker file are ignored (and listed as a warning in the sidebar's error area).
- Every other `*.yaml` file inside a collection/folder directory is a request.
- Files and directories not matching these rules are ignored and never touched.

## 2. Common header (every file)

```yaml
xenon: 1                # format version (integer). Readers reject versions > the latest they know
kind: request           # workspace | collection | folder | request
id: 01JA2Z6Q8K4S7M3N5P9R0T1V2W   # ULID, generated once, never changes (survives renames/moves)
name: Create invoice    # display name, any Unicode, not necessarily unique
description: |          # markdown, optional, default ""
  Creates a draft invoice.
```

- `id` is what tabs, history and UI state refer to; the file path is derived and may change.
- If two files carry the same `id` (e.g., a file copied by hand), the second one loaded gets a new
  `id` written back on next save and a warning is shown.

## 3. `workspace.yaml`

```yaml
xenon: 1
kind: workspace
id: 01JA2Z5V3B6C7D8E9F0G1H2J3K
name: Acme
description: ""
order:                  # collection directory names, in sidebar order
  - billing-api
  - auth-service
variables: []           # stage 2: workspace globals (same row shape as §6.1, secret rows have no value)
sync: {}                # stage 7: repo-level settings that are safe to share (no credentials)
```

## 4. `collection.yaml` and `folder.yaml`

```yaml
xenon: 1
kind: collection        # or: folder
id: 01JA2Z6A1B2C3D4E5F6G7H8J9K
name: Billing API
description: ""
order:                  # child entries in sidebar order: folder directory names and request file names
  - create-customer.yaml
  - invoices
auth:                   # §6.2 — collection "inherit" means "no auth"
  type: bearer
  bearer:
    token: "{{billing_token}}"
headers:                # §6.1 rows, applied to every request below unless overridden
  - key: X-Tenant
    value: acme
    enabled: true
    description: ""
variables: []           # stage 2: collection variables
scripts:                # stage 4
  preRequest: ""
  postResponse: ""
inheritScripts: true    # folder only (stage 4); ignored on collections
postman: {}             # stage 3: Postman-only fields kept verbatim for round trip
```

## 5. Request file (`<name>.yaml`)

```yaml
xenon: 1
kind: request
id: 01JA2Z7C2D3E4F5G6H7J8K9M0N
name: Create invoice
description: ""
method: POST            # GET POST PUT PATCH DELETE HEAD OPTIONS or any custom token (stored upper-case)
url: "https://{{host}}/v1/customers/:customerId/invoices?draft=true"
params:                 # ALL query rows incl. disabled; enabled rows == the URL's query string, in order
  - key: draft
    value: "true"
    enabled: true
    description: ""
  - key: debug
    value: "1"
    enabled: false
    description: only in dev
pathVariables:          # one row per :name segment in the URL, in order
  - key: customerId
    value: "42"
    description: ""
headers:                # §6.1
  - key: Accept
    value: application/json
    enabled: true
    description: ""
autoHeaders:            # auto-generated headers the user switched off (FR-027)
  disabled: [User-Agent]
auth:
  type: inherit
body:                   # §6.3
  mode: raw
  raw:
    language: json
    content: |
      { "amount": 1000 }
settings:
  timeoutMs: 0          # 0 = no timeout
  followRedirects: true
  verifySsl: true
  encodeUrl: true
scripts:                # stage 4
  preRequest: ""
  postResponse: ""
inheritScripts: true    # stage 4
postman: {}             # stage 3
```

## 6. Shared shapes

### 6.1 Key/value row

```yaml
key: string
value: string           # may contain {{variables}}; stage 2 secret rows: no value, `secret: true`
enabled: bool           # default true
description: string     # default ""
```

### 6.2 `auth`

`type` is one of: `inherit` · `none` · `bearer` · `basic` · `apikey` · `oauth2` · `awsv4` ·
`digest`. Stage 1 implements the first four; the others are **reserved** — readers accept and
preserve them, the UI shows "available in a later stage".

```yaml
auth: { type: inherit }
auth: { type: none }
auth: { type: bearer, bearer: { token: "…" } }
auth: { type: basic,  basic:  { username: "…", password: "…" } }
# reserved (stage 5), sub-key = type name, fields defined then:
auth: { type: apikey, apikey: { key: "…", value: "…", in: header } }
auth: { type: oauth2, oauth2: { grantType: authorization_code, … } }
auth: { type: awsv4,  awsv4:  { … } }
auth: { type: digest, digest: { … } }
```

Only the sub-key matching `type` is meaningful; other sub-keys are preserved (so switching type
back and forth in the UI does not lose what was typed — Postman behaviour).

### 6.3 `body`

`mode`: `none` · `formdata` · `urlencoded` · `raw` · `binary` · `graphql`. Like `auth`, every
mode's sub-key is preserved when switching modes.

```yaml
body:
  mode: formdata
  formdata:
    - { key: note, type: text, value: "hello", enabled: true, description: "" }
    - { key: file, type: file, src: "fixtures/invoice.pdf", enabled: true, description: "" }
  urlencoded:
    - { key: a, value: "1", enabled: true, description: "" }
  raw:
    language: json      # json | xml | text | html | javascript
    content: ""
  binary:
    src: "/Users/luis/Desktop/payload.bin"
  graphql:
    query: |
      query { me { id } }
    variables: |
      {}
```

File paths (`src`): relative to the **workspace folder** with `/` separators when the file is
inside it; absolute (OS-native) otherwise. Resolution happens at send time; a missing file blocks
the send (FR-028a).

## 7. Writing rules

- **Encoding**: UTF-8, LF line endings, 2-space indent, trailing newline. Multi-line strings use
  literal block style (`|`), so bodies and scripts diff line by line.
- **Field order** on write is the order shown in this document; unknown fields are kept in their
  original position (or appended at the end of their mapping if new).
- **Unknown fields are preserved**: the file is loaded as a YAML node tree, Xenon updates only the
  fields it owns, and writes the tree back. Comments in hand-edited files are preserved where the
  YAML library allows.
- **Defaults are written explicitly** for the fields in this document (no "omit if default"), so
  every file is self-describing; empty reserved blocks (`variables: []`, `scripts`, `postman: {}`)
  are written too.
- **Atomic writes**: write to a temp file in the same directory, fsync, rename over the target.
- **Ordering**: a parent's `order` lists its children. On load, listed entries that do not exist
  are skipped; existing entries that are not listed are appended in name order (so hand-added or
  git-merged files appear instead of vanishing). The next save of the parent rewrites `order` to
  match what is shown.

## 8. File names

Derived from `name` on create and rename, never shown to the user:

1. Unicode NFC; lower-case.
2. Replace any of `< > : " / \ | ? *`, control chars, and whitespace runs with `-`; collapse
   repeated `-`; trim `-`, `.` and spaces from both ends.
3. Windows reserved names (`con`, `prn`, `aux`, `nul`, `com1`–`com9`, `lpt1`–`lpt9`) get a `_`
   suffix. Empty result → `untitled`.
4. Truncate to 64 characters (on a rune boundary).
5. Make unique among siblings **case-insensitively** by appending `-2`, `-3`, …
6. Requests get `.yaml`; collections and folders are directories.

Renaming an item renames its file/directory (git sees a rename) and updates the parent's
`order`. `collection.yaml` and `folder.yaml` are reserved and never produced by rule 1–5
(a request named "collection" becomes `collection-2.yaml`).

## 9. `.gitignore`

Created in the workspace folder if missing, and Xenon only ever appends its own block:

```
# Xenon — local-only files (never commit)
.xenon/
*.xenon-tmp
```

Stage 1 writes nothing into `.xenon/`; the entry reserves the name for any future per-folder
local data. All local state lives in app data (see data-model.md).

## 10. Versioning

`xenon: 1` for everything in this document. A later stage that adds a field in a reserved slot
does **not** bump the version. A version bump is only for a breaking change and comes with an
automatic, in-place migration on load (confirmed with the user before rewriting files).
