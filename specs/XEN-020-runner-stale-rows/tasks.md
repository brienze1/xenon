---

description: "Task list for XEN-020 — Runner follows the current tree"
---

# Tasks: Runner Follows the Current Tree

**Input**: Design documents from `specs/XEN-020-runner-stale-rows/`

**Prerequisites**: plan.md, spec.md, research.md, data-model.md, contracts/runner-store.md, quickstart.md

**Tests**: MANDATORY (constitution V, BDD-first). The defect lives in the frontend Runner store,
so the red-first proof is Vitest (`store.test.ts`, `bootstrap.test.ts`). The two Godog
scenarios guard backend behaviour that already holds on `main` (research R1, R7): they are
written first and are **expected green** before the implementation; if either is red, the
backend has a defect this plan did not find — fix it in the layer the failing step points to
before continuing.

**Organization**: Tasks are grouped by user story. Paths are repository-relative.

## Format: `[ID] [P?] [Story] Description`

- **[P]**: Can run in parallel (different files, no dependencies)
- **[Story]**: Which user story this task belongs to (US1, US2, US3)

---

## Phase 1: Setup (Shared Infrastructure)

**Purpose**: A worktree where Vitest and Godog run

- [ ] T001 Prepare the frontend test environment on branch `xen-020`: run `npm ci` in `frontend/` and `make bindings` at the repo root (fresh worktrees lack `frontend/bindings/`), then confirm `cd frontend && npx vitest run src/features/runner/store.test.ts src/app/bootstrap.test.ts` passes on the untouched code

---

## Phase 2: Foundational (Blocking Prerequisites)

**Purpose**: Read/edit sequencing in the Runner store that every story's refresh builds on

- [ ] T002 In `frontend/src/features/runner/store.ts`, add module-level `reads: Map<string, number>`, `edits: Map<string, number>` and `rereading: Map<string, boolean>` beside `saveTimers`/`refreshing`/`finishedRuns`, each cleared in `clearTimers()`; bump `edits` for the target inside `scheduleSave(itemId)`; make `load(itemId)` take the next read sequence for the target at its start and apply its response (success or failure patch, `setActive`, `fetchRun`) only while that sequence is still `reads.get(itemId)` (data-model.md "Module-level bookkeeping")
- [ ] T003 In `frontend/src/features/runner/store.ts`, add an internal `flushSave(itemId)` that, when `saveTimers` holds a timer for the target, clears and deletes it and awaits `save(itemId)`; no-op otherwise

**Checkpoint**: existing `store.test.ts` still green (behaviour unchanged for a single load).

---

## Phase 3: User Story 1 - The open Runner picks up tree changes (Priority: P1) 🎯 MVP

**Goal**: Every Runner target read this session re-reads its reconciled rows when the sidebar tree changes, keeping the user's edits, view and runs.

**Independent Test**: With a target loaded (2 rows), a tree change makes `planRun` return 3 rows → the target shows 3 rows without `ensure`/remount; view and shown run unchanged.

### Tests for User Story 1 (MANDATORY — write first, prove red) ⚠️

- [ ] T004 [P] [US1] In `frontend/src/features/runner/store.test.ts`, add `describe('refreshAll')` using the existing `loaded()`, `runPlan`, `planRow` fixtures and `vi.mocked(service.planRun)`, with cases: (a) a loaded target re-reads and replaces `rows`/`config`/`target`/`environments`/`activeEnvironmentId`/`dataFileMissing` (2 → 3 rows, new row `selected: true` as the backend returns it, and a renamed row showing its new `name` at its old position); (b) a target never shown (`status: 'idle'` / absent) is not read; (c) `view: 'results'`, `runId` and `run` stay as they were; (d) a pending row edit (`setRows` then `refreshAll` before `SAVE_DELAY_MS`, fake timers) calls `saveRunnerPreferences` with the edited rows **before** `planRun`, and the debounce timer no longer fires afterwards; (e) an edit made while `planRun` is pending (deferred promise) drops that response and triggers exactly one trailing save + `planRun`, ending with the trailing plan's rows; (f) three `refreshAll` calls while one read is pending produce exactly 2 `planRun` calls in total and end with the last plan; (g) a failure (`planRun` rejects `AppError('RUN-07404')`) leaves `config: null`, `rows: []`, `target: null`, `loadError.code === 'RUN-07404'`; (h) a response that arrives after the active workspace changed (`useWorkspacesStore.setState({ activeLocalId: 'L2' })` while pending) or after `reset()` patches nothing; (i) a first `load` response arriving after a newer refresh response is ignored. Run them and record that they fail (red) because `refreshAll` does not exist
- [ ] T005 [P] [US1] In `frontend/src/app/bootstrap.test.ts`, add a case next to "reloads the sidebar when the active workspace tree changed outside Xenon": spy `useRunnerStore.getState().refreshAll` (`vi.spyOn(...).mockResolvedValue()`, import `useRunnerStore` from `@/features/runner/store` the way the file imports other stores), trigger `handlers.treeChanged?.({ localId: 'A' })` with `tree.getTree` resolving a new tree, and `vi.waitFor` the spy to have been called; also assert it is called after `useSidebarStore.getState().setTree([...])` (a UI-originated tree change, which emits no backend event), and that `handlers.treeChanged?.({ localId: 'other' })` (another workspace, FR-008) does not call it. Prove red
- [ ] T006 [P] [US1] In `test/integration/features/runner/runner_plan.feature`, add a scenario tagged `@xen_020 @reconcile` "A subfolder and requests added after the rows were remembered join selected, in sidebar order": reuse the Background's Shop/Orders tree, save preferences via `RunnerService.SavePreferences` with `get` unchecked, then write (with the existing "the workspace file … contains" step) a new request into `orders/` listed after `create.yaml` in `orders/folder.yaml` and a new subfolder `shop/payments/` (`folder.yaml` + two requests) appended to `shop/collection.yaml`'s `order`; call `RunnerService.Plan` and assert the row count, each new row's `request_id` position after the nearest preceding request in sidebar order, `selected` `true` for new rows, `get` still `false`, and `number` renumbered. Use only existing steps from `test/integration/steps/`. Run `make bdd TAG=@xen_020` — expected green (guard, see finding R7)

### Implementation for User Story 1

- [ ] T007 [US1] In `frontend/src/features/runner/store.ts`, add the internal `refresh(itemId)` per contracts/runner-store.md steps 1–7: in-flight guard via `rereading` (mark trailing and return), `await flushSave(itemId)`, capture `edits.get(itemId)`, take the next read sequence, call `planRun` through `inWorkspace`; drop the response when `inWorkspace` returned `null`, the sequence is no longer `reads.get(itemId)`, or `edits` moved (then mark trailing); on success patch `target`, `rows`, `config`, `environments`, `activeEnvironmentId`, `dataFileMissing`, `loadError: null` and `setActive(itemId, plan.activeRun)`; on failure patch `{ config: null, rows: [], target: null, loadError }`; in `finally` clear the in-flight flag and run one trailing `refresh` if marked
- [ ] T008 [US1] In `frontend/src/features/runner/store.ts`, add `refreshAll: () => Promise<void>` to `RunnerState` with the doc comment `/** Reads every Runner target again (after the workspace tree changed). */` and implement it as `Promise.all` of `refresh(itemId)` over `targets` whose `status !== 'idle'`; T004 green
- [ ] T009 [US1] In `frontend/src/app/bootstrap.ts`, inside the `useSidebarStore.subscribe` branch `state.tree !== previous.tree && state.status === 'ready'`, add `void useRunnerStore.getState().refreshAll()` after the documentation refresh; T005 green

**Checkpoint**: US1 complete — an open Runner follows sidebar, AI, git and external tree changes.

---

## Phase 4: User Story 2 - Runs started from the Runner include current requests (Priority: P1)

**Goal**: `start()` sends the refreshed rows.

**Independent Test**: After a refresh to 3 rows, `start('c1')` calls `startRun` with a config whose `rows` are the 3 refreshed rows.

### Tests for User Story 2 (MANDATORY — write first, prove red) ⚠️

- [ ] T010 [US2] In `frontend/src/features/runner/store.test.ts` `describe('refreshAll')`, add: after `loaded()` (2 rows) and `refreshAll()` resolving a 3-row plan (one row removed, two added), `start('c1')` calls `service.startRun` with `config.rows` equal to the refreshed rows (request ids, alias ids, selection) in order. Written alongside T004 it is red before T007–T008

### Implementation for User Story 2

- [ ] T011 [US2] No production change: `start()` in `frontend/src/features/runner/store.ts` already sends `configWithRows(target.config, target.rows)`; confirm T010 is green after T007–T008

**Checkpoint**: US1 + US2 — the reported repro (64 → 93 rows, run sends 93) is fixed.

---

## Phase 5: User Story 3 - Headless runs include current requests (Priority: P2)

**Goal**: Lock that an MCP `run` `start` without `rows` uses the remembered rows reconciled with the current tree.

**Independent Test**: `make bdd TAG=@xen_020` passes the MCP scenario.

### Tests for User Story 3 (MANDATORY — write first) ⚠️

- [ ] T012 [P] [US3] In `test/integration/features/mcp/runner.feature`, add a scenario tagged `@xen_020` "A run started without rows includes requests added after the rows were remembered": save preferences for Shop with `logout` unchecked (`RunnerService.SavePreferences`, as the scenario "A run started in the UI is readable over MCP…" does), write a new request file into `orders/` and add it to `orders/folder.yaml`'s `order`, mock its endpoint like the Background's requests, call the MCP `run` tool with `{ "action": "start", "item_id": "{{shop}}" }` (no `rows`), wait for the run to finish and read it with the `runs` tool the way "An AI client starts a run with defaults and reads it once finished" does, then assert the start response's `planned` equals the remembered selected rows plus the new request, and "the mock API should have received the requests … in order" lists the new request after the nearest preceding request in sidebar order and no `POST /logout`. Use only existing steps. Run `make bdd TAG=@xen_020` — expected green (guard, see findings R1 and R7)

### Implementation for User Story 3

- [ ] T013 [US3] No production change: `internal/application/service/start_run.go` `StartForAI` → `runPlanner.plan` already reconciles against `treePersistence.Load` (disk rescan); T012 green confirms it

---

## Phase 6: Polish & Cross-Cutting Concerns

- [ ] T014 Run the full frontend gate in `frontend/`: `npm run lint`, `npm run typecheck`, `npm test` — all green
- [ ] T015 Run `make validate` (Go unit tests, full BDD suite including `@xen_020`, linters) — all green
- [ ] T016 Walk the manual steps in `specs/XEN-020-runner-stale-rows/quickstart.md` in `wails3 dev` (Runner open, add requests from the sidebar and from an AI client, delete one, delete the folder, Run) and confirm each expected outcome, including SC-002 (the Runner shows new rows within 1 s of the sidebar)

---

## Dependencies & Execution Order

### Phase Dependencies

- Setup (T001) → Foundational (T002–T003) → US1 (T004–T009) → US2 (T010–T011) → US3 (T012–T013, independent of the frontend; can run any time after T001) → Polish (T014–T016)

### User Story Dependencies

- **US1**: needs T002–T003.
- **US2**: needs US1's `refreshAll` (T007–T008) to have refreshed rows to send.
- **US3**: backend guard only; independent of US1/US2.

### Within Each User Story

- Tests first and run (red for Vitest, green-guard for Godog), then implementation, then re-run.

### Parallel Opportunities

- T004, T005, T006 and T012 touch different files and can be written in parallel.
- T007 → T008 → T009 are sequential (T009 calls the action T008 adds; T007/T008 share `store.ts`).

## Parallel Example: User Story 1

```text
Task: "T004 refreshAll cases in frontend/src/features/runner/store.test.ts"
Task: "T005 Runner refresh case in frontend/src/app/bootstrap.test.ts"
Task: "T006 @xen_020 guard scenario in test/integration/features/runner/runner_plan.feature"
```

## Implementation Strategy

### MVP First (User Story 1 Only)

1. T001–T003, then T004–T009 → the Runner follows the tree. Validate with T004/T005 green.

### Incremental Delivery

1. US1 → open Runner is current.
2. US2 → runs from the Runner send current rows (test only).
3. US3 → headless guard.
4. Polish → full gates and manual quickstart.

## Notes

- No Go production code changes in this task; if a Godog guard is red, that contradicts research R1 and must be fixed in the backend layer it points to before moving on.
- FR-009 needs no new task: the only write path is the unchanged `RunnerService.SavePreferences` (local UI state), already locked by the XEN-017 `runner_plan.feature` scenarios that check nothing is written to workspace files.
- Do not merge rows on the client — reconciliation stays in Go (`entity.ReconcileRunRows`).
