---
name: Mock - Rekognition (AWS)
description: Use when creating or referencing the AWS Rekognition mock for BDD tests.
version: 1.0.0
---

# Mock: Rekognition (AWS Facial Recognition)

Configurable Rekognition mock that simulates face operations with controllable approval/rejection and match results.

**Location:** `test/integration/mock/rekognition.go`

```go
package mock

import (
    "context"
    "sync"
    "github.com/aws/aws-sdk-go-v2/service/rekognition"
    "github.com/aws/aws-sdk-go-v2/service/rekognition/types"
    "github.com/google/uuid"
)

type RekognitionClient struct {
    mu       sync.RWMutex
    approve  bool
    matches  []types.UserMatch
}

func NewRekognitionClient() *RekognitionClient {
    return &RekognitionClient{}
}

func (r *RekognitionClient) SetResponse(approve bool, matches []types.UserMatch) {
    r.mu.Lock()
    defer r.mu.Unlock()
    r.approve = approve
    r.matches = matches
}

func (r *RekognitionClient) CreateCollection(ctx context.Context, params *rekognition.CreateCollectionInput, ...) (*rekognition.CreateCollectionOutput, error) {
    return &rekognition.CreateCollectionOutput{}, nil
}

func (r *RekognitionClient) CreateUser(ctx context.Context, params *rekognition.CreateUserInput, ...) (*rekognition.CreateUserOutput, error) {
    return &rekognition.CreateUserOutput{}, nil
}

func (r *RekognitionClient) IndexFaces(ctx context.Context, params *rekognition.IndexFacesInput, ...) (*rekognition.IndexFacesOutput, error) {
    faceId := uuid.NewString()
    return &rekognition.IndexFacesOutput{
        FaceRecords: []types.FaceRecord{{Face: &types.Face{FaceId: &faceId}}},
    }, nil
}

func (r *RekognitionClient) SearchUsersByImage(ctx context.Context, params *rekognition.SearchUsersByImageInput, ...) (*rekognition.SearchUsersByImageOutput, error) {
    r.mu.RLock()
    defer r.mu.RUnlock()
    return &rekognition.SearchUsersByImageOutput{UserMatches: r.matches}, nil
}

func (r *RekognitionClient) Reset() {
    r.mu.Lock()
    defer r.mu.Unlock()
    r.approve = false
    r.matches = nil
}
```

**Lifecycle:** The Rekognition mock is created once and injected once in `InitializeScenario`. The `Reset()` method clears approval state and match results between BDD scenarios. The mock instance is never recreated or re-injected — only its internal state is cleared.

---

- [ ] Follows the comments skill — no inline comments, no unnecessary Go doc comments (see comments.md)
