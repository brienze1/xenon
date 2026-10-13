package binding

import (
	"context"

	"github.com/brienze1/xenon/internal/integration/adapter"
	"github.com/brienze1/xenon/internal/integration/entrypoint/dto"
)

type ExampleService struct {
	manage adapter.ExampleManageController
}

func NewExampleService(manage adapter.ExampleManageController) *ExampleService {
	return &ExampleService{manage: manage}
}

func (e *ExampleService) Get(ctx context.Context, request dto.ExampleReference) (*dto.Example, error) {
	return e.manage.Get(ctx, request)
}

func (e *ExampleService) Create(ctx context.Context, request dto.CreateExample) (*dto.Example, error) {
	return e.manage.Create(ctx, request)
}

func (e *ExampleService) CreateFromResponse(ctx context.Context, request dto.CreateExampleFromResponse) (*dto.Example, error) {
	return e.manage.CreateFromResponse(ctx, request)
}

func (e *ExampleService) Save(ctx context.Context, request dto.SaveExample) (*dto.Example, error) {
	return e.manage.Save(ctx, request)
}

func (e *ExampleService) Rename(ctx context.Context, request dto.RenameExample) (*dto.Example, error) {
	return e.manage.Rename(ctx, request)
}

func (e *ExampleService) Duplicate(ctx context.Context, request dto.ExampleReference) (*dto.Example, error) {
	return e.manage.Duplicate(ctx, request)
}

func (e *ExampleService) Delete(ctx context.Context, request dto.ExampleReference) error {
	return e.manage.Delete(ctx, request)
}
