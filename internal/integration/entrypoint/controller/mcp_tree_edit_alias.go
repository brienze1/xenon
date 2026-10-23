package controller

import (
	"context"

	"github.com/brienze1/xenon/internal/domain/entity"
	"github.com/brienze1/xenon/internal/integration/entrypoint/dto"
	mcpErrors "github.com/brienze1/xenon/internal/integration/entrypoint/errors"
)

const mcpTreeEditAliasReason = "an alias supports create, update and rename"

func (m *mcpTreeEdit) editAlias(ctx context.Context, call dto.MCPToolCall, arguments dto.MCPTreeEdit) (dto.MCPToolResult, error) {
	entry, err := findMCPTreeEntry(ctx, m.searchTreeService, call.LocalId, arguments.RequestItemId())
	if err != nil {
		return dto.MCPToolResult{}, err
	}
	if !entry.IsRequest() {
		return dto.MCPToolResult{}, mcpErrors.RejectInvalidArguments(call.Tool, mcpTreeEditRequestReason)
	}

	switch call.Action {
	case mcpTreeEditCreateAction:
		return m.createAlias(ctx, call, entry, arguments)
	case mcpTreeEditUpdateAction, mcpTreeEditRenameAction:
		return m.saveAlias(ctx, call, entry, arguments)
	default:
		return dto.MCPToolResult{}, mcpErrors.RejectInvalidArguments(call.Tool, mcpTreeEditAliasReason)
	}
}

func (m *mcpTreeEdit) createAlias(ctx context.Context, call dto.MCPToolCall, entry entity.TreeEntry, arguments dto.MCPTreeEdit) (dto.MCPToolResult, error) {
	request := arguments.CreateAlias(call.LocalId, entry.ID())
	err := validateCreateAlias(&request)
	if err != nil {
		return dto.MCPToolResult{}, err
	}

	saved, err := m.manageAliasService.Create(ctx, request.LocalId, request.RequestId, request.Name, request.EnvironmentIds, request.BaseRevision, true)
	if err != nil {
		return dto.MCPToolResult{}, err
	}
	return m.aliasEdited(ctx, call, entry, dto.NewRequest(*saved), saved.NewestAliasID())
}

func (m *mcpTreeEdit) saveAlias(ctx context.Context, call dto.MCPToolCall, entry entity.TreeEntry, arguments dto.MCPTreeEdit) (dto.MCPToolResult, error) {
	stored, err := m.openRequest(ctx, call, entry.ID())
	if err != nil {
		return dto.MCPToolResult{}, err
	}
	request := arguments.SaveAlias(call.LocalId, stored)
	err = validateSaveAlias(&request)
	if err != nil {
		return dto.MCPToolResult{}, err
	}
	if call.Action == mcpTreeEditUpdateAction && arguments.SetsAliasEnvironments() {
		if _, err := m.setAliasEnvironments(ctx, arguments.SetAliasEnvironments(call.LocalId, stored)); err != nil {
			return dto.MCPToolResult{}, err
		}
	}

	saved, err := m.manageAliasService.Save(ctx, request.LocalId, request.RequestId, request.ToEntity(), request.BaseRevision, request.Force)
	if err != nil {
		return dto.MCPToolResult{}, err
	}
	return m.aliasEdited(ctx, call, entry, dto.NewRequest(*saved), request.Alias.Id)
}

func (m *mcpTreeEdit) setAliasEnvironments(ctx context.Context, request dto.SetAliasEnvironments) (*entity.Request, error) {
	err := validateSetAliasEnvironments(&request)
	if err != nil {
		return nil, err
	}
	return m.manageAliasService.SetEnvironments(ctx, request.LocalId, request.RequestId, request.AliasId, request.EnvironmentIds, request.BaseRevision, true)
}

func (m *mcpTreeEdit) aliasEdited(ctx context.Context, call dto.MCPToolCall, entry entity.TreeEntry, saved dto.Request, aliasId string) (dto.MCPToolResult, error) {
	scheduleAutoSync(ctx, m.errorHandler, m.autoSyncService.Saved, call.LocalId)
	environments, err := m.manageEnvironmentService.List(ctx, call.LocalId)
	if err != nil {
		return dto.MCPToolResult{}, err
	}
	alias := dto.NewMCPTreeAlias(saved.AliasOrEmpty(aliasId), environments.Environments)
	return dto.NewMCPToolResult(alias).WithTarget(entry.Target()).WithChange(dto.NewRequestAIChange(saved.Id, saved.Revision)), nil
}
