package dto

import (
	"github.com/brienze1/xenon/internal/domain/enums"
	"github.com/brienze1/xenon/internal/domain/enums/runexportformat"
)

type ExportRun struct {
	LocalId string `json:"local_id" validate:"required,not_blank"`
	RunId   string `json:"run_id" validate:"required,not_blank"`
	Format  string `json:"format" validate:"required,oneof=json junit"`
	Path    string `json:"path" validate:"required,not_blank"`
}

func (e *ExportRun) RunExportFormat() enums.RunExportFormat {
	return runexportformat.GetRunExportFormat(e.Format)
}
