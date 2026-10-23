package service

import (
	"context"

	"github.com/brienze1/xenon/internal/application/adapter"
	"github.com/brienze1/xenon/internal/application/usecase"
	"github.com/brienze1/xenon/internal/domain/entity"
	"github.com/brienze1/xenon/pkg/logger"
)

type findRuns struct {
	findWorkspaceRegistry usecase.FindWorkspaceRegistry
	loadTree              usecase.LoadTree
	findActiveRuns        usecase.FindActiveRuns
	findRuns              usecase.FindRuns
	findResponse          usecase.FindResponse
	lookup                runLookup
}

func NewFindRunsService(
	findWorkspaceRegistry usecase.FindWorkspaceRegistry,
	loadTree usecase.LoadTree,
	findActiveRuns usecase.FindActiveRuns,
	findStoredRuns usecase.FindRuns,
	findResponse usecase.FindResponse,
) adapter.FindRunsService {
	return &findRuns{
		findWorkspaceRegistry: findWorkspaceRegistry,
		loadTree:              loadTree,
		findActiveRuns:        findActiveRuns,
		findRuns:              findStoredRuns,
		findResponse:          findResponse,
		lookup:                runLookup{findActiveRuns: findActiveRuns, findRuns: findStoredRuns, findResponse: findResponse},
	}
}

func (f *findRuns) Get(ctx context.Context, localId, runId string) (*entity.Run, error) {
	logger.Info(ctx, "Started", map[string]any{"localId": localId, "runId": runId})

	run, err := f.lookup.held(ctx, localId, runId)
	if err != nil {
		return nil, err
	}

	logger.Info(ctx, "Finished", map[string]any{"localId": localId, "runId": runId, "status": run.Status})
	return run, nil
}

func (f *findRuns) List(ctx context.Context, localId, itemId string) ([]entity.Run, error) {
	logger.Info(ctx, "Started", map[string]any{"localId": localId, "itemId": itemId})

	if _, err := f.findWorkspaceRegistry.FindWorkspace(ctx, localId); err != nil {
		return nil, err
	}
	runs, err := f.findRuns.FindRuns(ctx, localId, itemId)
	if err != nil {
		return nil, err
	}

	logger.Info(ctx, "Finished", map[string]any{"localId": localId, "itemId": itemId, "runs": len(runs)})
	return runs, nil
}

func (f *findRuns) ListForAI(ctx context.Context, localId, path, itemId string) ([]entity.Run, entity.AIActivityTarget, error) {
	logger.Info(ctx, "Started", map[string]any{"localId": localId, "path": path, "itemId": itemId})

	workspace, err := f.findWorkspaceRegistry.FindWorkspace(ctx, localId)
	if err != nil {
		return nil, entity.AIActivityTarget{}, err
	}
	tree, err := f.loadTree.Load(ctx, workspace.Path)
	if err != nil {
		return nil, entity.AIActivityTarget{}, err
	}
	targetId, target := tree.ResolveRunTarget(path, itemId)
	runs, err := f.findRuns.FindRuns(ctx, localId, targetId)
	if err != nil {
		return nil, target, err
	}

	logger.Info(ctx, "Finished", map[string]any{"localId": localId, "path": path, "itemId": itemId, "targetId": targetId, "runs": len(runs)})
	return runs, target, nil
}

func (f *findRuns) Active(ctx context.Context, localId string) ([]entity.Run, error) {
	logger.Info(ctx, "Started", map[string]any{"localId": localId})

	runs := f.findActiveRuns.FindByWorkspace(ctx, localId)

	logger.Info(ctx, "Finished", map[string]any{"localId": localId, "runs": len(runs)})
	return runs, nil
}

func (f *findRuns) Response(ctx context.Context, localId, runId string, index int) (*entity.Response, error) {
	logger.Info(ctx, "Started", map[string]any{"localId": localId, "runId": runId, "index": index})

	run, err := f.lookup.held(ctx, localId, runId)
	if err != nil {
		return nil, err
	}
	responseKey, err := run.KeptResponseKey(index)
	if err != nil {
		return nil, err
	}
	response, err := f.findResponse.Find(ctx, responseKey)
	if err != nil {
		return nil, err
	}

	logger.Info(ctx, "Finished", map[string]any{"localId": localId, "runId": runId, "index": index, "status": response.StatusCode})
	return response, nil
}
