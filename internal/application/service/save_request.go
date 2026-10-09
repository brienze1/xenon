package service

import (
	"context"

	"github.com/brienze1/xenon/internal/application/adapter"
	"github.com/brienze1/xenon/internal/application/usecase"
	"github.com/brienze1/xenon/internal/domain/entity"
	"github.com/brienze1/xenon/internal/domain/enums/itemkind"
	"github.com/brienze1/xenon/pkg/errs"
	"github.com/brienze1/xenon/pkg/idutils"
	"github.com/brienze1/xenon/pkg/logger"
)

type saveRequest struct {
	findWorkspaceRegistry usecase.FindWorkspaceRegistry
	findRequest           usecase.FindRequest
	saveRequest           usecase.SaveRequest
	loadTree              usecase.LoadTree
	createTreeItem        usecase.CreateTreeItem
}

func NewSaveRequestService(
	findWorkspaceRegistry usecase.FindWorkspaceRegistry,
	findRequest usecase.FindRequest,
	saveRequestFile usecase.SaveRequest,
	loadTree usecase.LoadTree,
	createTreeItem usecase.CreateTreeItem,
) adapter.SaveRequestService {
	return &saveRequest{
		findWorkspaceRegistry: findWorkspaceRegistry,
		findRequest:           findRequest,
		saveRequest:           saveRequestFile,
		loadTree:              loadTree,
		createTreeItem:        createTreeItem,
	}
}

func (s *saveRequest) Open(ctx context.Context, localId, id string) (*entity.Request, error) {
	logger.Info(ctx, "Started", map[string]any{"localId": localId, "id": id})

	workspace, err := s.findWorkspaceRegistry.FindWorkspace(ctx, localId)
	if err != nil {
		return nil, err
	}
	request, err := s.findRequest.Find(ctx, workspace.Path, id)
	if err != nil {
		return nil, err
	}

	logger.Info(ctx, "Finished", map[string]any{"localId": localId, "id": id, "result": *request})
	return request, nil
}

func (s *saveRequest) Save(ctx context.Context, localId string, request entity.Request, baseRevision string, force bool) (*entity.Request, error) {
	logger.Info(ctx, "Started", map[string]any{"localId": localId, "request": request, "baseRevision": baseRevision, "force": force})

	workspace, err := s.findWorkspaceRegistry.FindWorkspace(ctx, localId)
	if err != nil {
		return nil, err
	}
	current, err := s.findRequest.FindRevision(ctx, workspace.Path, request.ID)
	if err != nil {
		return nil, err
	}
	if current.BlocksSave(baseRevision, force) {
		return nil, errs.ConflictError("the request changed on disk since it was opened", "REQ-02409")
	}
	saved, err := s.saveRequest.Save(ctx, workspace.Path, request.WithFileRefsRelativeTo(workspace.Path))
	if err != nil {
		return nil, err
	}

	logger.Info(ctx, "Finished", map[string]any{"localId": localId, "request": request, "baseRevision": baseRevision, "force": force, "result": *saved})
	return saved, nil
}

func (s *saveRequest) SaveAs(ctx context.Context, localId, parentId, name string, request entity.Request) (*entity.Request, error) {
	logger.Info(ctx, "Started", map[string]any{"localId": localId, "parentId": parentId, "name": name, "request": request})

	workspace, err := s.findWorkspaceRegistry.FindWorkspace(ctx, localId)
	if err != nil {
		return nil, err
	}
	tree, err := s.loadTree.Load(ctx, workspace.Path)
	if err != nil {
		return nil, err
	}
	parent, err := usecase.ResolveParent(*tree, parentId)
	if err != nil {
		return nil, err
	}
	copied := request.CopyAs(idutils.NewUlid(), name).WithFileRefsRelativeTo(workspace.Path)
	fileName := usecase.DeriveFileName(copied.Name, itemkind.Request, parent.Entries)
	saved, err := s.createTreeItem.CreateRequest(ctx, workspace.Path, parent.ID, fileName, copied)
	if err != nil {
		return nil, err
	}

	logger.Info(ctx, "Finished", map[string]any{"localId": localId, "parentId": parentId, "name": name, "request": request, "result": *saved})
	return saved, nil
}
