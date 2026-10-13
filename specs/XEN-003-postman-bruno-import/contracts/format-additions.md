# Contract: workspace format additions (stage 3, still `xenon: 1`)

Additions to [XEN-001 workspace-format.md](../../XEN-001-shell-and-storage/contracts/workspace-format.md)
and [XEN-002 format-additions.md](../../XEN-002-environments-and-secrets/contracts/format-additions.md).

## 1. Auth type `unsupported`

```yaml
auth:
  type: unsupported
  unsupported:
    source: postman            # postman | bruno
    type: hawk                 # original type name
    data:                      # original attributes, verbatim
      - { key: authId, value: "...", type: string }
```
Sends as no auth, never inherits, read-only in the UI ("unsupported auth type (kept)").

## 2. `postman:` passthrough blocks

Present on `collection.yaml`, `folder.yaml`, request files, environment files, examples and
variable rows (row-level key `postman:`) when the item came from Postman. Holds **only the
source keys Xenon does not map**, as YAML equivalents of the JSON. Mapped keys per kind are
listed in [postman-mapping.md](postman-mapping.md); everything else from the source goes here.

```yaml
postman:
  id: "8c1f…"                         # source id (item.id / _postman_id)
  protocolProfileBehavior: { disableBodyPruning: true, strictSSL: false }
  variable: [ { key: page, value: "1", type: string } ]   # folder/request-level variables
  proxy: { … }
  certificate: { … }
  event: { prerequest: { id: "…", name: "…", packages: {} } }  # script metadata only (text is in `scripts`)
  header: { raw: null }                # e.g. header given as a raw string originally
```

A `bruno:` block plays the same role for Bruno-only fields (e.g., `meta.tags`, `settings` keys
Xenon lacks).

## 3. Saved examples (request files)

```yaml
examples:
  - id: 01JA4D…                       # ULID
    name: Created
    request:                          # same shapes as the request file (§5 of the base format)
      method: POST
      url: "{{base_url}}/invoices"
      params: []
      pathVariables: []
      headers: []
      body: { mode: raw, raw: { language: json, content: '{"amount":1000}' } }
      auth: { type: inherit }
    response:
      code: 201
      status: Created
      contentType: application/json
      headers:
        - { key: Content-Type, value: application/json }
      bodyEncoding: text              # text | base64
      body: |
        { "id": 42 }
    postman: {}                       # cookie, responseTime, timings, _postman_previewlanguage, ids
```
Order of `examples` = sidebar order. Example names need not be unique.

## 4. Script disabled flags

```yaml
scripts:
  preRequest: "…"
  postResponse: "…"
  disabled: [preRequest]        # scripts kept but not run (default [])
```
Applies to collection, folder and request files.

## 5. File reference state

Form-data rows and binary bodies gain `srcState: ok | repick` (default `ok`). Imports write
`repick`; the send pipeline treats `repick` like a missing file; picking a file sets `ok`.

## 6. `.gitignore` block addition
`.xenon-import-*/` (staging directories of in-progress imports).
