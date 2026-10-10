package dto

import "github.com/brienze1/xenon/internal/domain/entity"

type ImportItemResult struct {
	Key     string `json:"key"`
	Status  string `json:"status"`
	LocalId string `json:"local_id"`
	Id      string `json:"id"`
	Name    string `json:"name"`
}

type ImportWorkspaceResult struct {
	LocalId  string       `json:"local_id"`
	Created  ImportCounts `json:"created"`
	Replaced int          `json:"replaced"`
	Skipped  int          `json:"skipped"`
}

type ImportFailure struct {
	ItemKey string `json:"item_key"`
	Name    string `json:"name"`
	Code    string `json:"code"`
	Message string `json:"message"`
}

type ImportResult struct {
	Items      []ImportItemResult      `json:"items"`
	Workspaces []ImportWorkspaceResult `json:"workspaces"`
	Warnings   []ImportWarning         `json:"warnings"`
	PmUsages   []PmUsage               `json:"pm_usages"`
	Failed     []ImportFailure         `json:"failed"`
}

func NewImportResult(result entity.ImportResult) ImportResult {
	dto := ImportResult{
		Items:      make([]ImportItemResult, 0, len(result.Items)),
		Workspaces: make([]ImportWorkspaceResult, 0, len(result.Workspaces)),
		Warnings:   NewImportWarnings(result.Warnings),
		PmUsages:   NewPmUsages(result.PmUsages),
		Failed:     make([]ImportFailure, 0, len(result.Failed)),
	}
	for _, item := range result.Items {
		dto.Items = append(dto.Items, ImportItemResult{Key: item.Key, Status: item.Status.String(), LocalId: item.LocalID, Id: item.ID, Name: item.Name})
	}
	for _, workspace := range result.Workspaces {
		dto.Workspaces = append(dto.Workspaces, ImportWorkspaceResult{LocalId: workspace.LocalID, Created: NewImportCounts(workspace.Created), Replaced: workspace.Replaced, Skipped: workspace.Skipped})
	}
	for _, failure := range result.Failed {
		dto.Failed = append(dto.Failed, ImportFailure{ItemKey: failure.ItemKey, Name: failure.Name, Code: failure.Code, Message: failure.Message})
	}
	return dto
}

func (i ImportResult) LogValue() any {
	return map[string]any{"items": len(i.Items), "warnings": len(i.Warnings), "failed": len(i.Failed)}
}
