package service

import (
	"context"

	"github.com/brienze1/xenon/internal/application/adapter"
	"github.com/brienze1/xenon/internal/application/usecase"
	"github.com/brienze1/xenon/internal/domain/entity"
	"github.com/brienze1/xenon/pkg/logger"
)

type sendRequest struct {
	findWorkspaceRegistry usecase.FindWorkspaceRegistry
	findTreeItem          usecase.FindTreeItem
	sendHttpRequest       usecase.SendHttpRequest
	saveResponse          usecase.SaveResponse
	appVersion            string
}

func NewSendRequestService(
	findWorkspaceRegistry usecase.FindWorkspaceRegistry,
	findTreeItem usecase.FindTreeItem,
	sendHttpRequest usecase.SendHttpRequest,
	saveResponse usecase.SaveResponse,
	appVersion string,
) adapter.SendRequestService {
	return &sendRequest{
		findWorkspaceRegistry: findWorkspaceRegistry,
		findTreeItem:          findTreeItem,
		sendHttpRequest:       sendHttpRequest,
		saveResponse:          saveResponse,
		appVersion:            appVersion,
	}
}

func (s *sendRequest) Send(ctx context.Context, localId, tabId string, request entity.Request) (*entity.Response, error) {
	logger.Info(ctx, "Started", request, map[string]any{"localId": localId, "tabId": tabId})

	workspace, err := s.findWorkspaceRegistry.FindWorkspace(ctx, localId)
	if err != nil {
		return nil, err
	}
	ancestors, err := s.findTreeItem.FindAncestors(ctx, workspace.Path, request.ID)
	if err != nil {
		return nil, err
	}
	effective := usecase.ResolveEffectiveRequest(request, ancestors, workspace.Path)
	effective = effective.WithAutoHeaders(usecase.ComputeAutoHeaders(effective, s.appVersion))
	response := s.sendHttpRequest.Send(ctx, effective)
	if err := s.saveResponse.Save(ctx, tabId, response); err != nil {
		return nil, err
	}

	logger.Info(ctx, "Finished", response, map[string]any{"localId": localId, "tabId": tabId})
	return &response, nil
}
