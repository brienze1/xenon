package controller

import (
	"context"

	"github.com/brienze1/xenon/internal/domain/entity"
	"github.com/brienze1/xenon/internal/domain/enums/aichangekind"
	"github.com/brienze1/xenon/internal/integration/entrypoint/dto"
	mcpErrors "github.com/brienze1/xenon/internal/integration/entrypoint/errors"
	"github.com/brienze1/xenon/pkg/timeutils"
)

func (m *mcpTreeEdit) saveSettings(ctx context.Context, call dto.MCPToolCall, entry entity.TreeEntry, arguments dto.MCPTreeEdit) (dto.MCPToolResult, error) {
	if !entry.IsContainer() {
		return dto.MCPToolResult{}, mcpErrors.RejectInvalidArguments(call.Tool, mcpTreeSettingsReason)
	}
	if arguments.HasVariables() {
		if !entry.IsCollection() {
			return dto.MCPToolResult{}, mcpErrors.RejectInvalidArguments(call.Tool, mcpTreeEditVariablesReason)
		}
		_, err := saveMCPVariableRows(ctx, m.manageVariableTableService, call, entity.NewCollectionScopeRef(entry.ID()).String(), arguments.VariableRows())
		if err != nil {
			return dto.MCPToolResult{}, err
		}
	}
	if arguments.EditsSettings() {
		err := m.saveItemSettings(ctx, call, entry, arguments)
		if err != nil {
			return dto.MCPToolResult{}, err
		}
	}

	result, err := m.edited(ctx, call, entry.ID(), "")
	if err != nil {
		return dto.MCPToolResult{}, err
	}
	if arguments.HasVariables() {
		return result.WithChanges(aichangekind.Variables), nil
	}
	return result, nil
}

func (m *mcpTreeEdit) saveItemSettings(ctx context.Context, call dto.MCPToolCall, entry entity.TreeEntry, arguments dto.MCPTreeEdit) error {
	reference := dto.NewTreeItemReference(call.LocalId, entry.ID())
	err := validateGetItemSettings(&reference)
	if err != nil {
		return err
	}

	stored, err := m.manageItemSettingsService.Get(ctx, reference.LocalId, reference.Id)
	if err != nil {
		return err
	}

	request := arguments.SaveItemSettings(call.LocalId, entry.ID(), dto.NewItemSettings(*stored))
	err = validateSaveItemSettings(&request)
	if err != nil {
		return err
	}

	_, err = m.manageItemSettingsService.Save(ctx, request.LocalId, request.ToEntity())
	if err != nil {
		return err
	}
	scheduleAutoSync(ctx, m.errorHandler, m.autoSyncService.Saved, request.LocalId)
	return nil
}

func (m *mcpTreeEdit) token(ctx context.Context, call dto.MCPToolCall, arguments dto.MCPTreeEdit) (dto.MCPToolResult, error) {
	owner, err := findMCPTreeEntry(ctx, m.searchTreeService, call.LocalId, arguments.OwnerItemId())
	if err != nil {
		return dto.MCPToolResult{}, err
	}

	var token *entity.OAuth2Token
	switch call.Action {
	case mcpTreeEditGetTokenAction:
		token, err = m.newToken(ctx, call, owner, arguments)
	case mcpTreeEditRefreshTokenAction:
		token, err = m.refreshToken(ctx, call, owner, arguments)
	default:
		token, err = m.useToken(ctx, call, owner, arguments)
	}
	if err != nil {
		return dto.MCPToolResult{}, err
	}
	return dto.NewMCPToolResult(dto.NewMCPTreeEditToken(dto.NewTokenInfo(*token, timeutils.Now()))).WithTarget(owner.Target()), nil
}

func (m *mcpTreeEdit) newToken(ctx context.Context, call dto.MCPToolCall, owner entity.TreeEntry, arguments dto.MCPTreeEdit) (*entity.OAuth2Token, error) {
	request := arguments.OAuthTokenRequest(call.LocalId, owner.ID())
	err := validateGetNewOAuthToken(&request)
	if err != nil {
		return nil, err
	}
	return m.manageOAuthService.GetNewToken(ctx, request.LocalId, request.OwnerId, request.Draft())
}

func (m *mcpTreeEdit) refreshToken(ctx context.Context, call dto.MCPToolCall, owner entity.TreeEntry, arguments dto.MCPTreeEdit) (*entity.OAuth2Token, error) {
	request := arguments.OAuthTokenRequest(call.LocalId, owner.ID())
	err := validateRefreshOAuthToken(&request)
	if err != nil {
		return nil, err
	}
	return m.manageOAuthService.Refresh(ctx, request.LocalId, request.OwnerId, request.Draft())
}

func (m *mcpTreeEdit) useToken(ctx context.Context, call dto.MCPToolCall, owner entity.TreeEntry, arguments dto.MCPTreeEdit) (*entity.OAuth2Token, error) {
	request := arguments.OAuthTokenUse(call.LocalId, owner.ID())
	err := validateUseOAuthToken(&request)
	if err != nil {
		return nil, err
	}
	return m.manageOAuthService.Use(ctx, request.LocalId, request.OwnerId, request.UseToken)
}
