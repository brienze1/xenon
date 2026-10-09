---
name: Mock - S3 (AWS S3)
description: Use when creating or referencing the AWS S3 mock for BDD tests.
version: 1.0.0
---

# Mock: S3 (AWS Simple Storage Service)

In-memory S3 mock with bucket/key object storage, supporting PUT, GET, DELETE, LIST, and presigned URL operations. Thread-safe.

**Location:** `test/integration/mock/s3.go`

> **Note:** Comments in code examples throughout this skill are for teaching purposes only. In actual code, follow the comments skill — no inline comments inside function bodies (see `comments.md`).

```go
package mock

import (
    "bytes"
    "context"
    "fmt"
    "io"
    "strings"
    "sync"

    "github.com/aws/aws-sdk-go-v2/service/s3"
)

type S3Object struct {
    Body        []byte
    ContentType string
}

type S3Client struct {
    mu      sync.RWMutex
    objects map[string]map[string]S3Object  // bucket -> key -> object
}

func NewS3Client() *S3Client {
    return &S3Client{objects: make(map[string]map[string]S3Object)}
}

func (s *S3Client) PutObject(ctx context.Context, params *s3.PutObjectInput, optFns ...func(*s3.Options)) (*s3.PutObjectOutput, error) {
    s.mu.Lock()
    defer s.mu.Unlock()

    bucket := *params.Bucket
    if s.objects[bucket] == nil {
        s.objects[bucket] = make(map[string]S3Object)
    }

    body, _ := io.ReadAll(params.Body)
    contentType := ""
    if params.ContentType != nil { contentType = *params.ContentType }

    s.objects[bucket][*params.Key] = S3Object{Body: body, ContentType: contentType}
    return &s3.PutObjectOutput{}, nil
}

func (s *S3Client) GetObject(ctx context.Context, params *s3.GetObjectInput, optFns ...func(*s3.Options)) (*s3.GetObjectOutput, error) {
    s.mu.RLock()
    defer s.mu.RUnlock()

    obj, exists := s.objects[*params.Bucket][*params.Key]
    if !exists {
        return nil, fmt.Errorf("NoSuchKey: %s", *params.Key)
    }
    return &s3.GetObjectOutput{
        Body:        io.NopCloser(bytes.NewReader(obj.Body)),
        ContentType: &obj.ContentType,
    }, nil
}

func (s *S3Client) DeleteObjectsByPrefix(ctx context.Context, bucket, prefix string) error { ... }
func (s *S3Client) PresignGetObject(ctx context.Context, bucket, key string) (string, error) {
    return fmt.Sprintf("https://%s.s3.amazonaws.com/%s?presigned=true", bucket, key), nil
}

// Inspection methods
func (s *S3Client) GetObjectCount(bucket string) int { ... }
func (s *S3Client) GetAllKeys(bucket string) []string { ... }

func (s *S3Client) Reset() {
    s.mu.Lock()
    defer s.mu.Unlock()
    s.objects = make(map[string]map[string]S3Object)
}
```

**Lifecycle:** The S3 mock is created once and injected once in `InitializeScenario`. The `Reset()` method clears all stored objects between BDD scenarios. The mock instance is never recreated or re-injected — only its internal state is cleared.

---

- [ ] Follows the comments skill — no inline comments, no unnecessary Go doc comments (see comments.md)
