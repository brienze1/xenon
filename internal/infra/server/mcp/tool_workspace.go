package mcp

import (
	"context"

	"github.com/brienze1/xenon/internal/integration/entrypoint/dto"
)

type workspaceListLine struct {
	Id     string `json:"id"`
	Name   string `json:"name"`
	Active bool   `json:"active,omitempty"`
}

type activeWorkspaceView struct {
	Id            string `json:"id"`
	Name          string `json:"name"`
	Path          string `json:"path"`
	EnvironmentId string `json:"environment_id,omitempty"`
}

func (t *tools) workspace(ctx context.Context, call Call) (Result, error) {
	workspaces, err := t.controllers.WorkspaceList.Handle(ctx)
	if err != nil {
		return Result{}, err
	}
	if call.Action == "active" {
		return t.activeWorkspaceResult(ctx, *workspaces)
	}
	lines := make([]workspaceListLine, 0, len(workspaces.Workspaces))
	for _, summary := range workspaces.Workspaces {
		lines = append(lines, workspaceListLine{Id: summary.LocalId, Name: summary.Name, Active: summary.LocalId == workspaces.ActiveLocalId})
	}
	return Result{Payload: List[workspaceListLine]{Items: lines}}, nil
}

func (t *tools) activeWorkspaceResult(ctx context.Context, workspaces dto.WorkspaceList) (Result, error) {
	for _, summary := range workspaces.Workspaces {
		if summary.LocalId != workspaces.ActiveLocalId {
			continue
		}
		environments, err := t.controllers.EnvironmentList.Handle(ctx, dto.ListEnvironments{LocalId: summary.LocalId})
		if err != nil {
			return Result{}, err
		}
		view := activeWorkspaceView{Id: summary.LocalId, Name: summary.Name, Path: summary.Path, EnvironmentId: environments.ActiveId}
		return Result{Payload: view, Target: dto.AIActivityTarget{Id: summary.LocalId, Name: summary.Name}}, nil
	}
	return Result{}, noActiveWorkspace()
}
