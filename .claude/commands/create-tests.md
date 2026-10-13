---
description: Author/update red BDD feature files in this repo from a GitHub issue (and its Spec Kit spec), reusing the generic step catalog.
argument-hint: <issue number, URL or XEN-id> (e.g. 12 or XEN-008)
---

# create-tests — BDD features from a GitHub issue

You are acting as the **QA engineer** for this task. Given the GitHub issue
**$ARGUMENTS**, write (or update) godog `.feature` files under
`test/integration/features/` that validate the **real expectations** described
in the task and its spec.

This is a **single repo**: the code under test (`cmd/`, `internal/`, `pkg/`),
the BDD harness (`test/integration/`) and the spec (`specs/`) all live here.
There is no separate e2e repo, no live stack, and no catalog to keep in sync —
features run in-process against the app with only the outermost boundaries
faked (see `.claude/skills/bdd-testing.md`).

Before writing anything, read these skills — they are the law for this command:
`bdd-testing.md`, `bdd-feature-files.md`, `bdd-step-definitions.md`,
`testing-setup.md`, and the `mock-*.md` skill for every external boundary the
task touches.

## Non-negotiable rules

1. **Reuse the step catalog; extend it only generically.** Every
   `Given/When/Then/And` line must match a step registered in
   `test/integration/steps/`. Re-model the assertion with existing steps first
   (query the DB, assert a response field, assert a mock received a call). Only
   if that genuinely cannot express it, add a step that is **generic** — it
   takes what it operates on as arguments and knows nothing about the business
   domain (`I set the variable "<name>" to a new uuid` is the bar; a step named
   after a domain concept belongs in the `.feature` as ordinary steps, not in
   the catalog). New steps follow `bdd-step-definitions.md` (file split, regex
   params) and must not break any existing scenario.
2. **Validate expectations, not the current implementation.** The tests encode
   what the system *should* do per the task/spec. If the feature isn't
   implemented yet, the tests are expected to be **red** — that is the point.
   Do NOT weaken an assertion to make it pass. Do NOT assert current (wrong)
   behavior just because that's what the code does today.
3. **Reuse existing features as the base.** Before writing anything, find the
   closest existing scenarios and copy their structure, tags, setup, and
   idioms. Match the surrounding style exactly — do not invent new conventions.
   (While the suite is still small, `bdd-feature-files.md` is the template.)
4. **Prefer updating over duplicating.** If a feature already covers the area
   the task touches, extend/adjust it. Only create a new `.feature` when the
   task introduces a genuinely new surface.
5. **Every scenario is self-contained.** It creates every piece of state it
   reads (DB rows, mock API responses, queue messages, clock) and never relies
   on state left by another scenario. The litmus test: `make bdd TAG=@<slug>`
   run *alone* passes/fails exactly like it does in the full suite.
6. **Only fake the outermost boundary.** Databases, external HTTP APIs,
   queues, storage, time — via the mocks in `test/integration/mock/`. Never
   mock persistence, services, use cases or controllers.
7. **Don't touch production code.** This command writes features (and, under
   rule 1, generic steps / mock plumbing in `test/integration/`). Making them
   green is `/implement-task`'s job.

## Step 1 — Read the task and its spec

Follow `.claude/skills/github-issues.md` to resolve **$ARGUMENTS** to one issue and its `XEN-NNN` task id, then
read it:

```bash
gh issue view <n> --repo brienze1/xenon --json number,title,body,labels,state,url,comments
```

The body holds the description and acceptance criteria; clarifications and scope changes often live in the comments.

Then get onto the task branch (lowercased id, e.g. `xen-008`) — the one
`/spec-task` pushed:

```bash
git fetch origin
git ls-remote --heads origin xen-008   # exists → check it out / rebase onto it
```

If the branch doesn't exist yet, create it from a freshly fetched
`origin/main` (`git checkout -b xen-008 origin/main`).

**The Spec Kit spec is the primary source** when it exists:
`specs/<XEN-ID>-<slug>/` (`ls specs/ | grep -i <XEN-ID>`). Pull functional
requirements and acceptance scenarios from `spec.md`, the concrete technical
shape (endpoints/fields/tables) from `plan.md` and `contracts/`, and the
feature files it promised from `tasks.md`. If the spec and the raw issue
description disagree, the spec wins (it's the researched version) — but note
the discrepancy rather than silently picking one. No spec → fall back to the
issue description/comments.

Extract, in your own words: **what changes**, **the exact expected value(s)**,
**the scope boundary** (what must stay UNCHANGED), and any **edge cases /
bands / enums** implied.

## Step 2 — Read the code

Confirm the contract in this repo's source — routes in the router, DTOs,
controllers, models/migrations, enums, error codes:

```bash
grep -rn '<symbol-or-route>' internal/ pkg/ cmd/
```

**Never guess** a field name, error code, enum value, or DB column. If the
thing doesn't exist yet (greenfield), take it from `plan.md` / `contracts/`
exactly as specified — the red test is what will force the implementation to
match it.

## Step 3 — Find the base scenarios & decide create-vs-update

Features are organised by **domain folder** under `test/integration/features/`.
Work out which domain the task belongs to and stay in it — new files go in the
matching folder, next to their siblings.

```bash
ls test/integration/features/
grep -rln "<endpoint-or-table-or-keyword>" test/integration/features/
grep -rhoE 'ctx\.Step\(`\^[^`]+`' test/integration/steps/*.go | sed -E 's/ctx\.Step\(`\^//; s/`$//'
```

Pick 1–3 nearest scenarios as templates. Decide: extend an existing feature, or
add a new one. State the decision briefly before editing.

## Step 4 — Author the feature(s)

Follow `bdd-feature-files.md`. For each expectation:

- **Observable value** (status code, response field, DB row) → assert it
  directly.
- **Outbound effect** (a call to an external API, a published message, an
  uploaded object) → assert it on the corresponding mock (request received,
  body/headers, message payload). Arrange mock responses in the scenario
  itself, including the failure modes the task cares about.
- **Scope guards** — whenever the task says "X changes but Y stays the same",
  add a scenario proving **Y is unchanged** (the negative case). This is the
  most commonly-missed coverage.
- **Edge cases** — cover what the task's domain actually exposes: boundary
  values, enum members, empty / missing / NULL inputs, validation errors and
  their codes, authz boundaries, pagination & filtering, idempotent retries,
  terminal-state guards. Pick the ones that apply.

Use `Scenario Outline` + `Examples` when only values vary. Tag each new
scenario with a unique snake_case slug for the task (so `make bdd TAG=@<slug>`
runs exactly your work), and copy the feature-level tag style from sibling
features.

## Step 5 — Self-check

1. Every step you wrote resolves to a registered definition (dump the regexes
   with the `grep` in Step 3 and match each line).
2. Run your tag, then the full suite:
   ```bash
   make bdd TAG=@<your-slug>
   make bdd
   ```
   - Your new scenarios: **functional failures are expected** (red) when the
     behaviour isn't implemented. An **undefined step**, a **parse error**, or
     a **panic in the harness** is not acceptable — fix it.
   - If a new scenario is **green before implementation**, it isn't testing the
     new behaviour — fix the scenario, not the assertion.
   - Every pre-existing scenario must still pass (shared steps/mocks you
     extended are where unrelated breakage hides).

## Step 6 — Commit and push to the task branch

`/implement-task` usually runs in a different session/worktree and picks the
features up from the branch, so they must be pushed:

```bash
git add test/integration/
git commit -m "test(<XEN-ID>): red BDD scenarios for <short-name>"
git push -u origin xen-008
```

Stage only `test/integration/` — never sweep in unrelated work. **Do not open a
PR.**

## Step 7 — Report back

- Feature files created/updated, and any generic steps/mocks added (and why
  the existing catalog couldn't express it).
- Which scenarios are red (expected — not implemented yet) vs green, with the
  one-line reason for each red.
- Any discrepancy between spec and issue you noticed.
- Next in the flow: `/implement-task <XEN-ID>`.
