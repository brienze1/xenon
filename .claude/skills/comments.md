---
name: Comments (Code Documentation)
description: This skill should be used when the user asks about "comments", "documentation", "godoc", "code comments", or when reviewing code that contains inline comments, block comments, or documentation comments.
version: 1.0.0
---

# Comments — Golang Clean Architecture

Comments are strictly controlled. Well-written code should be self-documenting. Comments inside functions are forbidden. Go doc comments are allowed only when truly necessary, and must be minimal.

---

## Core Rules

### Rule 1: No Comments Inside Functions (FORBIDDEN)

Comments inside function bodies are FORBIDDEN — no inline comments, no block comments, no TODO comments, no explanatory comments. If the code needs a comment to be understood, the code needs to be rewritten to be clearer.

```go
// BAD — comments inside function body
func (o *Order) IsReadyToFulfill() bool {
    // Check if the order is active and has items
    if o.Status != "active" {
        return false // not active, can't fulfill
    }
    // Make sure total is positive
    return o.Total > 0
}

// GOOD — self-documenting, no comments needed
func (o *Order) IsReadyToFulfill() bool {
    return o.Status == "active" && o.Total > 0
}
```

```go
// BAD — TODO comments
func (s *processOrder) Process(ctx context.Context, ...) error {
    // TODO: add caching here
    // FIXME: this might break with nil values
    // HACK: temporary workaround for the timezone issue
}

// GOOD — no TODOs in code (use issue tracker instead)
```

---

### Rule 2: Go Doc Comments — Only When Truly Necessary

Go doc comments (the `// FunctionName ...` pattern above exported functions) are allowed but only when the function name alone does not convey enough information. Most well-named functions do NOT need a doc comment.

**When to add a Go doc comment:**
- The function has non-obvious behavior (e.g., panics on invalid input)
- The function has important constraints callers must know
- The function is a public API in a `pkg/` package

**When NOT to add a Go doc comment:**
- The function name already explains what it does (`IsActive`, `ToEntity`, `FindById`)
- The function is an interface method with a clear name
- The function is unexported

```go
// BAD — unnecessary doc comment (name says it all)
// IsActive checks if the campaign is active.
func (c *Campaign) IsActive() bool {
    return campaignstatus.Participating.Equals(c.StatusId)
}

// BAD — restating the obvious
// ToEntity converts the model to a domain entity.
func (m *User) ToEntity() entity.User {
    return entity.User{Id: m.ID, Name: m.Name}
}

// BAD — describing downstream behavior
// FindById finds a user by their ID in the database and returns a pointer to the
// user entity. If the user is not found, it returns a NotFoundError. The error
// is then handled by the error handler which formats it as a 404 response.
func (p *userPersistence) FindById(ctx context.Context, userId int64) (*entity.User, error) {

// GOOD — no comment needed, name is clear
func (c *Campaign) IsActive() bool {
    return campaignstatus.Participating.Equals(c.StatusId)
}

// GOOD — no comment needed, name is clear
func (m *User) ToEntity() entity.User {
    return entity.User{Id: m.ID, Name: m.Name}
}

// GOOD — doc comment adds value (non-obvious constraint)
// ShouldAlert panics if weights don't sum to 1 or scores are out of [0,1] range.
func ShouldAlert(channel enums.Channel, mediaType enums.MediaType, minScore, maxScore, visualWeight, audioWeight float64) bool {
```

---

### Rule 3: Doc Comments Must Be Short

When a doc comment IS needed, it must be one line — two lines maximum. It must describe only the function itself, never downstream behavior.

```go
// GOOD — one line, specific to this function
// CalculateScore panics if weights don't sum to 1.
func CalculateScore(visual, audio, vWeight, aWeight float64) float64 {

// BAD — too long, describes downstream behavior
// CalculateScore calculates a weighted similarity score between two media files.
// The score is used by the service to determine if an alert should be sent.
// The result is stored in the database and later retrieved by the dashboard.
// If weights don't sum to 1, it panics.
func CalculateScore(visual, audio, vWeight, aWeight float64) float64 {
```

---

### Rule 4: No Comment Noise

Never add comments that are just noise:

```go
// BAD — package declaration comments restating the package name
// Package entity contains domain entities.
package entity

// BAD — section dividers
// ==========================================
// Private methods
// ==========================================

// BAD — closing brace comments
} // end of function

// BAD — commented-out code (delete it, git has history)
// func oldFunction() {
//     // ...
// }
```

---

## Quick Reference

| Situation | Comment? | Example |
|---|---|---|
| Inside function body | NEVER | No inline, block, or TODO comments |
| Clear function name | NO | `IsActive()`, `ToEntity()`, `FindById()` |
| Non-obvious behavior | YES (1 line) | `// Panics if weights don't sum to 1.` |
| Important constraint | YES (1 line) | `// Returns nil if not found (no error).` |
| Public pkg API | YES (1 line) | `// HttpClient creates a client with logging and tracing.` |
| Unexported function | NO | Private functions don't need doc comments |
| Commented-out code | DELETE IT | Git has the history |
