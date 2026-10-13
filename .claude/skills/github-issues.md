# GitHub Issues — the task tracker

Xenon's tasks live in **GitHub issues** on `brienze1/xenon` (there is no Linear workspace). The task commands
(`/research-task`, `/spec-task`, `/create-tests`, `/implement-task`) read and update issues with the `gh` CLI. Never
use Linear tools for this repo.

## The argument

A command's argument names one issue, in any of these forms:

- `12` or `#12` — the issue number
- `https://github.com/brienze1/xenon/issues/12` — its URL
- `XEN-008` — a task id; find its issue with
  `gh issue list --repo brienze1/xenon --state all --search "XEN-008 in:title" --json number,title`

If none of these resolves to exactly one issue, stop and ask.

## Task id, branch and spec folder

Every task has an id `XEN-NNN` (three digits, zero-padded), continuing the stage numbering in `specs/`
(`XEN-001` … `XEN-007` came before issues were used). Issue numbers are **not** task ids: GitHub shares one counter
between issues and pull requests, so `#2` must never become `XEN-002`.

- **The id lives in the issue title**, as a prefix: `XEN-008 Collection runner`.
- **No id in the title yet** (a fresh issue): take the next free number — one more than the highest `XEN-NNN` in
  `specs/` and in issue titles
  (`gh issue list --repo brienze1/xenon --state all --search "XEN- in:title" --json title --limit 500`) — and
  prefix the title with it: `gh issue edit <n> --title "XEN-008 <title>"`. Tell the user which id was assigned.
- **Branch**: the lowercased id, `xen-008` — one branch per task, shared by all four commands.
- **Spec folder**: `specs/XEN-008-<short-slug>/`.

## Reading an issue

```bash
gh issue view <n> --repo brienze1/xenon --json number,title,body,labels,state,url,comments
```

Comments come after the body in time: a later comment that clarifies or changes scope overrides the body.

## Updating an issue

- **Description**: write the full new body to a file in the scratchpad, then
  `gh issue edit <n> --repo brienze1/xenon --body-file <file>`. Fetch the body again right before editing, in case it
  changed.
- **Comment**: `gh issue comment <n> --repo brienze1/xenon --body-file <file>` (format below).
- **Status**: labels stand in for Linear's states. `/implement-task` adds `in progress` when it starts:
  `gh label create "in progress" --repo brienze1/xenon --color FBCA04 --description "Being implemented" 2>/dev/null || true`
  then `gh issue edit <n> --repo brienze1/xenon --add-label "in progress"`. Issues are not closed by the commands;
  an issue closes when its work merges to `main`.
- Do not change the title (other than adding the missing id), labels, assignees or milestones unless the user asks.

## Comment format

Every comment a command posts follows this pattern:

```
🤖 **{command}**: {emoji} {VERDICT/STATUS}

{body}
```

Emoji:

- ✅ Success / approved / passed
- ❌ Failed / rejected / error
- 🔧 Fix applied / refactored
- 🔍 Diagnosed / analysed
- 📦 Committed / pushed
- 🧪 Test result
- ⏳ In progress / waiting

Body:

```markdown
**Branch:** `xen-008`

**Summary:** One line on what happened.

**Details:**
- Item 1
- Item 2

**Files changed:**
- `internal/path/to/file.go`
```

### Examples

Spec linked:

```
🤖 **spec-task**: ✅ SPEC READY

**Branch:** `xen-008`

**Spec:** https://github.com/brienze1/xenon/tree/xen-008/specs/XEN-008-collection-runner
```

Implementation complete:

```
🤖 **implement-task**: ✅ IMPLEMENTED

**Branch:** `xen-008`

**Summary:** Collections run their requests in order with per-request results.

**Files changed:**
- `internal/application/usecase/run_collection.go`
- `test/integration/features/runner/run_collection.feature`
```

### Rules

- Start with 🤖 and the command name in bold.
- Always name the branch.
- Keep it short and scannable.
- Backticks for code, file paths and branch names; bullet lists for several items.
- Refer to other tasks by their `XEN-NNN` id and issue link.
- Never paste secrets, tokens or variable values into an issue.
