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
	Description:     "Send a saved request (or one of its aliases, with alias_id) or an ad-hoc request with the session's environment and return the response.",
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
	manageAISessionService   appAdapter.ManageAISessionService
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
	manageAISessionService appAdapter.ManageAISessionService,
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
		manageAISessionService:   manageAISessionService,
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

	environmentId, err := m.environment(ctx, call, arguments)
	if err != nil {
		return dto.MCPToolResult{}, err
	}

	request, target, err := resolveMCPSendTarget(ctx, m.searchTreeService, m.saveRequestService, call, arguments.SendTarget())
	if err != nil {
		return dto.NewMCPTargetResult(target), err
	}

	sent, err := m.sendRequest(ctx, arguments.SendRequest(call.Client, call.SessionTag(), call.LocalId, call.TabId(), environmentId, "", request))
	if environment, unconfirmed := mcpErrors.UnconfirmedSendEnvironment(err); unconfirmed {
		sent, err = m.sendConfirmed(ctx, call, arguments, environmentId, request, environment)
	}
	if err != nil {
		return dto.NewMCPTargetResult(target), err
	}

	response, err := newMCPSendResponse(ctx, m.errorHandler, m.manageResponseService, m.guardAIOutputService, call, sent)
	return dto.NewMCPToolResult(response).WithTarget(target).WithChanges(aichangekind.History, aichangekind.Cookies).AsRecorded(), err
}

func (m *mcpSend) environment(ctx context.Context, call dto.MCPToolCall, arguments dto.MCPSend) (string, error) {
	if !arguments.OverridesEnvironment() {
		return call.EnvironmentId, nil
	}
	environmentId, known, err := m.manageAISessionService.ResolveEnvironment(ctx, call.LocalId, arguments.Environment)
	if err != nil {
		return "", err
	}
	if !known {
		return "", mcpErrors.RejectInvalidArguments(call.Tool, mcpRunEnvironmentNotFound)
	}
	return environmentId, nil
}

func (m *mcpSend) sendConfirmed(ctx context.Context, call dto.MCPToolCall, arguments dto.MCPSend, environmentId string, request dto.Request, environment entity.AIApprovalEnvironment) (dto.SendResult, error) {
	settings, err := m.manageAppSettingsService.Get(ctx)
	if err != nil {
		return dto.SendResult{}, err
	}
	if !settings.SkipsAIApprovals() {
		err = askMCPApproval(ctx, m.approveAISendService, dto.NewAIApprovalRequest(call.Session, call.Client, request.Method, request.Url, dto.NewAIApprovalEnvironment(environment)))
		if err != nil {
			return dto.SendResult{}, err
		}
	}
	return m.sendRequest(ctx, arguments.SendRequest(call.Client, call.SessionTag(), call.LocalId, call.TabId(), environmentId, environment.ID, request))
}

func (m *mcpSend) sendRequest(ctx context.Context, request dto.SendRequest) (dto.SendResult, error) {
	err := validateSendRequest(&request)
	if err != nil {
		return dto.SendResult{}, err
	}

	sent, err := m.sendRequestService.Send(ctx, request.WorkspaceId(), request.TargetTabId(), request.AliasId, request.EnvironmentId, request.ToEntity(), properties.Properties().Application.Version, request.ToAcknowledgement(), request.ToSendOrigin())
	if err != nil {
		return dto.SendResult{}, err
	}
	return dto.NewSendResult(request.TargetTabId(), *sent), nil
}
