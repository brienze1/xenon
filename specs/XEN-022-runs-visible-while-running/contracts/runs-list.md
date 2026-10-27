# Contract: listing runs

Applies to the Wails binding `RunnerService.List({ local_id, item_id })` and the MCP tool
`runs` `{ "action": "list", "target" | "path" | "item_id" }`. Response shape unchanged (array of
run summary views: `id`, `item_id`, `target_kind`, `target_name`, `type`, `status`,
`environment_name`, `started_at`, `finished_at`, `progress`, `summary{…}`).

## Content (new)

1. Every stored run of the target **and** the target's in-progress run (status `running` or
   `paused`), if any.
2. Ordered by `started_at` descending — the in-progress run is element `0`.
3. Each `id` at most once; a run both stored and in progress is returned as stored.
4. An in-progress element's `summary` reflects the rows recorded so far
   (`total_requests` = rows done, `tests_passed/failed/skipped` so far, `duration_ms` = elapsed).
5. Runs of other targets / workspaces are never included.
6. Any `id` that `RunnerService.Get` / MCP `runs get` returns for the target is in the list.

Errors unchanged (unknown workspace, unknown target).
