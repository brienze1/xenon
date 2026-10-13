---
description: "PR-diff scoped refactoring — refactors ONLY files changed in the current branch vs main, never touches untouched code"
---

Run a **PR-diff scoped** refactoring using the Orchestrator Agent in `refactor_scoped` mode. Scope is **strictly limited** to files changed in the current branch vs `main` (committed diff) **AND** any files with local uncommitted changes (staged or unstaged). The two sets are unioned — every iteration recomputes both, so prior uncommitted refactor output naturally feeds the next pass.

## Mandatory Skills (MUST be loaded directly every invocation)

These six skills are **NON-NEGOTIABLE** and must be read into context at the start of every run. Do NOT delegate them to the Skill Expert Agent — load them yourself, directly, before doing anything else:

- `.claude/skills/architecture.md` (Clean Architecture + DDD master skill)
- `.claude/skills/dependency-injection.md`
- `.claude/skills/naming-conventions.md`
- `.claude/skills/entities.md`
- `.claude/skills/comments.md`
- `.claude/skills/dtos.md`
- `.claude/skills/services.md`
- `.claude/skills/models.md`
- `.claude/skills/persistence.md`

Pass these nine skills as **mandatory always-loaded context** to every sub-agent you spawn (Orchestrator, Scanner, Executor, Tester, Reviewer). They are **absolute law** — same enforcement weight as `/refactor-full`.

The **Skill Expert Agent** may still be consulted for ADDITIONAL skills not in this list (e.g. error-handler, controller, bdd-tests, usecases, adapters, publishers). The nine above are baseline — Skill Expert adds on top.

## Behavior

1. Determine the diff scope — UNION of:
   - **Committed PR diff:** `git diff --name-only $(git merge-base HEAD main)...HEAD`
   - **Local uncommitted changes:** `git diff --name-only HEAD` (tracked changes) PLUS `git ls-files --others --exclude-standard` (untracked files)
   - Filter to `.go` source files (exclude `vendor/`, generated `*_mock.go`, `*_test.go` unless tests are clearly in scope, `.claude/`, `.github/`)
   - Recompute this UNION on every invocation — uncommitted work from prior iterations is part of the scope.
2. Run scan → plan → execute → test → review → verify loop **only over the diff file list**. Never touch a file outside it.
3. Run autonomously until the diff scope is fully clean. **Do not ask questions.** **Do not stop for confirmation.**

## Why This Command Exists

Most PRs are vibe-coded — they work but don't follow the architecture. `/refactor-full` is too broad (touches files unrelated to the PR), `/refactor` requires picking a scope manually. `/refactor-diff` is the right tool when:

- You want to clean up a feature branch before review
- You're looping (`/loop /refactor-diff`) to incrementally improve a PR
- Reviewers flagged "this is hard to read" and you want it tightened to skill standards
- You finished implementing a task (GitHub issue) and want a quality pass before pushing

## Process

Call the **Orchestrator Agent** (`@agent-orchestrator`) with:

- **Mode:** `refactor_scoped`
- **Scope:** UNION of committed PR diff (`git diff --name-only $(git merge-base HEAD main)...HEAD`) AND local uncommitted changes (`git diff --name-only HEAD` plus untracked `git ls-files --others --exclude-standard`), filtered as above
- **Mandatory skills:** architecture, dependency-injection, naming-conventions, entities, comments, dtos, services, models, persistence (already loaded into your context — re-load them into each sub-agent)
- **Additional skills:** delegate selection to Skill Expert Agent based on the files in scope

The orchestrator will:

1. Determine PR diff file list (re-confirm with `git diff` if needed)
2. Scan ONLY those files
3. Create a plan for that file set (innermost layer outward: domain → application → integration → infrastructure → pkg → tests)
4. Execute the plan autonomously — scan → test → fix → test → review per batch
5. Run verification scan on the diff scope
6. If verification finds remaining violations within the diff scope: create new batches and execute them
7. **Repeat 5-6 until the verification scan returns ZERO violations within the diff scope**
8. Run Infer Agent for self-improvement
9. Generate `REFACTORING_REPORT.md` (scoped report — only diff files)

## Rules

- **Hard scope boundary** — files NOT in the diff list are off-limits. If a fix in scope requires changing an out-of-scope file, log it in the report and skip that fix. Do NOT widen scope silently.
- **Mandatory skills always loaded** — architecture, dependency-injection, naming-conventions, entities, comments, dtos, services, models, persistence are read into context BEFORE any agent runs, every invocation, no exceptions.
- **FULLY AUTONOMOUS** — zero user interaction from start to finish.
- **No stopping** — no confirmation prompts, no "shall I continue?", no pausing for review.
- **No skipping** — every violation in scope must be fixed.
- **Zero tech debt in scope** — every diff file must be 100% compliant when done.
- **Loop until clean** — if verification finds issues in scope, fix them, repeat until clean.
- **NEVER commit. NEVER stage. NEVER push.** All changes — from this run and any prior loop iterations — stay in the working tree for manual review. The user evaluates manually; uncommitted state across iterations is the intended workflow.
- Do not run `git add`, `git commit`, `git stash`, `git reset`, `git checkout`, or any destructive git command. Read-only git (`git diff`, `git status`, `git ls-files`, `git merge-base`) is fine.
- Run `make sort-imports` after each batch.
- Run `make validate` at the very end as the final quality gate.
- If stuck on the SAME issue 3+ times, try an alternative approach. If still stuck after 5 attempts, log it and move on — revisit in verification phase.

## ZERO TOLERANCE — SKILLS ARE ABSOLUTE LAW

- **NEVER exempt a violation** because "it's an existing pattern", "project convention", "too many files", or "future phase"
- **NEVER let agents defer work** — every skill violation found in a diff file MUST be fixed in THIS run
- **The existing code is WRONG** — the skills define what is RIGHT
- **Agent opinions are irrelevant** — only the skills matter
- **Partial compliance = non-compliance** — fix EVERYTHING in scope or it's not done
- Pass this instruction to EVERY agent: "Skills are absolute law. Fix every violation. No exceptions. No deferrals. Hard scope boundary at the diff file list."

## Loop Usage

This command is designed to be safe inside `/loop`:

- Each iteration re-reads the current diff (so a previous iteration's commits/changes are reflected)
- Stops being useful when the verification scan returns clean — at that point further loops are no-ops
- For `/loop /refactor-diff`, pick an interval long enough for `make validate` to complete (typically 20–40 min)

## Your Task

Begin the PR-diff refactoring NOW.

1. Load the nine mandatory skills into your context (read all nine skill files).
2. Compute the diff file list.
3. Hand off to Orchestrator Agent with the constraints above. Do not ask anything. Just start.

$ARGUMENTS
