package mcp

import (
	"context"

	"github.com/brienze1/xenon/internal/domain/enums/aichangekind"
	"github.com/brienze1/xenon/internal/integration/entrypoint/dto"
)

type workspaceEditArguments struct {
	Id   string `json:"id"`
	Name string `json:"name"`
	Path string `json:"path"`
}

type workspaceEditView struct {
	Id   string `json:"id"`
	Name string `json:"name"`
	Path string `json:"path"`
}

func (t *tools) workspaceEdit(ctx context.Context, call Call) (Result, error) {
	var arguments workspaceEditArguments
	if err := call.Decode(&arguments); err != nil {
		return Result{}, err
	}
	summary, err := t.editWorkspace(ctx, call.Action, arguments)
	if err != nil {
		return Result{}, err
	}
	changes := []dto.AIChange{{Kind: aichangekind.Workspaces.String()}}
	if call.Action == "switch" {
		changes = append(changes, dto.AIChange{Kind: aichangekind.Tree.String()})
	}
	return Result{
		Payload: workspaceEditView{Id: summary.LocalId, Name: summary.Name, Path: summary.Path},
		Target:  dto.AIActivityTarget{Id: summary.LocalId, Name: summary.Name, Path: summary.Path},
		Changes: changes,
	}, nil
}

func (t *tools) editWorkspace(ctx context.Context, action string, arguments workspaceEditArguments) (*dto.WorkspaceSummary, error) {
	switch action {
	case "create":
		return t.controllers.WorkspaceCreate.Handle(ctx, dto.CreateWorkspace{Name: arguments.Name, Path: arguments.Path})
	case "open":
		return t.controllers.WorkspaceOpen.Handle(ctx, dto.OpenWorkspace{Path: arguments.Path})
	case "rename":
		return t.controllers.WorkspaceRename.Handle(ctx, dto.RenameWorkspace{LocalId: arguments.Id, Name: arguments.Name})
	default:
		snapshot, err := t.controllers.WorkspaceActivate.Handle(ctx, dto.ActivateWorkspace{LocalId: arguments.Id})
		if err != nil {
			return nil, err
		}
		return &snapshot.Summary, nil
	}
}
