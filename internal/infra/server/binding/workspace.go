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
	open     adapter.WorkspaceOpenController
	rename   adapter.WorkspaceRenameController
	remove   adapter.WorkspaceRemoveController
	locate   adapter.WorkspaceLocateController
}

func NewWorkspaceService() *WorkspaceService {
	return &WorkspaceService{
		list:     controller.GetWorkspaceList(),
		create:   controller.GetWorkspaceCreate(),
		activate: controller.GetWorkspaceActivate(),
		open:     controller.GetWorkspaceOpen(),
		rename:   controller.GetWorkspaceRename(),
		remove:   controller.GetWorkspaceRemove(),
		locate:   controller.GetWorkspaceLocate(),
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

func (w *WorkspaceService) Open(ctx context.Context, request dto.OpenWorkspace) (*dto.WorkspaceSummary, error) {
	return w.open.Handle(ctx, request)
}

func (w *WorkspaceService) Rename(ctx context.Context, request dto.RenameWorkspace) (*dto.WorkspaceSummary, error) {
	return w.rename.Handle(ctx, request)
}

func (w *WorkspaceService) Remove(ctx context.Context, request dto.RemoveWorkspace) (*dto.WorkspaceList, error) {
	return w.remove.Handle(ctx, request)
}

func (w *WorkspaceService) Locate(ctx context.Context, request dto.LocateWorkspace) (*dto.WorkspaceSummary, error) {
	return w.locate.Handle(ctx, request)
}
