---
name: Mock - Secrets Manager (AWS)
description: Use when creating or referencing the AWS Secrets Manager mock for BDD tests.
version: 1.0.0
---

# Mock: Secrets Manager (AWS)

In-memory mock for AWS Secrets Manager. Stores secrets as key-value maps and returns them as JSON.

**Location:** `test/integration/mock/secrets_manager.go`

> **Note:** Comments in code examples throughout this skill are for teaching purposes only. In actual code, follow the comments skill — no inline comments inside function bodies (see `comments.md`).

```go
package mock

import (
    "context"
    "encoding/json"
    "sync"
    "github.com/aws/aws-sdk-go-v2/service/secretsmanager"
    "github.com/google/uuid"
)

type SecretsManagerClient struct {
    mu      sync.RWMutex
    secrets map[string]map[string]any  // secretName -> key-value pairs
}

func NewSecretsManagerClient() *SecretsManagerClient {
    return &SecretsManagerClient{secrets: make(map[string]map[string]any)}
}

func (s *SecretsManagerClient) SetSecret(name string, values map[string]any) {
    s.mu.Lock()
    defer s.mu.Unlock()
    s.secrets[name] = values
}

func (s *SecretsManagerClient) GetSecretValue(ctx context.Context, params *secretsmanager.GetSecretValueInput, optFns ...func(*secretsmanager.Options)) (*secretsmanager.GetSecretValueOutput, error) {
    s.mu.RLock()
    defer s.mu.RUnlock()

    secret, exists := s.secrets[*params.SecretId]
    if !exists {
        return nil, fmt.Errorf("secret not found: %s", *params.SecretId)
    }

    jsonBytes, _ := json.Marshal(secret)
    secretString := string(jsonBytes)
    arn := fmt.Sprintf("arn:aws:secretsmanager:us-east-1:000000000000:secret:%s", *params.SecretId)
    versionId := uuid.NewString()

    return &secretsmanager.GetSecretValueOutput{
        ARN:          &arn,
        Name:         params.SecretId,
        SecretString: &secretString,
        VersionId:    &versionId,
    }, nil
}

func (s *SecretsManagerClient) Reset() {
    s.mu.Lock()
    defer s.mu.Unlock()
    s.secrets = make(map[string]map[string]any)
}
```

**Lifecycle:** The Secrets Manager mock is created once and injected once in `InitializeScenario`. The `Reset()` method clears all stored secrets between BDD scenarios. The mock instance is never recreated or re-injected — only its internal state is cleared.

---

- [ ] Follows the comments skill — no inline comments, no unnecessary Go doc comments (see comments.md)
