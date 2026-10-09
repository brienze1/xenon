---
name: Mock - SES (AWS SES)
description: Use when creating or referencing the AWS SES email mock for BDD tests.
version: 1.0.0
---

# Mock: SES (AWS Simple Email Service)

In-memory SES mock that tracks sent emails with template data and allows response customization.

**Location:** `test/integration/mock/ses.go`

> **Note:** Comments in code examples throughout this skill are for teaching purposes only. In actual code, follow the comments skill — no inline comments inside function bodies (see `comments.md`).

```go
package mock

import (
    "context"
    "sync"
    "github.com/aws/aws-sdk-go-v2/service/sesv2"
    "github.com/google/uuid"
)

type SesEmail struct {
    From         string
    To           string
    TemplateName string
    TemplateData string
}

type SesClient struct {
    mu     sync.RWMutex
    emails []SesEmail
    err    error  // Optional: simulate failures
}

func NewSesClient() *SesClient {
    return &SesClient{emails: make([]SesEmail, 0)}
}

func (s *SesClient) SendEmail(ctx context.Context, params *sesv2.SendEmailInput, optFns ...func(*sesv2.Options)) (*sesv2.SendEmailOutput, error) {
    s.mu.Lock()
    defer s.mu.Unlock()

    if s.err != nil {
        return nil, s.err
    }

    email := SesEmail{
        From:         *params.FromEmailAddress,
        To:           params.Destination.ToAddresses[0],
        TemplateName: *params.Content.Template.TemplateName,
        TemplateData: *params.Content.Template.TemplateData,
    }
    s.emails = append(s.emails, email)

    msgId := uuid.NewString()
    return &sesv2.SendEmailOutput{MessageId: &msgId}, nil
}

// Inspection methods
func (s *SesClient) GetEmails() []SesEmail { ... }
func (s *SesClient) GetEmailCount() int { ... }

func (s *SesClient) SetError(err error) { s.err = err }

func (s *SesClient) Reset() {
    s.mu.Lock()
    defer s.mu.Unlock()
    s.emails = make([]SesEmail, 0)
    s.err = nil
}
```

**Lifecycle:** The SES mock is created once and injected once in `InitializeScenario`. The `Reset()` method clears all sent emails and error state between BDD scenarios. The mock instance is never recreated or re-injected — only its internal state is cleared.

---

- [ ] Follows the comments skill — no inline comments, no unnecessary Go doc comments (see comments.md)
