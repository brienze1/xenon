# Data Model: Import and Saved Examples (Stage 3)

Extends XEN-001/XEN-002 data models. File shapes: [contracts/format-additions.md](contracts/format-additions.md).

## Persistent (workspace files)

### SavedExample (`internal/domain/entity/saved_example.go`)
| Field | Type | Rules |
|---|---|---|
| ID | ULID | immutable |
| Name | string | 1–200 chars, duplicates allowed |
| Request | ExampleRequest | method, URL, params, path variables, headers, body, auth (no settings/scripts) |
| Response | ExampleResponse | Code (100–599), Status text, ContentType, Headers (ordered, repeats kept), Body bytes, BodyEncoding (`text\|base64`) |
| Passthrough | YAML node | `postman:` block |

Owned by a Request (`Request.Examples []SavedExample`, ordered). Tree nodes gain kind `example`
(child of a request; leaf).

### Passthrough
A YAML node attached to Workspace/Collection/Folder/Request/Environment/VariableRow/Example
(`postman:` and `bruno:` keys), preserved verbatim by XEN-001's unknown-field rule and editable
only by import/export.

### Auth addition
`AuthType` gains `unsupported` with `Unsupported { Source, Type, Data (YAML node) }`.

## In memory (import pipeline)

### ImportSource
`{ id (ULID), kind: postman_collection | postman_environment | postman_globals | postman_dump |
bruno_bru | bruno_yaml | curl, version, root (temp path or text), files[] }` — created by Detect,
cached for 30 min or until Import/Discard.

### ImportItem
One importable unit from a source:
`{ key (stable within source), kind: collection | environment | globals | request,
name, counts { folders, requests, examples, variables, scripts }, payload (Xenon entities:
Collection tree with requests+examples / Environment + secret values / VariableRows / Request),
warnings[] }`.

### ImportWarning
`{ code, severity: info | warning, message, item_key, path (e.g., "Billing/Invoices/Create"),
location? (file:line for scripts) }`. Codes: `unsupported_auth`, `pm_require`, `missing_file`,
`folder_variables`, `request_variables`, `untranslated_script`, `skipped_part`,
`unsupported_request_type`, `disabled_body`, `raw_header_string`, `pm_api_outside_mvp`.

### PmUsage
`{ api (e.g., "pm.cookies.jar"), occurrences: [{ item_key, path, script: pre|post, line }] }` —
only APIs outside the MVP subset.

### ImportPlan
`{ source_id, items: [{ key, include, target_local_id, target_parent_id? (collection/folder for
collections or requests; empty = new collection), resolution?: replace | keep_both | skip }],
apply_to_all?: resolution }`.

### Clash
`{ item_key, kind: collection | environment | request | variable, existing_id, existing_name }`.

### ImportResult
`{ per_workspace: [{ local_id, created: counts, replaced, skipped }], warnings[], pm_usage[],
failed: [{ item_key, error }] }`.

## State machine (one import)
```
Detected(source, items, warnings) --Plan--> Planned
Planned --clashes unresolved--> NeedsResolution --resolutions--> Planned
Planned --Import--> Writing (progress events per item) --> Done(result)
Detected/Planned --Discard or 30 min--> (temp files removed)
Writing --item failure--> that item rolled back (staging removed), others continue
Cancel allowed until Writing starts; during Writing, cancel stops after the current item.
```
