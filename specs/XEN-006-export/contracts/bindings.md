# Contract: Go ↔ UI bindings — stage 6 additions

Prefix `EXP`.

## `ExportService`
| Method | Input | Output | Notes |
|---|---|---|---|
| `Options(req)` | `{ local_id, item_kind, item_id? }` | `{ formats: [{ format, shapes[], include_secrets_allowed }], has_secrets, has_current_values, suggested_file_name }` | drives the dialog |
| `PickTarget(req)` | `{ suggested_file_name, format, shape }` | `{ path }` | native save dialog (overwrite confirmation is native) |
| `Export(req)` | `ExportRequest` | `ExportSummary` | cancellable via ctx; `EXP-00400` invalid combination; `EXP-01400` passphrase too short/mismatch; `EXP-02500` write failure (temp removed) |
| `ShowInFolder(req)` | `{ path }` | — | OS file manager |

Event `export:progress { done, total }`.

## `ImportService` additions (XEN-003)
- `Detect` recognises Xenon kinds; preview includes `encrypted_secrets`.
- `Import` plan accepts `passphrase`; wrong passphrase → `IMP-08401`, missing passphrase →
  `IMP-50401` (UI offers retry or `skip_secrets: true`).
- Workspace zip/folder import requires `items[].target_folder` (empty or non-existent; then
  `target_local_id` is not needed) → registered as a new workspace.
