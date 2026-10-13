# Known differences from postman-sandbox 6.7.4

`make parity` runs the 249-case corpus (`test/integration/resources/scripts/corpus/`) through
Xenon's sandbox and compares each outcome (variables, sent request, tests, console, errors,
sendRequest/runRequest calls, vault, skip) with what postman-sandbox 6.7.4 produced for the same
case. **246 of 249 cases match (98.8%)**; SC-001 requires at least 98%.

The parity test fails on any difference that is not listed below, and on a listed case that
starts matching again (so this list stays exact). Each row gives the cause.

| Case | What differs | Cause |
|---|---|---|
| `error-type-error-message` | Reading a property of `null` reports `Cannot read properties of undefined (reading 'property')`; Postman says `of null`. | Xenon's engine (goja) words this TypeError as "Cannot read property 'x' of undefined" for both `null` and `undefined` and does not keep which one it was. Xenon rewrites goja's wording to V8's, but cannot recover `null`. The error, its line and column are otherwise identical. |
| `require-stream` | `require('stream')` streams emit their `data` and `end` events in Xenon; in postman-sandbox they never fire. | postman-sandbox's bundled Node `process.nextTick` shim calls a `setTimeout` its bundle scope cannot see, so stream events are lost there (a Postman bug). Xenon's Node shims schedule on the sandbox's own event loop. |
| `require-timers` | `require('timers').setTimeout(...)` works in Xenon; postman-sandbox throws `ReferenceError: setTimeout is not defined`. | Same postman-sandbox bundling bug: its `timers` module cannot reach the sandbox timers. Scripts that work in Postman are unaffected; Xenon only succeeds where Postman fails. |

## Harness inputs that mirror the oracle

These are not differences: the parity harness gives the sandbox the same inputs the oracle gives
postman-sandbox, which a real Xenon send fills in differently.

- `pm.info.iteration`, `iterationCount`, `requestName`, `requestId` and `pm.execution.location`
  are undefined/empty in the oracle (it passes no cursor or item). A Xenon send always has
  iteration 0 of 1, the request's name and id, and its collection/folder path.
- `pm.cookies` comes from the case's `cookies` fixture and `pm.response.cookies` is empty, as in the
  oracle. A Xenon send fills both from the response's `Set-Cookie` headers until the stage 5 cookie
  jar exists.
- Script errors from unhandled promise rejections carry the prefix `Unhandled promise rejection: `
  in Xenon's own reports; the harness drops it because postman-sandbox reports the rejected error
  itself.
- `parity-diff` runs with `TZ=UTC`, as the oracle does.
