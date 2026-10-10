package binding

import (
	"context"

	"github.com/brienze1/xenon/internal/integration/adapter"
	"github.com/brienze1/xenon/internal/integration/entrypoint/dto"
)

type EnvironmentService struct {
	list      adapter.EnvironmentListController
	setActive adapter.EnvironmentSetActiveController
}

func NewEnvironmentService(list adapter.EnvironmentListController, setActive adapter.EnvironmentSetActiveController) *EnvironmentService {
	return &EnvironmentService{list: list, setActive: setActive}
}

func (e *EnvironmentService) List(ctx context.Context, request dto.ListEnvironments) (*dto.EnvironmentList, error) {
	return e.list.Handle(ctx, request)
}

func (e *EnvironmentService) SetActive(ctx context.Context, request dto.SetActiveEnvironment) (*dto.ActiveEnvironment, error) {
	return e.setActive.Handle(ctx, request)
}
