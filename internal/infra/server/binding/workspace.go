package binding

import (
	"context"

	"github.com/brienze1/xenon/internal/infra/dependency/controller"
	"github.com/brienze1/xenon/internal/integration/adapter"
	"github.com/brienze1/xenon/internal/integration/entrypoint/dto"
)

type WorkspaceService struct {
	list     adapter.WorkspaceListController
	create   adapter.WorkspaceCreateController
	activate adapter.WorkspaceActivateController
}

func NewWorkspaceService() *WorkspaceService {
	return &WorkspaceService{
		list:     controller.GetWorkspaceList(),
		create:   controller.GetWorkspaceCreate(),
		activate: controller.GetWorkspaceActivate(),
	}
}

func (w *WorkspaceService) List(ctx context.Context) (*dto.WorkspaceList, error) {
	return w.list.Handle(ctx)
}

func (w *WorkspaceService) Create(ctx context.Context, request dto.CreateWorkspace) (*dto.WorkspaceSummary, error) {
	return w.create.Handle(ctx, request)
}

func (w *WorkspaceService) Activate(ctx context.Context, request dto.ActivateWorkspace) (*dto.WorkspaceSnapshot, error) {
	return w.activate.Handle(ctx, request)
}
