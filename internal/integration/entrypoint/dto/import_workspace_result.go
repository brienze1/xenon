package dto

import "github.com/brienze1/xenon/internal/domain/entity"

type ImportWorkspaceResult struct {
	LocalId  string       `json:"local_id"`
	Created  ImportCounts `json:"created"`
	Replaced int          `json:"replaced"`
	Skipped  int          `json:"skipped"`
}

func NewImportWorkspaceResult(workspace entity.ImportWorkspaceResult) ImportWorkspaceResult {
	return ImportWorkspaceResult{LocalId: workspace.LocalID, Created: NewImportCounts(workspace.Created), Replaced: workspace.Replaced, Skipped: workspace.Skipped}
}
