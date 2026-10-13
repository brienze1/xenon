package entity

import "github.com/brienze1/xenon/internal/domain/enums"

type ExportWarning struct {
	Code     enums.WarningCode
	ItemPath string
	Message  string
}

func NewExportWarning(code enums.WarningCode, itemPath, message string) ExportWarning {
	return ExportWarning{Code: code, ItemPath: itemPath, Message: message}
}
