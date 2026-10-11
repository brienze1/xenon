package dto

import "github.com/brienze1/xenon/internal/domain/entity"

const ExportProgressEvent = "export:progress"

type ExportProgress struct {
	Done  int `json:"done"`
	Total int `json:"total"`
}

func NewExportProgress(progress entity.ExportProgress) ExportProgress {
	return ExportProgress{Done: progress.Done, Total: progress.Total}
}
