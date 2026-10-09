# Contract: Go ↔ UI bindings — stage 2 additions

Same conventions as [XEN-001 bindings](../../XEN-001-shell-and-storage/contracts/bindings.md)
(ctx first, snake_case DTOs, error JSON with codes). New prefixes: `ENV` environments ·
`VAR` variables · `VLT` vault.

**Secret rule for every DTO in this file**: secret values are never included in responses,
except `VaultService.Reveal`. Secret rows carry `value: ""`, `secret: true` and
`secret_state: set | missing | locked`.

`VariableRowDTO { key, initial_value, current_value, has_current_override, secret, secret_state?,
enabled, description }` — for secret rows `initial_value` and `current_value` are `""`.

## 1. `EnvironmentService`
| Method | Input | Output | Notes / errors |
|---|---|---|---|
| `List(req)` | `{ local_id }` | `{ environments: [{ id, name, color, confirm_before_send, error? }], active_id }` | malformed files listed with `error` |
| `Get(req)` | `{ local_id, id }` | `EnvironmentDTO { id, name, description, color, confirm_before_send, variables: VariableRowDTO[] }` | `ENV-01404` |
| `Create(req)` | `{ local_id, name }` | `EnvironmentDTO` | name unique `ENV-02409`; prod defaults |
| `Update(req)` | `{ local_id, id, name, description, color, confirm_before_send }` | `EnvironmentDTO` | rename renames file |
| `Duplicate(req)` | `{ local_id, id }` | `EnvironmentDTO` | copies current values + secrets |
| `Delete(req)` | `{ local_id, id }` | `{ active_id }` | removes file, current values, vault entries |
| `SetActive(req)` | `{ local_id, id \| "" }` | `{ active_id }` | `""` = No environment |

## 2. `VariableService` (all scopes: `scope` = `env:<id>` \| `collection:<id>` \| `globals`)
| Method | Input | Output | Notes |
|---|---|---|---|
| `GetTable(req)` | `{ local_id, scope }` | `VariableRowDTO[]` | |
| `SaveTable(req)` | `{ local_id, scope, rows: [{ key, previous_key?, initial_value, current_value?, set_current: bool, secret_value?, secret, enabled, description }] }` | `VariableRowDTO[]` | writes the file (initial values), local current values and vault in one operation; `previous_key` carries renames; plain→secret moves value to vault (response `warnings: ["git_history"]`); secret→plain needs `confirm_reveal: true` else `VAR-01409`; invalid key `VAR-00400` |
| `SetCurrent(req)` | `{ local_id, scope, key, value }` | `VariableRowDTO` | secret rows: value goes to vault |
| `Reset(req)` | `{ local_id, scope, keys[] }` | `VariableRowDTO[]` | empty `keys` = all |
| `Persist(req)` | `{ local_id, scope, keys[] }` | `VariableRowDTO[]` | plain rows only (`VAR-02422` for secret) |
| `Inspect(req)` | `{ local_id, request }` | `{ references: [{ field, start, end, name, status, scope, display_value, secret }], variables: [{ name, status, scope, display_value, secret, used_in: [fields] }] }` | for highlighting + Variables panel; includes inherited auth/headers; `display_value` masked for secrets, `""` for dynamic |
| `Suggest(req)` | `{ local_id, request_id?, prefix }` | `[{ name, scope, secret, dynamic, description }]` | for `{{` autocomplete, max 50 |
| `AddVariable(req)` | `{ local_id, scope, key, value, secret }` | `VariableRowDTO` | "Add to environment / collection / globals" |

## 3. `VaultService`
| Method | Input | Output | Notes |
|---|---|---|---|
| `Status()` | — | `{ mode: keychain \| passphrase \| none, needs_passphrase_setup, locked_local_ids[], broken_local_ids[] }` | |
| `SetupPassphrase(req)` | `{ passphrase, confirmation }` | `Status` | mismatch `VLT-00400`; min 8 chars `VLT-01400` |
| `Unlock(req)` | `{ passphrase }` | `Status` | wrong `VLT-02401`; unlocks every vault |
| `Reveal(req)` | `{ local_id, scope, key }` | `{ value }` | only call that returns a secret value; `VLT-03404` missing |
| `CopySecret(req)` | `{ local_id, scope, key }` | — | Go writes the value to the clipboard itself (value never crosses the bridge) |
| `Reset(req)` | `{ local_id, confirm: true }` | `Status` | discards stored values for that workspace |

## 4. Changes to XEN-001 services
- `RequestService.Send` input gains `{ acknowledge_missing_secrets: bool, confirmed_environment_id:
  string }`. Preflight errors (no request is sent, no history recorded):
  - `SND-05409` confirmation required → `error_details` carries `{ environment_id, name, color }`.
  - `SND-06409` missing/locked secrets → `error_details` lists the names.
  The UI prompts, then calls `Send` again with the acknowledgement.
- `RequestService.Send` resolves variables on a copy; `SendResult` and History keep unresolved
  text.
- `TreeService.GetSettings/SaveSettings` for collections gain `variables: VariableRowDTO[]` (via
  `VariableService` on `collection:<id>`; the settings dialog uses `VariableService` directly).
- `WorkspaceService.Remove` also deletes vault, key file, keychain entry and `variables.json`.

## 5. Events
| Name | Payload | When |
|---|---|---|
| `variables:changed` | `{ local_id, scopes[] }` | any table/current value/active env change (from UI or outside edits) → UI re-inspects |
| `vault:status-changed` | `Status` | locked/unlocked/broken changes |
