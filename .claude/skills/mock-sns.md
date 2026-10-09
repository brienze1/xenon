---
name: Mock - SNS (AWS SNS)
description: Use when creating or referencing the AWS SNS mock for BDD tests.
version: 1.0.0
---

# Mock: SNS (AWS Simple Notification Service)

In-memory SNS mock that tracks published messages per topic ARN.

**Location:** `test/integration/mock/sns.go`

> **Note:** Comments in code examples throughout this skill are for teaching purposes only. In actual code, follow the comments skill — no inline comments inside function bodies (see `comments.md`).

```go
package mock

import (
    "context"
    "sync"
    "github.com/aws/aws-sdk-go-v2/service/sns"
    "github.com/google/uuid"
)

type SnsMessage struct {
    Body              string
    MessageAttributes map[string]string
}

type SnsClient struct {
    mu       sync.RWMutex
    messages map[string][]SnsMessage  // key: topic ARN
}

func NewSnsClient() *SnsClient {
    return &SnsClient{messages: make(map[string][]SnsMessage)}
}

func (s *SnsClient) Publish(ctx context.Context, params *sns.PublishInput, optFns ...func(*sns.Options)) (*sns.PublishOutput, error) {
    s.mu.Lock()
    defer s.mu.Unlock()

    attrs := make(map[string]string)
    for k, v := range params.MessageAttributes {
        attrs[k] = *v.StringValue
    }

    msg := SnsMessage{Body: *params.Message, MessageAttributes: attrs}
    s.messages[*params.TopicArn] = append(s.messages[*params.TopicArn], msg)

    msgId := uuid.NewString()
    return &sns.PublishOutput{MessageId: &msgId}, nil
}

// Inspection methods
func (s *SnsClient) GetMessages(topicArn string) []SnsMessage { ... }
func (s *SnsClient) GetMessageCount(topicArn string) int { ... }

func (s *SnsClient) Reset() {
    s.mu.Lock()
    defer s.mu.Unlock()
    s.messages = make(map[string][]SnsMessage)
}
```

**Lifecycle:** The SNS mock is created once and injected once in `InitializeScenario`. The `Reset()` method clears all accumulated messages between BDD scenarios. The mock instance is never recreated or re-injected — only its internal state is cleared.

---

- [ ] Follows the comments skill — no inline comments, no unnecessary Go doc comments (see comments.md)
