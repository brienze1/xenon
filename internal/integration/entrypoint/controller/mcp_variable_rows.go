package controller

import (
	"context"

	appAdapter "github.com/brienze1/xenon/internal/application/adapter"
	"github.com/brienze1/xenon/internal/integration/entrypoint/dto"
	mcpErrors "github.com/brienze1/xenon/internal/integration/entrypoint/errors"
)

const (
	mcpMisplacedSecretValueReason = "secret_value is only accepted on a row with secret: true"
	mcpRevealedSecretAction       = "Turning a secret variable plain"
)

func saveMCPVariableRows(ctx context.Context, manageVariableTableService appAdapter.ManageVariableTableService, call dto.MCPToolCall, scope string, rows []dto.MCPVariableRow) (dto.VariableTable, error) {
	if dto.HasMisplacedSecretValue(rows) {
		return dto.VariableTable{}, mcpErrors.RejectInvalidArguments(call.Tool, mcpMisplacedSecretValueReason)
	}

	request := dto.NewSaveVariableTable(call.LocalId, scope, dto.NewVariableEdits(rows))
	err := validateSaveVariableTable(&request)
	if err != nil {
		return dto.VariableTable{}, err
	}
	reference := dto.NewVariableTableReference(request.LocalId, request.ScopeRef())
	err = validateGetVariableTable(&reference)
	if err != nil {
		return dto.VariableTable{}, err
	}

	stored, err := manageVariableTableService.GetTable(ctx, reference.LocalId, reference.ScopeRef())
	if err != nil {
		return dto.VariableTable{}, err
	}
	if stored.RevealedBy(request.ToEntity()) {
		return dto.VariableTable{}, mcpErrors.RejectHumanOnlyAIAction(mcpRevealedSecretAction)
	}

	saved, err := manageVariableTableService.SaveTable(ctx, request.LocalId, request.ScopeRef(), request.ToEntity(), request.ConfirmReveal)
	if err != nil {
		return dto.VariableTable{}, err
	}
	return dto.NewVariableTable(*saved), nil
}
