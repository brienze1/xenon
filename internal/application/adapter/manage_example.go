package adapter

import (
	"context"

	"github.com/brienze1/xenon/internal/domain/entity"
)

type ManageExampleService interface {
	Get(ctx context.Context, localId, requestId, exampleId string) (*entity.ExampleDetails, error)
	Create(ctx context.Context, localId, requestId, name string) (*entity.ExampleDetails, error)
	CreateFromResponse(ctx context.Context, localId, requestId, tabId, name string) (*entity.ExampleDetails, error)
	Save(ctx context.Context, localId, requestId string, example entity.SavedExample, keepBody bool, baseRevision string, force bool) (*entity.ExampleDetails, error)
	Rename(ctx context.Context, localId, requestId, exampleId, name string) (*entity.ExampleDetails, error)
	Duplicate(ctx context.Context, localId, requestId, exampleId string) (*entity.ExampleDetails, error)
	Delete(ctx context.Context, localId, requestId, exampleId string) error
}
