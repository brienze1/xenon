package dto

import "github.com/brienze1/xenon/internal/domain/entity"

type ImportResult struct {
	Items      []ImportItemResult      `json:"items"`
	Workspaces []ImportWorkspaceResult `json:"workspaces"`
	Warnings   []ImportWarning         `json:"warnings"`
	PmUsages   []PmUsage               `json:"pm_usages"`
	Failed     []ImportFailure         `json:"failed"`
}

func NewImportResult(result entity.ImportResult) ImportResult {
	imported := ImportResult{
		Items:      make([]ImportItemResult, 0, len(result.Items)),
		Workspaces: make([]ImportWorkspaceResult, 0, len(result.Workspaces)),
		Warnings:   NewImportWarnings(result.Warnings),
		PmUsages:   NewPmUsages(result.PmUsages),
		Failed:     make([]ImportFailure, 0, len(result.Failed)),
	}
	for _, item := range result.Items {
		imported.Items = append(imported.Items, NewImportItemResult(item))
	}
	for _, workspace := range result.Workspaces {
		imported.Workspaces = append(imported.Workspaces, NewImportWorkspaceResult(workspace))
	}
	for _, failure := range result.Failed {
		imported.Failed = append(imported.Failed, NewImportFailure(failure))
	}
	return imported
}

func (i ImportResult) LogValue() any {
	return map[string]any{"items": len(i.Items), "warnings": len(i.Warnings), "failed": len(i.Failed)}
}
