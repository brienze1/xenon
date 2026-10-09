---
name: Mock - DynamoDB (AWS DynamoDB)
description: Use when creating or referencing the AWS DynamoDB mock for BDD tests.
version: 1.0.0
---

# Mock: DynamoDB (AWS DynamoDB)

In-memory DynamoDB mock with table management, CRUD operations, query/scan with filtering, and batch operations.

**Location:** `test/integration/mock/dynamodb.go`

> **Note:** Comments in code examples throughout this skill are for teaching purposes only. In actual code, follow the comments skill — no inline comments inside function bodies (see `comments.md`).

```go
package mock

import (
    "context"
    "sync"
    "github.com/aws/aws-sdk-go-v2/service/dynamodb"
    "github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
)

type DynamoDbTable struct {
    PartitionKey string
    SortKey      string
    Items        []map[string]types.AttributeValue
}

type DynamoDbClient struct {
    mu     sync.RWMutex
    tables map[string]*DynamoDbTable
}

func NewDynamoDbClient() *DynamoDbClient {
    return &DynamoDbClient{tables: make(map[string]*DynamoDbTable)}
}

func (d *DynamoDbClient) AddTable(name, partitionKey string, sortKey ...string) {
    sk := ""
    if len(sortKey) > 0 { sk = sortKey[0] }
    d.tables[name] = &DynamoDbTable{PartitionKey: partitionKey, SortKey: sk}
}

func (d *DynamoDbClient) AddItem(tableName string, item map[string]types.AttributeValue) { ... }
func (d *DynamoDbClient) GetItem(ctx context.Context, params *dynamodb.GetItemInput, ...) (*dynamodb.GetItemOutput, error) { ... }
func (d *DynamoDbClient) PutItem(ctx context.Context, params *dynamodb.PutItemInput, ...) (*dynamodb.PutItemOutput, error) { ... }
func (d *DynamoDbClient) DeleteItem(ctx context.Context, params *dynamodb.DeleteItemInput, ...) (*dynamodb.DeleteItemOutput, error) { ... }
func (d *DynamoDbClient) Query(ctx context.Context, params *dynamodb.QueryInput, ...) (*dynamodb.QueryOutput, error) { ... }
func (d *DynamoDbClient) Scan(ctx context.Context, params *dynamodb.ScanInput, ...) (*dynamodb.ScanOutput, error) { ... }
func (d *DynamoDbClient) BatchGetItem(ctx context.Context, params *dynamodb.BatchGetItemInput, ...) (*dynamodb.BatchGetItemOutput, error) { ... }

// Inspection methods
func (d *DynamoDbClient) GetAllItems(tableName string) []map[string]types.AttributeValue { ... }
func (d *DynamoDbClient) GetItemCount(tableName string) int { ... }

func (d *DynamoDbClient) Reset() {
    d.mu.Lock()
    defer d.mu.Unlock()
    for _, table := range d.tables {
        table.Items = nil
    }
}
```

**Lifecycle:** The DynamoDB mock is created once and injected once in `InitializeScenario`. The `Reset()` method clears all items from all tables between BDD scenarios (preserving table definitions). The mock instance is never recreated or re-injected — only its internal state is cleared.

---

- [ ] Follows the comments skill — no inline comments, no unnecessary Go doc comments (see comments.md)
