package service

import (
	"context"

	"github.com/brienze1/xenon/internal/application/adapter"
	"github.com/brienze1/xenon/internal/application/usecase"
	"github.com/brienze1/xenon/internal/domain/entity"
	"github.com/brienze1/xenon/pkg/idutils"
	"github.com/brienze1/xenon/pkg/logger"
	"github.com/brienze1/xenon/pkg/timeutils"
)

type sendRequest struct {
	findWorkspaceRegistry usecase.FindWorkspaceRegistry
	findTreeItem          usecase.FindTreeItem
	sendHttpRequest       usecase.SendHttpRequest
	saveResponse          usecase.SaveResponse
	saveHistoryEntry      usecase.SaveHistoryEntry
	appVersion            string
}

func NewSendRequestService(
	findWorkspaceRegistry usecase.FindWorkspaceRegistry,
	findTreeItem usecase.FindTreeItem,
	sendHttpRequest usecase.SendHttpRequest,
	saveResponse usecase.SaveResponse,
	saveHistoryEntry usecase.SaveHistoryEntry,
	appVersion string,
) adapter.SendRequestService {
	return &sendRequest{
		findWorkspaceRegistry: findWorkspaceRegistry,
		findTreeItem:          findTreeItem,
		sendHttpRequest:       sendHttpRequest,
		saveResponse:          saveResponse,
		saveHistoryEntry:      saveHistoryEntry,
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
	sentAt := timeutils.Now()
	effective := usecase.ResolveEffectiveRequest(request, ancestors, workspace.Path)
	effective = effective.WithAutoHeaders(usecase.ComputeAutoHeaders(effective, s.appVersion))
	response := s.sendHttpRequest.Send(ctx, effective)
	if err := s.saveResponse.Save(ctx, tabId, response); err != nil {
		return nil, err
	}
	if err := s.saveHistoryEntry.Save(ctx, localId, entity.NewHistoryEntry(idutils.NewUlid(), sentAt, request, response)); err != nil {
		return nil, err
	}

	logger.Info(ctx, "Finished", response, map[string]any{"localId": localId, "tabId": tabId})
	return &response, nil
}
