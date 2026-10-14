package controller

import (
	"context"

	"github.com/brienze1/xenon/internal/domain/entity"
	"github.com/brienze1/xenon/internal/integration/entrypoint/dto"
)

func (m *mcpTreeEdit) createRequest(ctx context.Context, call dto.MCPToolCall, arguments dto.MCPTreeEdit) (dto.MCPToolResult, error) {
	request := arguments.CreateRequest(call.LocalId)
	err := validateCreateRequest(&request)
	if err != nil {
		return dto.MCPToolResult{}, err
	}

	created, err := m.manageTreeItemService.CreateRequest(ctx, request.LocalId, request.ParentId, request.Name, request.ToEntity())
	if err != nil {
		return dto.MCPToolResult{}, err
	}
	node := dto.NewTreeNode(*created)
	if !arguments.EditsRequest() {
		return m.edited(ctx, call, node.Id, node.MethodName())
	}

	stored, err := m.openRequest(ctx, call, node.Id)
	if err != nil {
		return dto.MCPToolResult{}, err
	}
	saved, err := m.saveRequest(ctx, arguments.SaveCreatedRequest(call.LocalId, node, stored))
	if err != nil {
		return dto.MCPToolResult{}, err
	}
	return m.edited(ctx, call, node.Id, saved.Method)
}

func (m *mcpTreeEdit) updateRequest(ctx context.Context, call dto.MCPToolCall, entry entity.TreeEntry, arguments dto.MCPTreeEdit) (dto.MCPToolResult, error) {
	stored, err := m.openRequest(ctx, call, entry.ID())
	if err != nil {
		return dto.MCPToolResult{}, err
	}
	saved, err := m.saveRequest(ctx, arguments.SaveRequest(call.LocalId, stored))
	if err != nil {
		return dto.MCPToolResult{}, err
	}

	result, err := m.edited(ctx, call, entry.ID(), saved.Method)
	if err != nil {
		return dto.MCPToolResult{}, err
	}
	return result.WithChange(dto.NewRequestAIChange(saved.Id, saved.Revision)), nil
}

func (m *mcpTreeEdit) openRequest(ctx context.Context, call dto.MCPToolCall, id string) (dto.Request, error) {
	request := dto.NewOpenRequest(call.LocalId, id)
	err := validateOpenRequest(&request)
	if err != nil {
		return dto.Request{}, err
	}

	stored, err := m.saveRequestService.Open(ctx, request.LocalId, request.Id)
	if err != nil {
		return dto.Request{}, err
	}
	return dto.NewRequest(*stored), nil
}

func (m *mcpTreeEdit) saveRequest(ctx context.Context, request dto.SaveRequest) (dto.Request, error) {
	err := validateSaveRequest(&request)
	if err != nil {
		return dto.Request{}, err
	}

	saved, err := m.saveRequestService.Save(ctx, request.LocalId, request.ToEntity(), request.BaseRevision, request.Force)
	if err != nil {
		return dto.Request{}, err
	}
	scheduleAutoSync(ctx, m.errorHandler, m.autoSyncService.Saved, request.LocalId)
	return dto.NewRequest(*saved), nil
}
