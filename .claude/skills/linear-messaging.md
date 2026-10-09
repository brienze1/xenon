# Linear Comment Formatting Guide

When posting comments to Linear issues, follow this structured format for consistency and readability.

## Comment Structure

Every comment MUST follow this pattern:

```
🤖 **{job-name}**: {emoji} {VERDICT/STATUS}

{body}
```

## Emoji Conventions

- ✅ Success / Approved / Passed
- ❌ Failed / Rejected / Error
- 🔧 Fix applied / Refactored
- 🔍 Diagnosed / Analyzed
- 📦 Committed / PR created
- 🧪 QA result
- ⏳ In progress / Waiting

## Body Formatting

Use markdown sections with clear headers:

```markdown
**Branch:** `plt-1234`

**Summary:** One-line description of what happened.

**Details:**
- Item 1
- Item 2

**Files changed:**
- `internal/path/to/file.go`
- `internal/other/file.go`
```

## Examples

### Implementation complete:
```
🤖 **implement-task**: ✅ IMPLEMENTED

**Branch:** `plt-1165`

**Summary:** Added new endpoint for creator payout processing.

**Files changed:**
- `internal/domain/entity/payout.go`
- `internal/application/usecase/process_payout.go`
- `internal/integration/entrypoint/controller/payout.go`
- `test/features/payout/process_payout.feature`
```

### Review rejected:
```
🤖 **review**: ❌ REJECTED

**Issues:**
- Missing error handling in `ProcessPayout` usecase (line 45)
- SQL injection risk in raw query at `persistence/payout.go:23`
- BDD step missing for error case
```

### CI diagnosis:
```
🤖 **diagnose-fix**: 🔧 FIXED

**Failed step:** tests
**Fix applied:** Added missing mock for PayoutRepository in process_payout_test.go
```

### PR created:
```
🤖 **commit-and-pr**: 📦 PR CREATED

**PR:** https://github.com/brienze1/xenon/pull/123 ✅ CI passed
```

### QA result:
```
🤖 **qa-hml**: 🧪 QA PASSED

**Verdict:** All endpoints responding correctly. Payout flow tested end-to-end.
```

## Rules

- Always start with 🤖 and the job name in bold
- Always include the branch involved
- Keep it concise — Linear comments should be scannable
- Use backticks for code, file paths, and branch names
- Use bullet lists for multiple items
- Include the task identifier when referencing other tasks
- Linear CLI: `linear issue comment add <identifier> --body "<comment>"`
