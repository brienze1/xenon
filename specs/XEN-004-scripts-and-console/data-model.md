# Data Model: Scripts, Tests and Console (Stage 4)

No new workspace file shapes: scripts, `scripts.disabled` and `inheritScripts` already exist
(XEN-001 format, XEN-003 additions). App settings gain one field.

## Persistent
- `app.json` → `settings.scriptTimeoutMs` (int, default 10000, range 1000–60000).

## In memory

### ScriptRef
`{ owner_kind: collection | folder | request, owner_id, owner_name, phase: pre_request |
post_response }` — identifies a script in errors, logs and test results.

### SendExecution (one per send)
| Field | Notes |
|---|---|
| ID | ULID; groups console entries |
| LocalVariables | map for `pm.variables` (lives for this send) |
| PreRuns / PostRuns | []ScriptRun in execution order |
| SentRequest | request after pre-request mutations and variable resolution (masked copy for console) |
| Outcome | `sent` \| `skipped_by_script` \| `stopped_by_script_error` \| `cancelled` \| `error` |
| Response | XEN-001 response (absent unless sent) |
| Tests | []TestResult |
| Nested | []NestedRequest (sendRequest/runRequest: request, response/error, timings) |

### ScriptRun
`{ script: ScriptRef, started_at, duration_ms, status: ok | error | skipped, error?: ScriptError,
test_count }`.

### ScriptError
`{ kind (sandbox-api §6), message, line?, column?, script: ScriptRef }`.

### TestResult
`{ index, name, status: passed | failed | skipped, error_message?, script: ScriptRef, async }`.

### ConsoleEntry
| Field | Notes |
|---|---|
| ID | ULID (time-ordered) |
| SendID | groups with its SendExecution |
| Kind | `request` \| `log` \| `error` \| `test_failure` \| `info` |
| Level | for logs: log, info, warn, error, debug |
| Script | ScriptRef (logs/errors) |
| Summary | one line (method URL status time size, or first 200 chars of the log) |
| Detail | request: resolved request + response headers/body (body inline ≤ 1 MB, else `body_url`); log: serialised arguments (JSON-like tree, cycles marked) |
| Masked | true when any secret value was replaced |

Ring buffer of 5,000 entries per app session (oldest dropped); cleared by Clear / `console.clear()`.

## State machine (send with scripts)
```
Start → PreScripts (in order; each: ok | error→Stop | skip→Skipped | timeout→Stop)
  → Resolve variables (local scope included) → Send (cancellable)
  → PostScripts (in order; errors recorded, chain continues)
  → Collect tests (pm.test + legacy tests) → Done
Stop / Skipped: no network call; history entry with outcome; Test Results shows pre-request tests.
```
History (XEN-001) records every attempted send with outcome; it never stores the script-mutated
request (spec Edge Cases).
