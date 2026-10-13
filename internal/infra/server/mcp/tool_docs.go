package mcp

import (
	"context"

	"github.com/brienze1/xenon/internal/integration/entrypoint/dto"
)

type docsArguments struct {
	CollectionId   string `json:"collection_id"`
	Yaml           string `json:"yaml"`
	ConfirmDeletes bool   `json:"confirm_deletes"`
}

type docsView struct {
	Yaml string `json:"yaml"`
}

type docsPlanView struct {
	Updated []string `json:"updated"`
	Created []string `json:"created"`
	Deleted []string `json:"deleted"`
}

func newDocsPlanView(plan dto.DocumentationPlan) docsPlanView {
	return docsPlanView{Updated: documentationNames(plan.Updated), Created: documentationNames(plan.Created), Deleted: documentationNames(plan.Deleted)}
}

func documentationNames(items []dto.DocumentationItem) []string {
	names := make([]string, 0, len(items))
	for _, item := range items {
		names = append(names, item.Name)
	}
	return names
}

func (t *tools) docs(ctx context.Context, call Call) (Result, error) {
	var arguments docsArguments
	if err := call.Decode(&arguments); err != nil {
		return Result{}, err
	}
	collection, err := t.findWorkspaceItem(ctx, call.LocalId, arguments.CollectionId)
	if err != nil {
		return Result{}, err
	}
	if call.Action == "plan" {
		return t.planDocumentation(ctx, call, collection, arguments)
	}
	documentation, err := t.controllers.TreeDocumentation.GetDocumentation(ctx, dto.TreeItemReference{LocalId: call.LocalId, Id: collection.Id})
	if err != nil {
		return Result{}, err
	}
	return Result{Payload: docsView{Yaml: documentation.Yaml}, Target: collection.target()}, nil
}

func (t *tools) planDocumentation(ctx context.Context, call Call, collection workspaceItem, arguments docsArguments) (Result, error) {
	plan, err := t.controllers.TreeDocumentation.PlanDocumentation(ctx, dto.DocumentationInput{LocalId: call.LocalId, Id: collection.Id, Yaml: arguments.Yaml})
	if err != nil {
		return Result{}, err
	}
	return Result{Payload: newDocsPlanView(*plan), Target: collection.target()}, nil
}
