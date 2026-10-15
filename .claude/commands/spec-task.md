---
description: Read a GitHub issue and drive it through the full GitHub Spec Kit flow (specify → clarify → plan → tasks → analyze → checklist) to produce a complete spec folder with every open question resolved before implementation.
argument-hint: <issue number, URL or XEN-id> (e.g. 12 or XEN-008)
---

# spec-task — GitHub issue → full Spec Kit spec folder

You are turning **$ARGUMENTS** into a complete Spec Kit feature spec. This repo
is already Spec Kit-initialized (`.specify/` exists). Run **every** phase of
the flow, in order, for real — do not summarize a phase instead of running it,
do not skip a phase because the task "seems simple," and do not merge phases
together. Each phase below maps to an actual skill invocation via the `Skill`
tool; invoke it and let it do its full work before moving to the next.

This is a **single repo**: the specs live in `specs/<XEN-ID>-<slug>/` right
next to the code they describe (`cmd/`, `internal/`, `pkg/`) and the BDD suite
that will verify them (`test/integration/`). Ground every phase in *this* repo's
code — there is no separate spec or e2e repo.

The task flow this command sits in, all on **one branch** (the lowercased
task id, e.g. `xen-008`; see `.claude/skills/github-issues.md`):

```
/research-task → /spec-task → /create-tests → /implement-task
  (issue desc)     (specs/)     (red .feature)   (green code → push → PR → merge on green CI)
```

## Non-negotiable rules

1. **Run the full chain, no shortcuts.** Order is fixed:
   `speckit-specify` → `speckit-clarify` → `speckit-plan` → `speckit-tasks` →
   `speckit-analyze` → `speckit-checklist`. Do not reorder, do not skip a step
   because it seems redundant with what you already know from the issue, and do
   not run `speckit-implement` — this command stops at a complete, analyzed
   spec folder, it does not write product code.
2. **Each phase must genuinely run**, not be hand-rolled by you. Invoke the
   skill via the `Skill` tool (e.g. `skill: "speckit-specify"`) so its full
   internal process (templates, validation loops, hooks) executes. Do not
   write `spec.md`/`plan.md`/`tasks.md` yourself by hand as a substitute.
3. **Zero open questions at the end. This is the definition of done.**
   The spec folder this command produces must be immediately implementable by
   someone who cannot ask anyone anything. When this command finishes, **no
   artifact in `specs/<XEN-ID>-<slug>/` may contain an unresolved decision** — no
   `[NEEDS CLARIFICATION]`, no `[Gap]` left dangling, no `TBD`, `TODO`, `???`,
   `to be decided`, `needs confirmation`, no unchecked checklist item, and no
   task phrased as "investigate/determine/decide whether X". Every such item is
   either **answered from code and data**, or **decided by the user before the
   command completes**. Deferring a question to "the implementer", "PR review",
   "finance", or "whoever picks this up" is a failure of this command, not a
   handoff.
4. **Resolve questions at the phase that raises them.** Whenever a phase
   surfaces a marker or a question — `speckit-specify`'s up-to-3, the
   `[Gap]`/`[Ambiguity]`/`[Conflict]`/`[Assumption]` markers a checklist emits,
   or anything `speckit-analyze` flags — close it before advancing. Order of
   preference, always:
   1. **Answer it from the code or the data.** Read the real source in this
      repo (against a freshly fetched `origin/main`), query the real database. A question the system already answers must never
      reach the user. Most "open questions" are unresearched, not undecidable.
   2. **Ask the user** via `AskUserQuestion` if it is genuinely a product,
      scope, priority, policy, or risk-tolerance decision. Batch related
      questions into one call (up to 4) so the user answers a short set rather
      than a drip feed.
   3. There is no third option. Do not guess, do not "note it for later", do
      not invent a policy you have no authority to set.
5. **If `speckit-analyze` finds inconsistencies** between spec/plan/tasks, fix
   them (update the relevant artifact) and re-run `speckit-analyze` until it's
   clean, applying rule 4 to anything that needs a decision.
6. **Never encode an unresolved question as an implementation task.** A task
   like "verify whether X holds and add a guard if not" is an open question
   wearing a task costume — it forces the implementer to make the call you
   avoided. Answer it now (rule 4.1: go read the code), then write the task as
   the concrete work that follows from the answer: either "add guard Y at Z" or
   "no guard needed — confirmed enforced at Z". The only legitimately deferred
   work is work whose *outcome is already decided* and merely needs doing.
7. **Do commit and push the spec folder** to the task branch —
   `/create-tests` and `/implement-task` (often run from other sessions /
   worktrees) pick the spec up from that branch, so the artifacts must
   actually be on GitHub when this command finishes. Still **don't open a
   PR** — just push the branch. Only stage the `specs/<XEN-ID>-<slug>/` folder for this commit, nothing else
   you didn't touch (`.specify/feature.json` is git-ignored — it is local to each worktree).
   **Do not commit until Step 7's gate passes** — never push a spec with open
   items in it.

## Step 0 — Read the issue

Follow `.claude/skills/github-issues.md` to resolve **$ARGUMENTS** to one issue and its `XEN-NNN` task id (assign
the next free id if the title has none), then read it:

```bash
gh issue view <n> --repo brienze1/xenon --json number,title,body,labels,state,url,comments
```

- The body holds the description and acceptance criteria; labels and linked issues give context.
- Comments: later clarifications may refine or override the description.

If this task was already run through `/research-task`, the description should
already contain a verified, code-grounded plan (Summary, Affected repos/files,
Technical plan, Out of scope). Prefer that content as your source of truth. If
the description still reads like raw, unverified requirements, that's fine too
— Spec Kit's own clarification phases exist to handle that — but don't invent
technical details that aren't in the issue or the linked code; carry unresolved
ambiguity into `speckit-clarify` rather than guessing it away here.

Compose a single, self-contained **feature description** paragraph from the
task (title + description + relevant comments) — this is what you'll pass as
the argument to `speckit-specify`. Do not just paste the raw issue markdown;
distill it into the plain-language feature description Spec Kit expects (what
users need and why — Spec Kit itself keeps this implementation-agnostic).

## Step 1 — `speckit-specify`

Invoke the `speckit-specify` skill with the feature description from Step 0.
**Pass the task id (e.g. `XEN-008`) so the spec folder is named
`specs/<XEN-ID>-<short-slug>/` (this repo's convention), NOT a sequential
`NNN-` prefix.** Task ids are assigned once, in the issue title, so they never
collide across parallel branches; sequential numbering does. Use the id of the
task this `/spec-task` run is for, uppercased (`XEN-008`), never the issue number.

Let it run its full flow: short-name generation, spec directory creation,
writing `spec.md`, generating `checklists/requirements.md`, and its own
validation loop (up to 3 iterations). If it surfaces `[NEEDS CLARIFICATION]`
questions, close every one of them **inside this phase** under rule 4 — issue
content first, then the code and the database, then the user. No marker may
survive into Step 2.

Before asking the user anything here, check whether production data settles it.
A question like "what should happen for value X?" is often answered by finding
that X has never occurred, or occurs 84 times and therefore matters a lot —
either way that is a fact you can go get, and it makes the question you *do*
ask far sharper.

Note the resulting `SPECIFY_FEATURE_DIRECTORY` and `SPEC_FILE` — every
subsequent phase operates on this same feature.

## Step 2 — `speckit-clarify`

Invoke the `speckit-clarify` skill (no special arguments needed unless Step 0
surfaced a specific area you want it to focus extra attention on). Let it ask
its up-to-5 targeted questions and encode the answers back into `spec.md`.
Answer every question it raises — don't dismiss the phase because you feel the
spec is "clear enough."

This phase has a question budget (5). Spend it only on things the code and data
cannot answer — resolve the rest yourself and say in the report which ones you
settled that way. If the budget runs out with real ambiguity left, that
ambiguity does **not** get to survive: keep resolving it in later steps under
rule 4, and use the Step 7 gate to prove none of it leaked through. The budget
caps how many questions this *phase* asks, not how complete the spec must be.

## Step 3 — `speckit-plan`

Invoke the `speckit-plan` skill to generate the design artifacts (`plan.md`
and whatever supporting design docs the template produces) for this feature.
If the issue (especially after a prior `/research-task` run) already named the real
files/endpoints involved, feed that as guidance so the plan is grounded
in the actual codebase rather than generic.

The plan MUST follow this repo's architecture skills: place every change in
its Clean Architecture layer per `.claude/skills/architecture.md` (domain →
application → integration → infrastructure) and name the BDD feature files
(`test/integration/features/<domain>/*.feature`) that will lock each
requirement — `/create-tests` writes exactly those next.

## Step 4 — `speckit-tasks`

Invoke the `speckit-tasks` skill to generate the dependency-ordered
`tasks.md` from the spec + plan.

## Step 5 — `speckit-analyze`

Invoke the `speckit-analyze` skill to cross-check `spec.md`, `plan.md`, and
`tasks.md` for consistency. This is non-destructive by design — read its
findings and then, in this command, fix anything it flags:
- Ambiguity/gap the artifacts disagree on → correct the wrong artifact.
- Something requiring a product decision → `AskUserQuestion`.
Re-run `speckit-analyze` after fixes until it comes back clean.

## Step 6 — `speckit-checklist` (generate **and then run** it)

Invoke the `speckit-checklist` skill to generate the requirement-quality
checklist(s) appropriate for this feature's domain (e.g. if it touches
payments, security, or data handling, make sure the checklist covers that
domain — pass that as the focus argument).

**Generating the checklist is only half of this step.** A checklist is "unit
tests for the requirements" — so run the tests. Go through every item and
determine whether the requirements actually pass it:

- **Passes** → tick it `[x]`. Where the answer isn't self-evident from the item
  text, add a short clause saying what makes it pass (which section, which
  verified fact).
- **Fails** → the requirements have a defect. Fix `spec.md` / `plan.md` /
  `contracts/` so it passes, then tick it. Apply rule 4: research it first, ask
  the user if it's genuinely their call.

Every box must end up ticked. An unchecked item at the end of this command is
an open question shipped to someone else, which rule 3 forbids. If an item
turns out to be inapplicable, don't leave it blank — tick it and say why it
doesn't apply, or delete it.

Items about policy you don't own (data retention, lawful basis, transmission
and approval of an export, who signs off) are **not** exempt. They are exactly
the kind of thing to put in front of the user via `AskUserQuestion` while you
have their attention. If the user genuinely defers to a third party (a DPO,
finance, legal), that *decision to route it* is itself the resolution: record
in the artifact who must decide, what specifically they must decide, and what
blocks until they do — then tick the item. "Someone should think about this"
is not a resolution; "this is blocked pending X's decision on Y, tracked in
the issue" is.

## Step 7 — Open-items gate (must pass before committing)

Before staging anything, sweep the whole feature folder and prove it is clean.
This is a hard gate, not a formality.

```bash
FEATURE_DIR=specs/<XEN-ID>-<slug>

# 1. Unresolved-decision markers anywhere in the spec folder
grep -rniE '\[NEEDS CLARIFICATION|\bTBD\b|\bTODO\b|\?\?\?|to be decided|to be determined|needs (confirmation|decision|input)|decide later|open question' "$FEATURE_DIR"

# 2. Unchecked checklist boxes
grep -rn '^- \[ \]' "$FEATURE_DIR/checklists/"

# 3. Tasks that are really unanswered questions
grep -rniE '^- \[ \] T.*\b(investigate|figure out|research|look into|assess)\b' "$FEATURE_DIR/tasks.md"
grep -rniE '^- \[ \] T.*\b(determine|decide|confirm|check|establish|find out|evaluate)\b.*\b(whether|if|which|where|what|any|must be added|is needed|exists)\b' "$FEATURE_DIR/tasks.md"
grep -rniE '^- \[ \] T.*\b(note|record)\b.*\bwhether\b|^- \[ \] T.*\bif (so|not|necessary|needed|absent|any)\b' "$FEATURE_DIR/tasks.md"
```

Expected result: **1 and 2 return nothing; all three of 3 return nothing.**

Query 3 is split deliberately: a single naive pattern misses the common shape
where the question verb and the question word are separated, e.g. *"Confirm X
can do Y — check for an existing Z, and note whether one must be added"*. That
is an unanswered design question, not a task. Go read the code and replace it
with the concrete work: "Add field F to entity E" plus a note on what already
exists.

**Verification tasks are legitimate and must not be swept up here.** "Verify
T004 and T005 pass", "Confirm no remaining caller exists — grep for X" and
"re-run the detector and confirm zero rows" all have a decided expected outcome
and describe real work. The distinction is whether the task's *result* changes
what gets built. If it does, it is a question and must be answered now.

Note that `tasks.md` legitimately contains unchecked `- [ ] T###` boxes — those
are *work to be done*, which is the point of the file. Query 2 is deliberately
scoped to `checklists/` only. Query 3 is what catches unanswered questions
hiding among that work.

Two allowances, and only these:

- Prose that *describes* the absence of markers ("no `[NEEDS CLARIFICATION]`
  remains") will trip query 1. Confirm by eye that each hit is descriptive, not
  an actual open item.
- A recorded, owned, and specific external decision — "blocked pending the DPO's
  ruling on retention of the extract; does not block implementation of FR-001–012"
  — is resolved for this command's purposes, because the decision, its owner,
  and its blast radius are all pinned down.

If any hit is a real open item, resolve it under rule 4 and re-run the gate.
**Do not proceed to Step 8 until the gate is clean.**

## Step 8 — Branch, commit, and push the spec

The goal is a branch on GitHub that `/create-tests` (and whoever implements
the task) can link back to from the issue.

1. **Check current branch state first** (`git status`, `git branch
   --show-current`) — a `before_specify` hook, or the harness that created this
   worktree, may already have put you on a branch. The task branch is **the
   lowercased task id** (e.g. `xen-008`) — the same branch `/create-tests`
   and `/implement-task` use, so all three commands stack onto one branch.
   - Already on `xen-008` → reuse it.
   - `xen-008` exists on origin (`git ls-remote --heads origin xen-008`) →
     check it out and build on it; never overwrite its history.
   - Otherwise → `git fetch origin main && git checkout -b xen-008 origin/main`.
     Always branch from a **freshly fetched** `origin/main`, never local `main`.

   The spec **folder** keeps the uppercase id (`specs/XEN-008-<slug>/`); only
   the branch is lowercase.
2. Confirm `git status` shows only the new `specs/<XEN-ID>-<slug>/` folder — do not sweep in unrelated
   in-progress work with a blanket `git add -A`.
3. Commit:
   ```
   git add specs/<XEN-ID>-<slug>
   git commit -m "spec($TASK_ID): add spec/plan/tasks for <short-name>"
   ```
4. Push the branch to `origin` and set upstream:
   ```
   git push -u origin <branch-name>
   ```
5. Confirm this action with the user before pushing if there is anything
   unusual (e.g. the branch already exists on origin with different content,
   or `git status` showed unexpected files) — otherwise this push is the
   expected, requested outcome of the command.

## Step 9 — Link the spec in the issue

Get the GitHub URL for the pushed folder:
```
gh repo view --json owner,name -q '.owner.login + "/" + .name'
```
Construct `https://github.com/<owner>/<repo>/tree/<branch-name>/specs/<XEN-ID>-<slug>`
(link the folder, not individual files, so `spec.md`/`plan.md`/`tasks.md`/
`checklists/` are all reachable from one link).

Post this as a comment on the issue (`gh issue comment <n> --repo brienze1/xenon --body-file <file>`) —
don't overwrite the description (that's `/research-task`'s job). Use the comment format in
`.claude/skills/github-issues.md`, e.g.:

```
🤖 **spec-task**: ✅ SPEC READY

**Branch:** `xen-008`

**Spec:** <github-tree-url>
- spec.md
- plan.md
- tasks.md
- checklists/
```

## Step 10 — Report back

Summarize for the user:
- The feature directory created (`specs/<XEN-ID>-<slug>/`), the branch name, and
  the GitHub link posted to the issue.
- Which phases ran clean vs. required fixes/clarification, and what was
  resolved.
- **Every question that came up and how each was closed** — split into the ones
  you answered from code/data (say which fact settled it) and the ones the user
  decided. This is the evidence that rule 3 held.
- **Explicit confirmation that the Step 7 gate passed**: zero unresolved
  markers, zero unchecked checklist items, zero investigate-style tasks. If any
  external decision was recorded as owned-and-pending, name it, name its owner,
  and state what it blocks.
- The final state: spec, plan, tasks, and checklist all present, consistent,
  fully resolved, and pushed — implementable as-is, with nothing left for the
  implementer to decide. Next in the flow: `/create-tests <XEN-ID>` (red
  BDD on the same branch), then `/implement-task <XEN-ID>`.

Do not describe the spec as complete if anything is still open. If you genuinely
could not close something, say so plainly at the top of the report and state
what it blocks — an honest gap is recoverable, a gap buried under "done" is not.

## Step 11 — Continue the flow automatically

The user wants the task flow to run end to end without being asked. Once the Step 7 gate passed and the spec is
pushed, invoke `/create-tests <XEN-ID>` yourself (Skill tool) on the same branch — do not stop and wait. Stop
instead only if something is genuinely open (a gap you could not close), and say so.

The chain always ends merged: `/create-tests` invokes `/implement-task`, whose last phase opens the pull request
into `main` and merges it once every GitHub Actions check passes (fixing and re-pushing until they do). This
command itself only pushes the spec — the PR covers spec, tests and implementation together.
