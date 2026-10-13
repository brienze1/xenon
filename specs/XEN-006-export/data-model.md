# Data Model: Export (Stage 6)

## ExportRequest (`internal/domain/entity/export_request.go`)
| Field | Type | Rules |
|---|---|---|
| LocalID | string | workspace |
| ItemKind | enum `workspace\|collection\|folder\|request\|environment\|globals` | |
| ItemID | ULID | empty for workspace/globals |
| Format | enum `xenon_yaml\|postman` | |
| Shape | enum `zip\|bundle\|file` | valid combinations: workspace→zip; collection/folder→zip\|bundle (xenon) or file (postman); request/environment/globals→file |
| IncludeSecrets | bool | only with `xenon_yaml`; requires Passphrase (≥ 8 chars, confirmed) |
| CurrentValuesAsInitial | bool | default false |
| TargetPath | string | from save dialog |

## ExportSummary
`{ counts: { collections, folders, requests, examples, environments, variables }, dropped:
[{ item_path, field, reason }], warnings: [{ item_path, code, message }] (e.g.,
literal_credential), secrets: included|omitted|names_only, path }`.

## Import additions (XEN-003 pipeline)
`ImportSource.kind` gains `xenon_workspace_zip | xenon_workspace_folder | xenon_bundle |
xenon_zip | xenon_request | xenon_environment`; `ImportSource.encrypted_secrets: bool`;
`ImportPlan.passphrase?`.

## State machine (export)
```
Configure → (Postman? hide Include secrets) → Choose path → Writing (progress, cancellable;
temp file) → Done(summary) | Cancelled (temp removed) | Failed (temp removed, error)
```
