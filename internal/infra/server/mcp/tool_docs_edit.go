package mcp

import (
	"context"

	"github.com/brienze1/xenon/internal/domain/enums/aichangekind"
	"github.com/brienze1/xenon/internal/integration/entrypoint/dto"
)

// docsEdit leaves the requests the document drops in place unless the call confirms the deletes.
func (t *tools) docsEdit(ctx context.Context, call Call) (Result, error) {
	var arguments docsArguments
	if err := call.Decode(&arguments); err != nil {
		return Result{}, err
	}
	collection, err := t.findWorkspaceItem(ctx, call.LocalId, arguments.CollectionId)
	if err != nil {
		return Result{}, err
	}
	applied, err := t.controllers.TreeDocumentation.ApplyDocumentation(ctx, dto.DocumentationInput{
		LocalId:          call.LocalId,
		Id:               collection.Id,
		Yaml:             arguments.Yaml,
		ConfirmDeletions: arguments.ConfirmDeletes,
		KeepDeleted:      !arguments.ConfirmDeletes,
	})
	if err != nil {
		return Result{}, err
	}
	return Result{Payload: newDocsPlanView(applied.DocumentationPlan), Target: collection.target(), Changes: []dto.AIChange{{Kind: aichangekind.Tree.String()}}}, nil
}
