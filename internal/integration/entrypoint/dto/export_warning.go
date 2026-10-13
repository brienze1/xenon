package dto

import "github.com/brienze1/xenon/internal/domain/entity"

type ExportWarning struct {
	ItemPath string `json:"item_path"`
	Code     string `json:"code"`
	Message  string `json:"message"`
}

func NewExportWarning(warning entity.ExportWarning) ExportWarning {
	return ExportWarning{ItemPath: warning.ItemPath, Code: warning.Code.String(), Message: warning.Message}
}
