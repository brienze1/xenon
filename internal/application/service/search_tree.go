package service

import (
	"context"

	"github.com/brienze1/xenon/internal/application/adapter"
	"github.com/brienze1/xenon/internal/application/usecase"
	"github.com/brienze1/xenon/internal/domain/entity"
	"github.com/brienze1/xenon/pkg/logger"
)

type searchTree struct {
	findWorkspaceRegistry usecase.FindWorkspaceRegistry
	loadTree              usecase.LoadTree
}

func NewSearchTreeService(findWorkspaceRegistry usecase.FindWorkspaceRegistry, loadTree usecase.LoadTree) adapter.SearchTreeService {
	return &searchTree{findWorkspaceRegistry: findWorkspaceRegistry, loadTree: loadTree}
}

func (s *searchTree) Tree(ctx context.Context, localId string) (*entity.Tree, error) {
	logger.Info(ctx, "Started", map[string]any{"localId": localId})

	workspace, err := s.findWorkspaceRegistry.FindWorkspace(ctx, localId)
	if err != nil {
		return nil, err
	}
	tree, err := s.loadTree.Load(ctx, workspace.Path)
	if err != nil {
		return nil, err
	}

	logger.Info(ctx, "Finished", map[string]any{"localId": localId, "result": tree})
	return tree, nil
}

func (s *searchTree) Search(ctx context.Context, localId, query string) (*entity.TreeSearch, error) {
	logger.Info(ctx, "Started", map[string]any{"localId": localId, "query": query})

	workspace, err := s.findWorkspaceRegistry.FindWorkspace(ctx, localId)
	if err != nil {
		return nil, err
	}
	tree, err := s.loadTree.Load(ctx, workspace.Path)
	if err != nil {
		return nil, err
	}
	result := usecase.SearchTree(*tree, query).WithTree(*tree)

	logger.Info(ctx, "Finished", map[string]any{"localId": localId, "query": query, "result": result})
	return &result, nil
}

func (s *searchTree) QuickSearch(ctx context.Context, localId, query string, limit int) ([]entity.QuickSearchHit, error) {
	logger.Info(ctx, "Started", map[string]any{"localId": localId, "query": query, "limit": limit})

	workspace, err := s.findWorkspaceRegistry.FindWorkspace(ctx, localId)
	if err != nil {
		return nil, err
	}
	tree, err := s.loadTree.Load(ctx, workspace.Path)
	if err != nil {
		return nil, err
	}
	result := usecase.RankQuickSearch(*tree, query, limit)

	logger.Info(ctx, "Finished", map[string]any{"localId": localId, "query": query, "limit": limit, "result": result})
	return result, nil
}
