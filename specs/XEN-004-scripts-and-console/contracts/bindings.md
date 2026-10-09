# Contract: Go ↔ UI bindings — stage 4 additions

Conventions as before. New prefixes: `SCR` scripts · `CON` console.

## 1. `RequestService.Send` result additions
`SendResult` gains:
- `outcome`: `sent | skipped_by_script | stopped_by_script_error | cancelled | error`
- `tests: [{ index, name, status, error_message?, script: { owner_kind, owner_name, phase } }]`
- `script_errors: [{ kind, message, line?, column?, script }]`
- `send_id` (to correlate console entries)
A pre-request failure returns a normal `SendResult` with `outcome: stopped_by_script_error`
(not an RPC error), so the UI shows it in the response area.

## 2. `ScriptService`
| Method | Input | Output | Notes |
|---|---|---|---|
| `Check(req)` | `{ text }` | `{ errors: [{ message, line, column }] }` | syntax only (no execution) |
| `Reference()` | — | `{ members: [{ path, signature, description, kind }] }` | pm API + globals + assertion helpers for autocomplete |
| `Snippets()` | — | `[{ id, title, phase, text }]` | Postman's standard snippets |

## 3. `ConsoleService`
| Method | Input | Output | Notes |
|---|---|---|---|
| `List(req)` | `{ after_id?, limit }` | `ConsoleEntry[]` | initial load / pagination |
| `Get(req)` | `{ id }` | `ConsoleEntry` with full detail | `CON-01404` dropped from buffer |
| `Clear()` | — | — | |

Event `console:entries` `{ entries: ConsoleEntry[] (summaries), cleared?: true }`, batched ≤ 50 ms.

## 4. `SettingsService` (new, minimal)
| Method | Input | Output |
|---|---|---|
| `Get()` | — | `{ script_timeout_ms }` |
| `Save(req)` | `{ script_timeout_ms }` | same (`SCR-01400` out of range 1000–60000) |
