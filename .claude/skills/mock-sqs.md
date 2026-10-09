---
name: Mock - SQS (AWS SQS)
description: Use when creating or referencing the AWS SQS mock for BDD tests.
version: 1.0.0
---

# Mock: SQS (AWS Simple Queue Service)

In-memory SQS mock that stores messages by queue URL with full attribute and deduplication ID tracking.

**Location:** `test/integration/mock/sqs.go`

> **Note:** Comments in code examples throughout this skill are for teaching purposes only. In actual code, follow the comments skill — no inline comments inside function bodies (see `comments.md`).

```go
package mock

import (
    "context"
    "sync"
    "github.com/aws/aws-sdk-go-v2/service/sqs"
    "github.com/google/uuid"
)

type SqsMessage struct {
    Body              string
    MessageAttributes map[string]string
    GroupId           *string
    DeduplicationId   *string
}

type SqsClient struct {
    mu       sync.RWMutex
    messages map[string][]SqsMessage  // key: queue URL
}

func NewSqsClient() *SqsClient {
    return &SqsClient{messages: make(map[string][]SqsMessage)}
}

func (s *SqsClient) SendMessage(ctx context.Context, params *sqs.SendMessageInput, optFns ...func(*sqs.Options)) (*sqs.SendMessageOutput, error) {
    s.mu.Lock()
    defer s.mu.Unlock()

    attrs := make(map[string]string)
    for k, v := range params.MessageAttributes {
        attrs[k] = *v.StringValue
    }

    msg := SqsMessage{
        Body:              *params.MessageBody,
        MessageAttributes: attrs,
        GroupId:           params.MessageGroupId,
        DeduplicationId:   params.MessageDeduplicationId,
    }

    s.messages[*params.QueueUrl] = append(s.messages[*params.QueueUrl], msg)
    msgId := uuid.NewString()
    return &sqs.SendMessageOutput{MessageId: &msgId}, nil
}

// Inspection methods
func (s *SqsClient) GetMessages(queueUrl string) []SqsMessage { ... }
func (s *SqsClient) GetMessageCount(queueUrl string) int { ... }
func (s *SqsClient) GetMessage(queueUrl string, index int) *SqsMessage { ... }

func (s *SqsClient) Reset() {
    s.mu.Lock()
    defer s.mu.Unlock()
    s.messages = make(map[string][]SqsMessage)
}
```

**Lifecycle:** The SQS mock is created once and injected once in `InitializeScenario`. The `Reset()` method clears all accumulated messages between BDD scenarios. The mock instance is never recreated or re-injected — only its internal state is cleared.

---

- [ ] Follows the comments skill — no inline comments, no unnecessary Go doc comments (see comments.md)
