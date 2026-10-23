package controller

import (
	"context"

	appAdapter "github.com/brienze1/xenon/internal/application/adapter"
	"github.com/brienze1/xenon/internal/infra/properties"
	"github.com/brienze1/xenon/internal/integration/adapter"
	"github.com/brienze1/xenon/internal/integration/entrypoint/dto"
	"github.com/brienze1/xenon/pkg/logger"
)

type runner struct {
	errorHandler        adapter.ErrorHandler
	manageRunnerService appAdapter.ManageRunnerService
}

func NewRunner(errorHandler adapter.ErrorHandler, manageRunnerService appAdapter.ManageRunnerService) adapter.RunnerController {
	return &runner{errorHandler: errorHandler, manageRunnerService: manageRunnerService}
}

func (r *runner) Start(bindingCtx context.Context, request dto.StartRun) (response *dto.RunSummaryView, err error) {
	ctx := logger.GetContext(bindingCtx)
	logger.Info(ctx, "Started", request)

	defer func(ctx context.Context, request any) {
		if recovered := recover(); recovered != nil {
			err = handlePanic(ctx, request, recovered)
		}
		if err != nil {
			response, err = nil, r.errorHandler.Handle(ctx, err, request)
		}
	}(ctx, request)

	err = validateStartRun(&request)
	if err != nil {
		return nil, err
	}

	run, err := r.manageRunnerService.Start(ctx, request.ToEntity(), properties.Properties().Application.Version)
	if err != nil {
		return nil, err
	}
	result := dto.NewRunSummaryView(*run)

	logger.Info(ctx, "Finished", result)
	return &result, nil
}

func (r *runner) Get(bindingCtx context.Context, request dto.RunReference) (response *dto.Run, err error) {
	ctx := logger.GetContext(bindingCtx)
	logger.Info(ctx, "Started", request)

	defer func(ctx context.Context, request any) {
		if recovered := recover(); recovered != nil {
			err = handlePanic(ctx, request, recovered)
		}
		if err != nil {
			response, err = nil, r.errorHandler.Handle(ctx, err, request)
		}
	}(ctx, request)

	err = validateRunReference(&request)
	if err != nil {
		return nil, err
	}

	run, err := r.manageRunnerService.Get(ctx, request.LocalId, request.RunId)
	if err != nil {
		return nil, err
	}
	result := dto.NewRun(*run)

	logger.Info(ctx, "Finished", result)
	return &result, nil
}

func (r *runner) Active(bindingCtx context.Context, request dto.WorkspaceReference) (response []dto.RunSummaryView, err error) {
	ctx := logger.GetContext(bindingCtx)
	logger.Info(ctx, "Started", request)

	defer func(ctx context.Context, request any) {
		if recovered := recover(); recovered != nil {
			err = handlePanic(ctx, request, recovered)
		}
		if err != nil {
			response, err = nil, r.errorHandler.Handle(ctx, err, request)
		}
	}(ctx, request)

	err = validateWorkspaceReference(&request)
	if err != nil {
		return nil, err
	}

	runs, err := r.manageRunnerService.Active(ctx, request.LocalId)
	if err != nil {
		return nil, err
	}
	result := dto.NewRunSummaryViews(runs)

	logger.Info(ctx, "Finished", result)
	return result, nil
}

func (r *runner) Pause(bindingCtx context.Context, request dto.RunReference) (response *dto.RunSummaryView, err error) {
	ctx := logger.GetContext(bindingCtx)
	logger.Info(ctx, "Started", request)

	defer func(ctx context.Context, request any) {
		if recovered := recover(); recovered != nil {
			err = handlePanic(ctx, request, recovered)
		}
		if err != nil {
			response, err = nil, r.errorHandler.Handle(ctx, err, request)
		}
	}(ctx, request)

	err = validateRunReference(&request)
	if err != nil {
		return nil, err
	}

	run, err := r.manageRunnerService.Pause(ctx, request.LocalId, request.RunId)
	if err != nil {
		return nil, err
	}
	result := dto.NewRunSummaryView(*run)

	logger.Info(ctx, "Finished", result)
	return &result, nil
}

func (r *runner) Resume(bindingCtx context.Context, request dto.RunReference) (response *dto.RunSummaryView, err error) {
	ctx := logger.GetContext(bindingCtx)
	logger.Info(ctx, "Started", request)

	defer func(ctx context.Context, request any) {
		if recovered := recover(); recovered != nil {
			err = handlePanic(ctx, request, recovered)
		}
		if err != nil {
			response, err = nil, r.errorHandler.Handle(ctx, err, request)
		}
	}(ctx, request)

	err = validateRunReference(&request)
	if err != nil {
		return nil, err
	}

	run, err := r.manageRunnerService.Resume(ctx, request.LocalId, request.RunId)
	if err != nil {
		return nil, err
	}
	result := dto.NewRunSummaryView(*run)

	logger.Info(ctx, "Finished", result)
	return &result, nil
}

func (r *runner) Stop(bindingCtx context.Context, request dto.RunReference) (response *dto.RunSummaryView, err error) {
	ctx := logger.GetContext(bindingCtx)
	logger.Info(ctx, "Started", request)

	defer func(ctx context.Context, request any) {
		if recovered := recover(); recovered != nil {
			err = handlePanic(ctx, request, recovered)
		}
		if err != nil {
			response, err = nil, r.errorHandler.Handle(ctx, err, request)
		}
	}(ctx, request)

	err = validateRunReference(&request)
	if err != nil {
		return nil, err
	}

	run, err := r.manageRunnerService.Stop(ctx, request.LocalId, request.RunId)
	if err != nil {
		return nil, err
	}
	result := dto.NewRunSummaryView(*run)

	logger.Info(ctx, "Finished", result)
	return &result, nil
}
