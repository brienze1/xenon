package binding

import (
	"context"

	"github.com/brienze1/xenon/internal/integration/adapter"
	"github.com/brienze1/xenon/internal/integration/entrypoint/dto"
)

type EnvironmentService struct {
	list      adapter.EnvironmentListController
	setActive adapter.EnvironmentSetActiveController
	manage    adapter.EnvironmentManageController
}

func NewEnvironmentService(list adapter.EnvironmentListController, setActive adapter.EnvironmentSetActiveController, manage adapter.EnvironmentManageController) *EnvironmentService {
	return &EnvironmentService{list: list, setActive: setActive, manage: manage}
}

func (e *EnvironmentService) List(ctx context.Context, request dto.ListEnvironments) (*dto.EnvironmentList, error) {
	return e.list.Handle(ctx, request)
}

func (e *EnvironmentService) SetActive(ctx context.Context, request dto.SetActiveEnvironment) (*dto.ActiveEnvironment, error) {
	return e.setActive.Handle(ctx, request)
}

func (e *EnvironmentService) Get(ctx context.Context, request dto.EnvironmentReference) (*dto.Environment, error) {
	return e.manage.Get(ctx, request)
}

func (e *EnvironmentService) Create(ctx context.Context, request dto.CreateEnvironment) (*dto.Environment, error) {
	return e.manage.Create(ctx, request)
}

func (e *EnvironmentService) Update(ctx context.Context, request dto.UpdateEnvironment) (*dto.Environment, error) {
	return e.manage.Update(ctx, request)
}

func (e *EnvironmentService) Duplicate(ctx context.Context, request dto.EnvironmentReference) (*dto.Environment, error) {
	return e.manage.Duplicate(ctx, request)
}

func (e *EnvironmentService) Delete(ctx context.Context, request dto.EnvironmentReference) (*dto.ActiveEnvironment, error) {
	return e.manage.Delete(ctx, request)
}

func (e *EnvironmentService) Reorder(ctx context.Context, request dto.ReorderEnvironment) (*dto.EnvironmentList, error) {
	return e.manage.Reorder(ctx, request)
}
