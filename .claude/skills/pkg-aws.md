---
name: Package - aws (AWS Services)
description: This skill should be used when the user asks to "create AWS integration", "add S3 operations", "add DynamoDB operations", "add SQS/SNS publishing", "create secrets manager", "add Rekognition", "add SES email", or mentions anything about the pkg/aws/ package.
version: 1.0.0
---

# Package: aws — AWS Services Integration

The `aws` package provides high-level wrappers for AWS services (S3, DynamoDB, SQS, SNS, SES, Rekognition, Secrets Manager). Each service exposes an interface, accepts AWS config via a shared factory, and includes DataDog tracing support.

**Location:** `pkg/aws/`

---

## Architecture

```
pkg/aws/
├── aws.go              ← Core AWS config factory with region caching + tracing
├── s3.go               ← S3 operations (Put, Get, Delete, Presign)
├── dynamodb.go         ← DynamoDB operations (CRUD, Query, BatchGet)
├── sqs.go              ← SQS message sending
├── sns.go              ← SNS message publishing
├── ses.go              ← SES templated email sending
├── rekognition.go      ← Rekognition face operations
├── secrets_manager.go  ← Secrets Manager with Redis caching
└── utils.go            ← Utility functions for parsing messages/streams
```

---

## Core AWS Config

The `aws.go` file provides a region-based config factory with caching:

> **Note:** Comments in code examples throughout this skill are for teaching purposes only. In actual code, follow the comments skill — no inline comments inside function bodies (see `comments.md`).

```go
// Each region's config is cached (different services may use different regions)
func GetAwsConfig(region string) aws.Config { ... }

// Includes DataDog tracing when DD_TRACE_ENABLED=true
```

---

## Service Interfaces

Each service exposes an interface with an unexported struct and factory function:

### S3
```go
type S3HelperAdapter interface {
    PutObject(ctx context.Context, bucket, key, contentType string, body io.Reader) error
    GetObject(ctx context.Context, bucket, key string) ([]byte, error)
    DeleteObjectsByPrefix(ctx context.Context, bucket, prefix string) error
    PresignGetObject(ctx context.Context, bucket, key string, expiresIn time.Duration) (string, error)
}
```

### DynamoDB
```go
type DynamoDbHelperAdapter interface {
    BatchGetItem(ctx, tableName string, keys []map[string]types.AttributeValue, dest any) error
    GetItem(ctx, tableName, partitionKey, partitionValue string, dest any) error
    GetItemWithRange(ctx, tableName, pk, pkVal, sk, skVal string, dest any) error
    GetItems(ctx, tableName, pk, pkVal string, dest any) error
    PutItem(ctx, tableName string, item any) error
    DeleteItem(ctx, tableName, pk, pkVal string) error
    Query(ctx, tableName, indexName, condition string, values map[string]types.AttributeValue, dest any) error
}
```

### SQS
```go
type SqsHelperAdapter interface {
    SendMessage(ctx context.Context, queueUrl string, message any) error
}
```

### SNS
```go
type SnsHelperAdapter interface {
    Publish(ctx context.Context, topicArn string, message any) error
}
```

### SES
```go
type SesHelperAdapter interface {
    SendTemplatedEmail(ctx context.Context, from, to, template string, data any) error
}
```

### Rekognition
```go
type RekognitionHelperAdapter interface {
    CreateCollection(ctx context.Context, collectionId string) error
    CreateUser(ctx context.Context, collectionId, userId string) error
    AssociateFace(ctx context.Context, collectionId, userId string, image []byte) error
    SearchUsersByImage(ctx context.Context, collectionId string, image []byte, threshold float32) ([]types.UserMatch, error)
    DisassociateFace(ctx context.Context, collectionId, faceId string) error
}
```

### Secrets Manager
```go
type SecretsManagerHelperAdapter interface {
    GetSecret(ctx context.Context, secretName string, dest any) error
}
// Includes Redis caching layer for secrets
```

---

## Key Rules

- **No internal imports** — standalone package
- **Logger via local interface** — each service accepts a `loggerAdapter`
- **Region-based config caching** — AWS configs cached per region
- **DataDog tracing** — optional, enabled via `DD_TRACE_ENABLED` env var
- **Exported interfaces, unexported structs**
- **Copy-ready** — can be dropped into any Go project with AWS SDK v2

---

- [ ] Follows the comments skill — no inline comments, no unnecessary Go doc comments (see comments.md)
