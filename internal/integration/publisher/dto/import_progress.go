package dto

import "github.com/brienze1/xenon/internal/domain/entity"

const ImportProgressEvent = "import:progress"

type ImportProgress struct {
	SourceId    string `json:"source_id"`
	CurrentItem string `json:"current_item"`
	Done        int    `json:"done"`
	Total       int    `json:"total"`
}

func NewImportProgress(progress entity.ImportProgress) ImportProgress {
	return ImportProgress{SourceId: progress.SourceID, CurrentItem: progress.CurrentItem, Done: progress.Done, Total: progress.Total}
}
