package controller

import (
	"context"

	appAdapter "github.com/brienze1/xenon/internal/application/adapter"
	"github.com/brienze1/xenon/internal/domain/entity"
	"github.com/brienze1/xenon/internal/domain/enums/aichangekind"
	"github.com/brienze1/xenon/internal/domain/enums/aitooleffect"
	"github.com/brienze1/xenon/internal/infra/properties"
	"github.com/brienze1/xenon/internal/integration/adapter"
	"github.com/brienze1/xenon/internal/integration/entrypoint/dto"
	mcpErrors "github.com/brienze1/xenon/internal/integration/entrypoint/errors"
	"github.com/brienze1/xenon/pkg/logger"
)

var sendAITool = entity.AITool{
	Name:            "send",
	Description:     "Send a saved request (or one of its aliases, with alias_id) or an ad-hoc request with the active environment and return the response.",
	Effect:          aitooleffect.Destructive,
	OpenWorld:       true,
	WorkspaceScoped: true,
}

type mcpSend struct {
	errorHandler             adapter.ErrorHandler
	searchTreeService        appAdapter.SearchTreeService
	saveRequestService       appAdapter.SaveRequestService
	sendRequestService       appAdapter.SendRequestService
	manageAppSettingsService appAdapter.ManageAppSettingsService
	approveAISendService     appAdapter.ApproveAISendService
	manageResponseService    appAdapter.ManageResponseService
	guardAIOutputService     appAdapter.GuardAIOutputService
}

func NewMCPSend(
	errorHandler adapter.ErrorHandler,
	searchTreeService appAdapter.SearchTreeService,
	saveRequestService appAdapter.SaveRequestService,
	sendRequestService appAdapter.SendRequestService,
	manageAppSettingsService appAdapter.ManageAppSettingsService,
	approveAISendService appAdapter.ApproveAISendService,
	manageResponseService appAdapter.ManageResponseService,
	guardAIOutputService appAdapter.GuardAIOutputService,
) adapter.MCPToolController {
	return &mcpSend{
		errorHandler:             errorHandler,
		searchTreeService:        searchTreeService,
		saveRequestService:       saveRequestService,
		sendRequestService:       sendRequestService,
		manageAppSettingsService: manageAppSettingsService,
		approveAISendService:     approveAISendService,
		manageResponseService:    manageResponseService,
		guardAIOutputService:     guardAIOutputService,
	}
}

func (m *mcpSend) Tool() dto.MCPTool {
	return dto.NewMCPTool(sendAITool, dto.NewMCPToolArguments[dto.MCPSend]())
}

func (m *mcpSend) Handle(toolCtx context.Context, call dto.MCPToolCall) (result dto.MCPToolResult, err error) {
	ctx := logger.GetContext(toolCtx)
	logger.Info(ctx, "Started", call)

	defer func(ctx context.Context, call any) {
		if recovered := recover(); recovered != nil {
			err = handlePanic(ctx, call, recovered)
		}
		if err != nil {
			err = m.errorHandler.Handle(ctx, err, call)
		}
	}(ctx, call)

	result, err = m.send(ctx, call)
	if err != nil {
		return result, err
	}

	logger.Info(ctx, "Finished", result)
	return result, nil
}

func (m *mcpSend) send(ctx context.Context, call dto.MCPToolCall) (dto.MCPToolResult, error) {
	var arguments dto.MCPSend
	err := call.Decode(&arguments)
	if err != nil {
		return dto.MCPToolResult{}, err
	}

	request, target, err := resolveMCPSendTarget(ctx, m.searchTreeService, m.saveRequestService, call, arguments.SendTarget())
	if err != nil {
		return dto.NewMCPTargetResult(target), err
	}

	sent, err := m.sendRequest(ctx, arguments.SendRequest(call.Client, call.LocalId, call.TabId(), "", request))
	if environment, unconfirmed := mcpErrors.UnconfirmedSendEnvironment(err); unconfirmed {
		sent, err = m.sendConfirmed(ctx, call, arguments, request, environment)
	}
	if err != nil {
		return dto.NewMCPTargetResult(target), err
	}

	response, err := newMCPSendResponse(ctx, m.errorHandler, m.manageResponseService, m.guardAIOutputService, call, sent)
	return dto.NewMCPToolResult(response).WithTarget(target).WithChanges(aichangekind.History, aichangekind.Cookies).AsRecorded(), err
}

func (m *mcpSend) sendConfirmed(ctx context.Context, call dto.MCPToolCall, arguments dto.MCPSend, request dto.Request, environment entity.AIApprovalEnvironment) (dto.SendResult, error) {
	settings, err := m.manageAppSettingsService.Get(ctx)
	if err != nil {
		return dto.SendResult{}, err
	}
	if !settings.SkipsAIApprovals() {
		err = m.approve(ctx, call, request, environment)
		if err != nil {
			return dto.SendResult{}, err
		}
	}
	return m.sendRequest(ctx, arguments.SendRequest(call.Client, call.LocalId, call.TabId(), environment.ID, request))
}

func (m *mcpSend) approve(ctx context.Context, call dto.MCPToolCall, request dto.Request, environment entity.AIApprovalEnvironment) error {
	approvalRequest := dto.NewAIApprovalRequest(call.Session, call.Client, request.Method, request.Url, dto.NewAIApprovalEnvironment(environment))
	err := validateAIArguments(&approvalRequest)
	if err != nil {
		return err
	}

	approval, err := m.approveAISendService.Request(ctx, approvalRequest.ToEntity(), approvalRequest.Session())
	if err != nil {
		return err
	}
	switch {
	case approval.IsApproved():
		return nil
	case approval.IsTimedOut():
		return mcpErrors.RejectTimedOutAISend()
	default:
		return mcpErrors.RejectDeniedAISend()
	}
}

func (m *mcpSend) sendRequest(ctx context.Context, request dto.SendRequest) (dto.SendResult, error) {
	err := validateSendRequest(&request)
	if err != nil {
		return dto.SendResult{}, err
	}

	sent, err := m.sendRequestService.Send(ctx, request.WorkspaceId(), request.TargetTabId(), request.AliasId, request.ToEntity(), properties.Properties().Application.Version, request.ToAcknowledgement(), request.ToSendOrigin())
	if err != nil {
		return dto.SendResult{}, err
	}
	return dto.NewSendResult(request.TargetTabId(), *sent), nil
}
