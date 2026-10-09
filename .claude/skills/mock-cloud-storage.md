---
name: Mock - Cloud Storage
description: Use when creating or referencing the cloud storage mock for BDD tests.
version: 1.0.0
---

# Mock: Cloud Storage

Abstracts S3/GCS as an in-memory cloud storage mock. Implements the `CloudStorage` adapter interface.

**Location:** `test/integration/mock/cloud_storage.go`

> **Note:** Comments in code examples throughout this skill are for teaching purposes only. In actual code, follow the comments skill — no inline comments inside function bodies (see `comments.md`).

```go
package mock

import (
    "bytes"
    "context"
    "fmt"
    "io"
    "sync"
)

type CloudStorageClient struct {
    mu      sync.RWMutex
    objects map[string]map[string][]byte  // bucket -> key -> data
}

func NewCloudStorageClient() *CloudStorageClient {
    return &CloudStorageClient{
        objects: make(map[string]map[string][]byte),
    }
}

func (c *CloudStorageClient) Put(ctx context.Context, bucket, key, contentType string, body io.Reader) error {
    c.mu.Lock()
    defer c.mu.Unlock()
    if c.objects[bucket] == nil {
        c.objects[bucket] = make(map[string][]byte)
    }
    data, _ := io.ReadAll(body)
    c.objects[bucket][key] = data
    return nil
}

func (c *CloudStorageClient) Get(ctx context.Context, bucket, key string) ([]byte, error) {
    c.mu.RLock()
    defer c.mu.RUnlock()
    data, exists := c.objects[bucket][key]
    if !exists {
        return nil, fmt.Errorf("object not found: %s/%s", bucket, key)
    }
    return data, nil
}

func (c *CloudStorageClient) Delete(ctx context.Context, bucket, prefix string) error { ... }

// Inspection methods
func (c *CloudStorageClient) GetObjectCount(bucket string) int { ... }

func (c *CloudStorageClient) Reset() {
    c.mu.Lock()
    defer c.mu.Unlock()
    c.objects = make(map[string]map[string][]byte)
}
```

**Lifecycle:** The cloud storage mock is created once and injected once via `storage.SetCloudStorageInstance()` in `InitializeScenario`. The `Reset()` method clears all stored objects between BDD scenarios. The mock instance is never recreated or re-injected — only its internal state is cleared.

---

- [ ] Follows the comments skill — no inline comments, no unnecessary Go doc comments (see comments.md)
