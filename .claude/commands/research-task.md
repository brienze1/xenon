---
description: Read a GitHub issue, research where the real changes must land in this repo, ask the user to resolve open questions, then rewrite the issue description with the verified plan.
argument-hint: <issue number, URL or XEN-id> (e.g. 12 or XEN-008)
---

# research-task — turn a GitHub issue into a verified, code-grounded plan

You are researching **$ARGUMENTS**. The issue description on GitHub was written
before anyone read the code — it may contain guesses, stale assumptions, or
gaps. Your job: replace those with what the code actually says, close every
open question with the user, then rewrite the issue description so it is
objective, accurate, and immediately actionable for whoever implements it.

This is a **single repo**: the service code (`cmd/`, `internal/`, `pkg/`), its
BDD suite (`test/integration/`) and its Spec Kit specs (`specs/`) all live here.
The plan you produce lands in this repo — there are no sibling service, e2e, or
spec repos to split the work across.

## Non-negotiable rules

1. **Don't implement anything.** This command produces a plan, not code. Do
   not edit files, do not open PRs.
2. **Don't guess.** Every claim in the final plan (endpoint, field name, table,
   enum, package, file) must be confirmed by reading actual source, not
   inferred from the task title or general knowledge of the domain.
3. **Ask, don't assume.** If research surfaces more than one plausible
   approach, an ambiguous requirement, or a decision only the user can make
   (scope, priority, contract shape), stop and ask via `AskUserQuestion` before
   finalizing the plan. Don't pepper the user with questions that the code
   already answers.
4. **The rewritten description must be objective and clear**, not a wall of
   research notes: what changes, where, why, and what's explicitly out of
   scope. Remove assumptions the research disproved — don't leave them
   alongside the correction.
5. **Do not commit, push, or open PRs.** The only artifact this command
   produces is the updated issue description.

## Step 1 — Read the issue

Follow `.claude/skills/github-issues.md` to resolve **$ARGUMENTS** to one issue and its `XEN-NNN` task id (assign
the next free id if the title has none), then read it:

```bash
gh issue view <n> --repo brienze1/xenon --json number,title,body,labels,state,url,comments
```

- The body is the current description; labels and linked issues/PRs give context.
- Comments: later clarifications often override the original description; treat comments as higher-priority
  truth.

Extract, in your own words:
- **What outcome is wanted** (the actual behavior/bug/feature).
- **Every factual claim already in the description** (endpoints, field names,
  "current behavior is X", "this lives in package Y") — these are exactly the
  assumptions Steps 2–3 must verify or refute.
- **What's unclear or unstated** — candidate questions for Step 4.

Also check whether a spec already exists for this task: `ls specs/ | grep -i
<XEN-ID>`. If it does, read it — it may already settle questions.

## Step 2 — Find where the change actually lives

Research against a **freshly fetched `origin/main`**, not whatever the local
checkout happens to be on:

```bash
git fetch origin main
git grep -n '<keyword-or-symbol>' origin/main -- cmd internal pkg test
git show origin/main:<path>
```

Use `.claude/skills/architecture.md` to reason about *which layer* owns the
behavior (domain / application / integration / infrastructure), then narrow to
the specific files, handlers, use cases, models and migrations.

If the task names a branch, read that branch's state (`git fetch origin
<branch>` then `git show origin/<branch>:<path>`) — in-flight work may already
address (or contradict) part of the task.

**External systems.** If the task depends on a third-party API or another
service this repo calls (or is called by), read the webservice/adapter code in
this repo that talks to it first. Only reach outside the repo (vendor docs via
Context7, or `gh` for another GitHub repo the user points you to) when the
contract genuinely isn't knowable from here — and say so in the plan.

## Step 3 — Confirm the technical shape of the change

For each area touched, read enough real source to state with confidence:
- The exact current behavior (endpoint, handler, query, condition) — quote it,
  don't paraphrase from the task description. For greenfield work, state
  plainly that nothing exists yet and what the nearest existing pattern is.
- What must change and where (layer, package, file, migration, config),
  grounded in what you just read and in the architecture skills.
- Which BDD features cover the area today (`grep -rln '<keyword>'
  test/integration/features/`) and which need to change.
- Anything the task's description got wrong or assumed incorrectly — note the
  discrepancy explicitly, you'll need it for Step 5.
- Blast radius: other callers/consumers of what you're about to change, and
  anything the task doesn't mention but the code says is also affected.

If personal data (CPF, phone, address, financial identifiers, etc.) is part of
what you're inspecting or would need to appear in the rewritten issue
description, apply the LGPD data-minimization rule: don't paste full
identifiers into the issue description — reference the field/column
generically, and flag it to the user if the task seems to require exporting or
repurposing such data.

## Step 4 — Ask the user

Before writing anything back to the issue, resolve every open question from Steps
1–3 with the user via `AskUserQuestion`. Typical triggers:
- The code supports the change in more than one plausible place/way.
- The task is ambiguous about scope (e.g. "fix X" could mean one endpoint or
  every caller).
- A discrepancy between the task's assumption and reality has more than one
  reasonable resolution (task might be stale, or code might be the actual bug).
- A decision has product/priority implications you can't infer from code.

Do not ask questions the code already answered in Steps 2–3.

## Step 5 — Write the plan

Synthesize a plan that is:
- **Objective** — states facts confirmed in Steps 2–3, not opinions.
- **Clear** — an implementer could start coding from it without re-doing your
  research.
- **Scoped** — explicit about what's in and out of scope, especially where the
  original task was vague and the user resolved it in Step 4.

Structure (adapt headers to fit, but keep this shape):

```
## Summary
<1-3 sentences: what changes and why, in plain terms>

## Affected files
- `<path>` (<layer>) — <what changes there and why>
- ...

## Technical plan
1. <concrete step, naming the real function/endpoint/table/field>
2. ...

## BDD coverage
- `test/integration/features/<domain>/<file>.feature` — <new / extend> — <expectations it locks>

## Out of scope
- <anything explicitly excluded, especially if the user drew this line in Step 4>

## Open decisions resolved
- <question> → <the user's answer> (only if Step 4 happened)
```

## Step 6 — Update the issue description

- Fetch the current body again (`gh issue view <n> --json body`) in case it
  changed since Step 1.
- Rewrite it: **keep** whatever context/background was accurate, **replace**
  claims that Steps 2–3 disproved, and **append** the plan from Step 5. Do not
  leave a stale assumption sitting next to its correction — remove it.
- Write the new body to a scratchpad file and save it with
  `gh issue edit <n> --repo brienze1/xenon --body-file <file>` (update the existing issue, not a new one).
- Do not touch labels, assignees or milestones unless the user asked you to (adding a missing `XEN-NNN` prefix
  to the title is the one exception, per the skill).

## Step 7 — Report back

Summarize to the user in a few sentences: what was confirmed, what was wrong
in the original description, what got resolved via questions, and confirm the
issue description was updated (issue link and `XEN-NNN` id). Keep it short — the description
itself is the durable artifact. Next step in the flow is `/spec-task`.
