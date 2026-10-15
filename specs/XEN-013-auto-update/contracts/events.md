# Contract: Go → UI events

## `update:status`

Emitted by `internal/integration/publisher/update_status.go` on every state change and on download
progress (at most every 250 ms, plus 0 % and 100 %). Payload = the full status:

```json
{
  "state": "downloading",
  "reason": "",
  "current_version": "0.1.4",
  "latest_version": "0.1.7",
  "release_url": "https://github.com/brienze1/xenon/releases/tag/v0.1.7",
  "received": 1048576,
  "total": 15728640,
  "folder": "",
  "automatic": true
}
```

| Field | Values |
|---|---|
| `state` | `idle` `checking` `up_to_date` `available` `no_package` `check_failed` `downloading` `ready` `installing` `failed` |
| `reason` | `""` or `unreachable` `download_failed` `unverified` `not_packaged` `not_writable` `install_failed` |
| `folder` | set only with `not_writable` |
| `automatic` | true when the startup check started the download |

The same object is the return value of every `UpdateService` method.

Frontend: `frontend/src/platform/events.ts` adds `UPDATE_STATUS = 'update:status'` and
`onUpdateStatus(cb)`, mapping through `platform/mappers/update-status.ts` to the camelCase
`UpdateStatus` type.
