package controller

import (
	"context"

	appAdapter "github.com/brienze1/xenon/internal/application/adapter"
	"github.com/brienze1/xenon/internal/domain/entity"
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

func (r *runner) Plan(bindingCtx context.Context, request dto.RunTargetReference) (response *dto.RunPlan, err error) {
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

	err = validateRunTargetReference(&request)
	if err != nil {
		return nil, err
	}

	plan, err := r.manageRunnerService.Plan(ctx, request.LocalId, request.ItemId)
	if err != nil {
		return nil, err
	}
	result := dto.NewRunPlan(*plan)

	logger.Info(ctx, "Finished", result)
	return &result, nil
}

func (r *runner) SavePreferences(bindingCtx context.Context, request dto.SaveRunnerPreferences) (err error) {
	ctx := logger.GetContext(bindingCtx)
	logger.Info(ctx, "Started", request)

	defer func(ctx context.Context, request any) {
		if recovered := recover(); recovered != nil {
			err = handlePanic(ctx, request, recovered)
		}
		if err != nil {
			err = r.errorHandler.Handle(ctx, err, request)
		}
	}(ctx, request)

	err = validateSaveRunnerPreferences(&request)
	if err != nil {
		return err
	}

	if err = r.manageRunnerService.SavePreferences(ctx, request.LocalId, request.ItemId, request.Config.ToEntity()); err != nil {
		return err
	}

	logger.Info(ctx, "Finished", request)
	return nil
}

func (r *runner) List(bindingCtx context.Context, request dto.RunTargetReference) (response []dto.RunSummaryView, err error) {
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

	err = validateRunTargetReference(&request)
	if err != nil {
		return nil, err
	}

	runs, err := r.manageRunnerService.List(ctx, request.LocalId, request.ItemId)
	if err != nil {
		return nil, err
	}
	result := dto.NewRunSummaryViews(runs)

	logger.Info(ctx, "Finished", result)
	return result, nil
}

func (r *runner) Response(bindingCtx context.Context, request dto.RunRowReference) (response *dto.SendResult, err error) {
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

	err = validateRunRowReference(&request)
	if err != nil {
		return nil, err
	}

	kept, err := r.manageRunnerService.Response(ctx, request.LocalId, request.RunId, *request.Index)
	if err != nil {
		return nil, err
	}
	result := dto.NewLastSendResult(entity.RunResponseKey(request.RunId, *request.Index), *kept)

	logger.Info(ctx, "Finished", result)
	return &result, nil
}

func (r *runner) Delete(bindingCtx context.Context, request dto.RunReference) (err error) {
	ctx := logger.GetContext(bindingCtx)
	logger.Info(ctx, "Started", request)

	defer func(ctx context.Context, request any) {
		if recovered := recover(); recovered != nil {
			err = handlePanic(ctx, request, recovered)
		}
		if err != nil {
			err = r.errorHandler.Handle(ctx, err, request)
		}
	}(ctx, request)

	err = validateRunReference(&request)
	if err != nil {
		return err
	}

	if err = r.manageRunnerService.Delete(ctx, request.LocalId, request.RunId); err != nil {
		return err
	}

	logger.Info(ctx, "Finished", request)
	return nil
}

func (r *runner) Clear(bindingCtx context.Context, request dto.RunTargetReference) (err error) {
	ctx := logger.GetContext(bindingCtx)
	logger.Info(ctx, "Started", request)

	defer func(ctx context.Context, request any) {
		if recovered := recover(); recovered != nil {
			err = handlePanic(ctx, request, recovered)
		}
		if err != nil {
			err = r.errorHandler.Handle(ctx, err, request)
		}
	}(ctx, request)

	err = validateRunTargetReference(&request)
	if err != nil {
		return err
	}

	if err = r.manageRunnerService.Clear(ctx, request.LocalId, request.ItemId); err != nil {
		return err
	}

	logger.Info(ctx, "Finished", request)
	return nil
}

func (r *runner) Export(bindingCtx context.Context, request dto.ExportRun) (err error) {
	ctx := logger.GetContext(bindingCtx)
	logger.Info(ctx, "Started", request)

	defer func(ctx context.Context, request any) {
		if recovered := recover(); recovered != nil {
			err = handlePanic(ctx, request, recovered)
		}
		if err != nil {
			err = r.errorHandler.Handle(ctx, err, request)
		}
	}(ctx, request)

	err = validateExportRun(&request)
	if err != nil {
		return err
	}

	if err = r.manageRunnerService.Export(ctx, request.LocalId, request.RunId, request.FormatEntity(), request.Path); err != nil {
		return err
	}

	logger.Info(ctx, "Finished", request)
	return nil
}

func (r *runner) Abandon(bindingCtx context.Context) (err error) {
	ctx := logger.GetContext(bindingCtx)
	logger.Info(ctx, "Started")

	defer func(ctx context.Context) {
		if recovered := recover(); recovered != nil {
			err = handlePanic(ctx, nil, recovered)
		}
		if err != nil {
			err = r.errorHandler.Handle(ctx, err, nil)
		}
	}(ctx)

	r.manageRunnerService.Abandon(ctx)

	logger.Info(ctx, "Finished")
	return nil
}
