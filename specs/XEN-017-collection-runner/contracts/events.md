# Contract: Runner events (Go → React)

The event types are registered in `internal/infra/server/event/events.go`. Their DTOs live in
`internal/integration/publisher/dto/`. The publisher is
`internal/integration/publisher/run_event.go`, behind the use case
`usecase.PublishRunEvent`. React subscribes in `frontend/src/platform/events.ts`, and
`app/bootstrap.ts` wires the subscriptions once.

## `runner:progress` — `dto.RunProgressChanged`

```json
{ "local_id": "…", "run_id": "…", "item_id": "…", "type": "functional", "status": "running",
  "progress": { /* as dto.Run.progress */ } }
```

- Functional runs emit it after every executed row and on every status change (pause, resume).
- Performance runs emit it once per second, with `progress.metrics = {active_users, requests,
  avg_ms, throughput, error_rate}` (FR-045).
- It is throttled to at most 10 events per second per run. The last state before
  `runner:finished` is always emitted.

## `runner:finished` — `dto.RunFinished`

```json
{ "local_id": "…", "run_id": "…", "item_id": "…", "target_name": "Orders API", "type": "functional",
  "status": "completed", "summary": { /* dto.Run.summary */ } }
```

It is emitted once per run, after the run was stored. React then:

- refreshes the target's Runs list;
- clears the tab's running indicator;
- shows the finish toast (FR-001b);
- if the Runner › Results view shows this run, loads it with `RunnerService.Get`.

## Reused events

- **`console:entries`**: unless logs are off, a run's entries reach the app console as they do for
  a single send. They are tagged with each row's send id.
- **`variables:changed`**: emitted only when keep-values persists changes, as for a single send.
- **`cookies:changed`**: emitted only when save-cookies merges changes in at the end.
- **`history:changed`**: never emitted by a run.
