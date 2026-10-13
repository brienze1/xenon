---
description: "Implement a task end-to-end (GitHub issue or free description): research → plan → branch → red BDD → green impl → refactor-diff loop → push (no PR)"
---

You are implementing a task end-to-end. Execute the phases below **strictly in order**. Do not skip a phase, do not reorder, and do not open a pull request at any point.

This is a **single repo**: the service code (`cmd/`, `internal/`, `pkg/`), its BDD suite (`test/integration/`) and its Spec Kit specs (`specs/`) all live here. Everything — spec, tests, implementation — ships on **one branch**.

The full task flow, all on that branch:

```
/research-task → /spec-task → /create-tests → /implement-task
  (issue desc)     (specs/)     (red .feature)   (green code → push)
```

Earlier steps are optional — this command works from a bare issue or description too — but when they ran, **their output is your starting point**: the spec in `specs/<XEN-ID>-<slug>/` is the source of truth for *what* to build, and the red features `/create-tests` pushed are the tests you must turn green.

---

## Input & mode detection

Argument: $ARGUMENTS

Determine the input mode:

- **Issue mode** — the argument is (or contains) a GitHub issue reference: a number (`12`, `#12`), an issue URL, or a task id like `XEN-008`. Resolve it per `.claude/skills/github-issues.md`. The task to implement is that issue.
- **Description mode** — the argument is a free-text description of work with no issue reference. There is no issue.

Mode differences (everything else is identical):

| | Issue mode | Description mode |
|---|---|---|
| Source of truth | `specs/<XEN-ID>-<slug>/` if present, else the issue (+ comments) | the supplied description |
| Label the issue **in progress** | **yes** | n/a — never create or change an issue |
| Branch name | the lowercased task id (e.g. `xen-008`) | the branch the user explicitly named; if none given, **ask** before Phase 3 |
| Final issue comment | **yes** | n/a |

In description mode: **do not create an issue and do not change or comment on any issue.** Only use a specific branch name if the user stated one — otherwise ask which branch name to use before Phase 3.

---

## Phase 1 — Research & verify the work is correct

Goal: confirm the work matches reality **before** writing anything. Specs (issue or description) are often written from stale assumptions — catch that now, not mid-implementation.

**Issue mode only — first action:** once you've confirmed the argument is a real issue and you're committing to work it, add the **in progress** label (commands in `.claude/skills/github-issues.md`). Then continue.

1. Load the spec:
   - **Issue mode:** `gh issue view <n> --repo brienze1/xenon --json number,title,body,labels,state,url,comments` — the full description, the task id in the title, and the comments (decisions and clarifications added after creation).
   - Check whether the task branch already exists on origin (`git fetch origin && git ls-remote --heads origin <branch>`). If it does, read what earlier steps left there:
     - `specs/<XEN-ID>-<slug>/` — `spec.md`, `plan.md`, `tasks.md`, `contracts/`. **The spec wins over the raw issue** where they disagree (note the discrepancy).
     - `git diff --stat origin/main...origin/<branch> -- test/integration/` — the red features `/create-tests` wrote.
   - **Description mode:** treat the supplied description as the spec. If it references companion docs you can reach, read them.
2. Verify the spec against the **freshly fetched `origin/main`** (`git fetch origin main`, then `git show origin/main:<path>` / `git grep … origin/main`) — do not trust a local checkout to be current:
   - Confirm the endpoints, DTOs, tables, fields, and behaviours described actually exist as stated (or, for greenfield work, that nothing conflicting exists).
   - Check whether any claim ("today the response returns X", "field Y exists") is actually true.
3. Produce a short **findings report**: what's accurate, what's wrong or stale, what's ambiguous, and any contract decisions left open.
   - If you find material gaps (premise wrong, missing decision, conflicting specs), surface them and ask the user how to proceed **before** planning. Do not silently "fix" the spec's intent.
   - If the spec checks out, say so explicitly and continue.

---

## Phase 2 — Plan (plan mode)

1. Enter plan mode (EnterPlanMode).
2. Plan **every** change needed, per layer (domain → application → integration → infrastructure → tests), following the Clean Architecture skills in `.claude/skills/`. If `plan.md` / `tasks.md` exist, the plan follows them — call out any deviation and why. Include:
   - the exact files/DTOs/services/validators/migrations to add or change,
   - the BDD scenarios that lock each expectation (by feature file + tag) — the ones `/create-tests` already wrote, plus any it missed,
   - contract/versioning notes (additive vs `/v2`), risk, and rollback.
3. Use AskUserQuestion for any genuinely open contract decision the research surfaced. In description mode, this is also where you confirm the branch name if the user didn't give one.
4. Exit plan mode (ExitPlanMode) and get the plan approved before touching code.

---

## Phase 3 — Branch (fresh from main) — **MANDATORY FETCH FIRST, EVERY TIME**

> **NON-NEGOTIABLE: you MUST `git fetch origin` before doing anything else in this phase, on every single run, no exceptions.** The local `main` branch and the local `origin/main` remote-tracking ref are **both untrustworthy** — they are whatever they were the last time someone fetched, which may be hundreds of commits stale. **Never trust local `main`. Never trust a local `origin/main` ref you have not just refreshed.** A "0 commits ahead / clean tree" check performed *before* a fetch is worthless.

Work in a dedicated worktree (the harness usually pre-creates one for the session — use it; otherwise `git worktree add`). Never reuse a dirty main checkout.

1. **Fetch first — always, before any inspection or branching:**
   ```bash
   git fetch origin
   ```

2. **Get onto the task branch, based on the freshly-fetched `origin/main`:**
   - **Branch exists on origin** (pushed by `/spec-task` / `/create-tests`): check it out (or reset the pre-created worktree branch to it if that branch has no work of its own), then rebase onto the fresh main, preserving the spec and test commits:
     ```bash
     git checkout <branch>            # or: git checkout -B <branch> origin/<branch>
     git rebase origin/main
     ```
   - **Branch only exists locally / pre-created by the harness:** rebase it onto `origin/main`.
   - **No branch yet:** `git checkout -b <branch> origin/main`.

   Resolve any conflicts now, before implementing. Never start work on a branch whose base predates the current `origin/main`.

3. **Verify freshness against the fetched ref — and prove it, do not assume:**
   ```bash
   git fetch origin main
   git rev-list --count HEAD..origin/main   # MUST be 0
   git merge-base --is-ancestor origin/main HEAD && echo FRESH || echo BEHIND   # MUST print FRESH
   ```
   Only proceed once it is genuinely fresh.

---

## Phase 4 — Red BDD scenarios

The BDD suite in `test/integration/` is part of this task and changing it is mandatory. Follow `bdd-testing.md`, `bdd-feature-files.md`, `bdd-step-definitions.md`, `testing-setup.md` and the relevant `mock-*.md` skills.

1. **If `/create-tests` already pushed features for this task:** review them against the approved plan. Keep them — they are the QA contract — and only *add* scenarios for expectations they missed. Do not weaken or delete an assertion to make your implementation's life easier; if you believe one is genuinely wrong, stop and ask the user.
2. **Otherwise:** write new scenarios (or extend existing feature files) that encode **every expectation**, following the suite's existing conventions.
3. **Existing feature files whose fixtures your change invalidates MUST be updated in this branch.** If a contract change makes an old scenario's setup illegal, rewriting it is part of the work, not a follow-up.
4. If the harness cannot express the new behaviour, extend the step definitions and mocks **generically** and additively — then **re-run the full suite**, because shared test infrastructure is exactly where a change quietly breaks unrelated scenarios.
5. Run only the task's scenarios (`make bdd TAG=@<slug>`) and **confirm they FAIL (red)** — proving they actually test the new behaviour. Quote the red output.
6. If a scenario passes before any implementation, it's not testing the new behaviour — fix the scenario, not the assertion.

---

## Phase 5 — Implement to green

1. Implement the code per the approved plan, innermost layer outward.
2. Run the task's scenarios until they **PASS (green)**.
3. Then run the **full** suite (`go test ./...` + `make bdd`) and keep everything green — including pre-existing tests. If a pre-existing test fails, verify on a clean `origin/main` checkout whether it's your regression or pre-existing, and say which.
4. Keep build/vet clean throughout.
5. If `tasks.md` exists, tick off (`[x]`) the tasks you completed, so the spec reflects what was built.

---

## Phase 6 — `/refactor-diff` in a 10-minute loop until clean

1. Start a session loop: `/loop 10m /refactor-diff` (or run `/refactor-diff` directly on a 10-minute cadence).
2. Each iteration runs the PR-diff scoped refactor over the union of the committed diff vs `main` and uncommitted changes.
3. **Stop the loop when an iteration returns a CLEAN verdict** (zero fixable violations) — cancel the cron/loop at that point. Do not keep looping on a clean tree.
4. Re-verify the full suite is green after the final refactor pass.

---

## Phase 7 — Ignore the refactoring report

1. `/refactor-diff` / `/refactor-full` generate `REFACTORING_REPORT.md` (and may emit other `*REFACTORING*` artifacts). These must **never** be committed.
2. Ensure `.gitignore` contains an entry for it (it does at repo bootstrap — re-add if missing). Add any other generated refactor artifacts you observe.
3. If a report file is already staged/tracked, untrack it (`git rm --cached`) — but do not delete other work.

---

## Phase 8 — Push (NO PR)

1. Commit the work with clear, scoped messages (one logical commit per concern is fine; reference the task id, e.g. `feat(XEN-008): …`, in issue mode). Spec/test commits from earlier steps stay as they are.
2. Push the branch:
   ```bash
   git push -u origin <branch>
   ```
   If the branch was rebased in Phase 3 and already existed on origin, use `git push --force-with-lease` — never a bare `--force`.
3. **Do NOT open a pull request.** Leave the branch pushed and available for review.
4. Report back: branch name, commit SHAs, what shipped, test status, and anything logged-but-skipped (out-of-scope findings, pre-existing failures).
   - **Issue mode:** comment on the issue with what was built (format per `.claude/skills/github-issues.md`), calling out anything that diverged from the description. Do not close the issue; it closes when the work merges to `main`.
   - **Description mode:** report back to the user only — touch no issue.

---

## Rules

- Phases run in the exact order above. Research and plan-approval gate all code.
- **One repo, one branch**: spec, tests and implementation all land on the task branch.
- **Issue mode:** label the issue in progress at the start; branch = lowercased task id; comment on the issue at the end.
- **Description mode:** never create or change an issue; branch = the one the user named (ask if none given).
- The spec (`specs/<ID>-<slug>/`), when present, is the source of truth for what to build; the features `/create-tests` wrote are the contract — extend, never weaken.
- **MANDATORY: `git fetch origin` FIRST, every run — never trust local `main` or an un-refreshed `origin/main` ref; verify `0` behind / `FRESH` before writing code, including when the worktree was pre-created.**
- Red before green — always prove the test fails first.
- Loop `/refactor-diff` until CLEAN, then stop.
- `REFACTORING_REPORT.md` is git-ignored, never committed.
- Push, never open a PR.
- Honour the repo's Clean Architecture skills and BDD-first discipline throughout.
