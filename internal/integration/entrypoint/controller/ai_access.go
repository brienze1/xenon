package controller

import (
	"context"

	appAdapter "github.com/brienze1/xenon/internal/application/adapter"
	"github.com/brienze1/xenon/internal/infra/properties"
	"github.com/brienze1/xenon/internal/integration/adapter"
	"github.com/brienze1/xenon/internal/integration/entrypoint/dto"
	"github.com/brienze1/xenon/internal/integration/entrypoint/validator"
	"github.com/brienze1/xenon/pkg/logger"
)

type aiAccess struct {
	errorHandler            adapter.ErrorHandler
	manageAIAccessService   appAdapter.ManageAIAccessService
	guardAIOutputService    appAdapter.GuardAIOutputService
	notifyAIChangeService   appAdapter.NotifyAIChangeService
	recordAIActivityService appAdapter.RecordAIActivityService
	approveAISendService    appAdapter.ApproveAISendService
}

func NewAIAccess(
	errorHandler adapter.ErrorHandler,
	manageAIAccessService appAdapter.ManageAIAccessService,
	guardAIOutputService appAdapter.GuardAIOutputService,
	notifyAIChangeService appAdapter.NotifyAIChangeService,
	recordAIActivityService appAdapter.RecordAIActivityService,
	approveAISendService appAdapter.ApproveAISendService,
) adapter.AIAccessController {
	return &aiAccess{
		errorHandler:            errorHandler,
		manageAIAccessService:   manageAIAccessService,
		guardAIOutputService:    guardAIOutputService,
		notifyAIChangeService:   notifyAIChangeService,
		recordAIActivityService: recordAIActivityService,
		approveAISendService:    approveAISendService,
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

	status := a.manageAIAccessService.Status(ctx)
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
	if err != nil {
		return nil, err
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

	status := a.manageAIAccessService.Stop(ctx)
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

	err = validator.CustomValidator().Struct("MCP-09400", &request)
	if err != nil {
		return nil, err
	}

	result := dto.NewAIGuarded(a.guardAIOutputService.Guard(ctx, request.LocalId, request.Text))

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

	err = validator.CustomValidator().Struct("MCP-10400", &request)
	if err != nil {
		return err
	}

	err = a.notifyAIChangeService.Notify(ctx, request.LocalId, request.ToEntity())
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

	err = validator.CustomValidator().Struct("MCP-11400", &request)
	if err != nil {
		return err
	}

	a.recordAIActivityService.Record(ctx, request.LocalId, request.Client, request.ToEntity())

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

	err = validator.CustomValidator().Struct("MCP-07400", &request)
	if err != nil {
		return nil, err
	}

	approval, err := a.approveAISendService.Request(ctx, request.ToEntity(), request.SessionId)
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

	err = validator.CustomValidator().Struct("MCP-07400", &request)
	if err != nil {
		return err
	}

	err = a.approveAISendService.Answer(ctx, request.Id, request.Approved)
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

	result := dto.NewAIApprovals(a.approveAISendService.Pending(ctx))

	logger.Info(ctx, "Finished", map[string]any{"pending": len(result)})
	return result, nil
}

func (a *aiAccess) DenyApprovals(bindingCtx context.Context, request dto.AIApprovalDenial) (err error) {
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

	if request.IsAll() {
		err = a.approveAISendService.DenyAll(ctx)
	} else {
		err = a.approveAISendService.DenySession(ctx, request.SessionId)
	}
	if err != nil {
		return err
	}

	logger.Info(ctx, "Finished", request)
	return nil
}
