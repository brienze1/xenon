# Contract: Runner store (frontend)

`frontend/src/features/runner/store.ts` — behaviour other features and components rely on.

## `applyProgress(event: RunProgressEvent)`

Unchanged guards: ignored for another workspace or a run already finished (`finishedRuns`);
`active[itemId]` always updated.

For a target whose Runner was read this session (`status !== 'idle'`):

| Condition | Effect |
| --- | --- |
| `runs !== null` and `runs` lacks `event.runId` | `loadRuns(itemId)` |
| `runId !== event.runId` and (`runId === null` or the run was not yet `active[itemId]`) | `runId = event.runId`, `run = null`, `view = 'results'`, read the run |
| `runId !== event.runId` otherwise (user opened another run during this one) | nothing more |
| `runId === event.runId` | unchanged: patch shown run's status/progress, read functional runs again |

## `clearRuns(itemId)`

On success `runs` keeps only in-progress entries (`isRunInProgress`), instead of `[]`.

## `isRunInProgress(run)` — `features/runner/lib/results.ts`

`run.status === 'running' || run.status === 'paused'`.

## `RunsList`

- No actions menu (Delete) on an in-progress row.
- **Clear runs** disabled when no listed run is finished.
