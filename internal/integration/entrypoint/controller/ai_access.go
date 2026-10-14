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

type aiAccess struct {
	errorHandler              adapter.ErrorHandler
	findAIAccessStatusService appAdapter.FindAIAccessStatusService
	manageAIAccessService     appAdapter.ManageAIAccessService
	guardAIOutputService      appAdapter.GuardAIOutputService
	notifyAIChangeService     appAdapter.NotifyAIChangeService
	recordAIActivityService   appAdapter.RecordAIActivityService
	approveAISendService      appAdapter.ApproveAISendService
	findAIApprovalsService    appAdapter.FindAIApprovalsService
	publishAIAccessStatus     appAdapter.PublishAIAccessStatusService
}

func NewAIAccess(
	errorHandler adapter.ErrorHandler,
	findAIAccessStatusService appAdapter.FindAIAccessStatusService,
	manageAIAccessService appAdapter.ManageAIAccessService,
	guardAIOutputService appAdapter.GuardAIOutputService,
	notifyAIChangeService appAdapter.NotifyAIChangeService,
	recordAIActivityService appAdapter.RecordAIActivityService,
	approveAISendService appAdapter.ApproveAISendService,
	findAIApprovalsService appAdapter.FindAIApprovalsService,
	publishAIAccessStatus appAdapter.PublishAIAccessStatusService,
) adapter.AIAccessController {
	return &aiAccess{
		errorHandler:              errorHandler,
		findAIAccessStatusService: findAIAccessStatusService,
		manageAIAccessService:     manageAIAccessService,
		guardAIOutputService:      guardAIOutputService,
		notifyAIChangeService:     notifyAIChangeService,
		recordAIActivityService:   recordAIActivityService,
		approveAISendService:      approveAISendService,
		findAIApprovalsService:    findAIApprovalsService,
		publishAIAccessStatus:     publishAIAccessStatus,
	}
}

func (a *aiAccess) Status(bindingCtx context.Context) (response *dto.AIAccessStatus, err error) {
	ctx := logger.GetContext(bindingCtx)
	logger.Info(ctx, "Started")

	defer func(ctx context.Context) {
		if recovered := recover(); recovered != nil {
			err = handlePanic(ctx, nil, recovered)
		}
		if err != nil {
			response, err = nil, a.errorHandler.Handle(ctx, err, nil)
		}
	}(ctx)

	status := a.findAIAccessStatusService.Find(ctx)
	result := dto.NewAIAccessStatus(status, properties.Properties().Application.Executable)

	logger.Info(ctx, "Finished", result)
	return &result, nil
}

func (a *aiAccess) Start(bindingCtx context.Context) (response *dto.AIAccessStatus, err error) {
	ctx := logger.GetContext(bindingCtx)
	logger.Info(ctx, "Started")

	defer func(ctx context.Context) {
		if recovered := recover(); recovered != nil {
			err = handlePanic(ctx, nil, recovered)
		}
		if err != nil {
			response, err = nil, a.errorHandler.Handle(ctx, err, nil)
		}
	}(ctx)

	status, err := a.manageAIAccessService.Start(ctx)
	if status == nil {
		return nil, err
	}
	if err != nil {
		a.errorHandler.Handle(ctx, err, nil)
	}
	result := dto.NewAIAccessStatus(*status, properties.Properties().Application.Executable)

	logger.Info(ctx, "Finished", result)
	return &result, nil
}

func (a *aiAccess) Stop(bindingCtx context.Context) (response *dto.AIAccessStatus, err error) {
	ctx := logger.GetContext(bindingCtx)
	logger.Info(ctx, "Started")

	defer func(ctx context.Context) {
		if recovered := recover(); recovered != nil {
			err = handlePanic(ctx, nil, recovered)
		}
		if err != nil {
			response, err = nil, a.errorHandler.Handle(ctx, err, nil)
		}
	}(ctx)

	status, stopErr := a.manageAIAccessService.Stop(ctx)
	if stopErr != nil {
		a.errorHandler.Handle(ctx, stopErr, nil)
	}
	result := dto.NewAIAccessStatus(status, properties.Properties().Application.Executable)

	logger.Info(ctx, "Finished", result)
	return &result, nil
}

func (a *aiAccess) Guard(bindingCtx context.Context, request dto.AIGuard) (response *dto.AIGuarded, err error) {
	ctx := logger.GetContext(bindingCtx)
	logger.Info(ctx, "Started", request)

	defer func(ctx context.Context, request any) {
		if recovered := recover(); recovered != nil {
			err = handlePanic(ctx, request, recovered)
		}
		if err != nil {
			response, err = nil, a.errorHandler.Handle(ctx, err, request)
		}
	}(ctx, request)

	err = validateAIGuard(&request)
	if err != nil {
		return nil, err
	}

	masked, guardErr := a.guardAIOutputService.Guard(ctx, request.WorkspaceId(), request.GuardedText())
	if guardErr != nil {
		a.errorHandler.Handle(ctx, guardErr, request)
	}
	result := dto.NewAIGuarded(masked)

	logger.Info(ctx, "Finished", result)
	return &result, nil
}

func (a *aiAccess) Notify(bindingCtx context.Context, request dto.AIChanges) (err error) {
	ctx := logger.GetContext(bindingCtx)
	logger.Info(ctx, "Started", request)

	defer func(ctx context.Context, request any) {
		if recovered := recover(); recovered != nil {
			err = handlePanic(ctx, request, recovered)
		}
		if err != nil {
			err = a.errorHandler.Handle(ctx, err, request)
		}
	}(ctx, request)

	err = validateAIChanges(&request)
	if err != nil {
		return err
	}

	err = a.notifyAIChangeService.Notify(ctx, request.WorkspaceId(), request.ToEntity())
	if err != nil {
		return err
	}

	logger.Info(ctx, "Finished", request)
	return nil
}

func (a *aiAccess) Record(bindingCtx context.Context, request dto.AIActivityRecord) (err error) {
	ctx := logger.GetContext(bindingCtx)
	logger.Info(ctx, "Started", request)

	defer func(ctx context.Context, request any) {
		if recovered := recover(); recovered != nil {
			err = handlePanic(ctx, request, recovered)
		}
		if err != nil {
			err = a.errorHandler.Handle(ctx, err, request)
		}
	}(ctx, request)

	err = validateAIActivityRecord(&request)
	if err != nil {
		return err
	}

	if recordErr := a.recordAIActivityService.Record(ctx, request.WorkspaceId(), request.ClientName(), request.ToEntity()); recordErr != nil {
		a.errorHandler.Handle(ctx, recordErr, request)
	}

	logger.Info(ctx, "Finished", request)
	return nil
}

func (a *aiAccess) RequestApproval(bindingCtx context.Context, request dto.AIApprovalRequest) (response *dto.AIApprovalOutcome, err error) {
	ctx := logger.GetContext(bindingCtx)
	logger.Info(ctx, "Started", request)

	defer func(ctx context.Context, request any) {
		if recovered := recover(); recovered != nil {
			err = handlePanic(ctx, request, recovered)
		}
		if err != nil {
			response, err = nil, a.errorHandler.Handle(ctx, err, request)
		}
	}(ctx, request)

	err = validateAIArguments(&request)
	if err != nil {
		return nil, err
	}

	approval, err := a.approveAISendService.Request(ctx, request.ToEntity(), request.Session())
	if err != nil {
		return nil, err
	}
	result := dto.NewAIApprovalOutcome(approval)

	logger.Info(ctx, "Finished", result)
	return &result, nil
}

func (a *aiAccess) AnswerApproval(bindingCtx context.Context, request dto.AIApprovalAnswer) (err error) {
	ctx := logger.GetContext(bindingCtx)
	logger.Info(ctx, "Started", request)

	defer func(ctx context.Context, request any) {
		if recovered := recover(); recovered != nil {
			err = handlePanic(ctx, request, recovered)
		}
		if err != nil {
			err = a.errorHandler.Handle(ctx, err, request)
		}
	}(ctx, request)

	err = validateAIArguments(&request)
	if err != nil {
		return err
	}

	err = a.approveAISendService.Answer(ctx, request.ApprovalId(), request.IsApproved())
	if err != nil {
		return err
	}

	logger.Info(ctx, "Finished", request)
	return nil
}

func (a *aiAccess) PendingApprovals(bindingCtx context.Context) (response []dto.AIApproval, err error) {
	ctx := logger.GetContext(bindingCtx)
	logger.Info(ctx, "Started")

	defer func(ctx context.Context) {
		if recovered := recover(); recovered != nil {
			err = handlePanic(ctx, nil, recovered)
		}
		if err != nil {
			response, err = nil, a.errorHandler.Handle(ctx, err, nil)
		}
	}(ctx)

	result := dto.NewAIApprovals(a.findAIApprovalsService.FindPending(ctx))

	logger.Info(ctx, "Finished", result)
	return result, nil
}

func (a *aiAccess) Shutdown(bindingCtx context.Context) (err error) {
	ctx := logger.GetContext(bindingCtx)
	logger.Info(ctx, "Started")

	defer func(ctx context.Context) {
		if recovered := recover(); recovered != nil {
			err = handlePanic(ctx, nil, recovered)
		}
		if err != nil {
			err = a.errorHandler.Handle(ctx, err, nil)
		}
	}(ctx)

	err = a.approveAISendService.Deny(ctx, entity.EverySessionID)
	if err != nil {
		return err
	}
	status, stopErr := a.manageAIAccessService.Stop(ctx)
	if stopErr != nil {
		a.errorHandler.Handle(ctx, stopErr, nil)
	}

	logger.Info(ctx, "Finished", status)
	return nil
}

func (a *aiAccess) ClientsChanged(bindingCtx context.Context) (err error) {
	ctx := logger.GetContext(bindingCtx)
	logger.Info(ctx, "Started")

	defer func(ctx context.Context) {
		if recovered := recover(); recovered != nil {
			err = handlePanic(ctx, nil, recovered)
		}
		if err != nil {
			err = a.errorHandler.Handle(ctx, err, nil)
		}
	}(ctx)

	err = a.publishAIAccessStatus.Publish(ctx)
	if err != nil {
		return err
	}

	logger.Info(ctx, "Finished")
	return nil
}
