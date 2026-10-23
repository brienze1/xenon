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
	errorHandler                 adapter.ErrorHandler
	planRunService               appAdapter.PlanRunService
	saveRunnerPreferencesService appAdapter.SaveRunnerPreferencesService
	startRunService              appAdapter.StartRunService
	controlRunService            appAdapter.ControlRunService
	findRunsService              appAdapter.FindRunsService
	deleteRunsService            appAdapter.DeleteRunsService
	exportRunService             appAdapter.ExportRunService
}

func NewRunner(
	errorHandler adapter.ErrorHandler,
	planRunService appAdapter.PlanRunService,
	saveRunnerPreferencesService appAdapter.SaveRunnerPreferencesService,
	startRunService appAdapter.StartRunService,
	controlRunService appAdapter.ControlRunService,
	findRunsService appAdapter.FindRunsService,
	deleteRunsService appAdapter.DeleteRunsService,
	exportRunService appAdapter.ExportRunService,
) adapter.RunnerController {
	return &runner{
		errorHandler:                 errorHandler,
		planRunService:               planRunService,
		saveRunnerPreferencesService: saveRunnerPreferencesService,
		startRunService:              startRunService,
		controlRunService:            controlRunService,
		findRunsService:              findRunsService,
		deleteRunsService:            deleteRunsService,
		exportRunService:             exportRunService,
	}
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

	run, err := r.startRunService.Start(ctx, request.ToEntity(), properties.Properties().Application.Version)
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

	run, err := r.findRunsService.Get(ctx, request.LocalId, request.RunId)
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

	runs, err := r.findRunsService.Active(ctx, request.LocalId)
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

	run, err := r.controlRunService.Pause(ctx, request.LocalId, request.RunId)
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

	run, err := r.controlRunService.Resume(ctx, request.LocalId, request.RunId)
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

	run, err := r.controlRunService.Stop(ctx, request.LocalId, request.RunId)
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

	plan, err := r.planRunService.Plan(ctx, request.LocalId, request.ItemId)
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

	if err = r.saveRunnerPreferencesService.Save(ctx, request.LocalId, request.ItemId, request.ToEntity()); err != nil {
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

	runs, err := r.findRunsService.List(ctx, request.LocalId, request.ItemId)
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

	kept, err := r.findRunsService.Response(ctx, request.LocalId, request.RunId, *request.Index)
	if err != nil {
		return nil, err
	}
	result := dto.NewRunRowSendResult(request.RunId, *request.Index, *kept)

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

	if err = r.deleteRunsService.Delete(ctx, request.LocalId, request.RunId); err != nil {
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

	if err = r.deleteRunsService.Clear(ctx, request.LocalId, request.ItemId); err != nil {
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

	if err = r.exportRunService.Export(ctx, request.LocalId, request.RunId, request.RunExportFormat(), request.Path); err != nil {
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

	r.controlRunService.Abandon(ctx)

	logger.Info(ctx, "Finished")
	return nil
}
