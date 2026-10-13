package dto

import "github.com/brienze1/xenon/internal/domain/entity"

type ImportWarning struct {
	Code     string `json:"code"`
	Severity string `json:"severity"`
	Message  string `json:"message"`
	ItemKey  string `json:"item_key"`
	ItemId   string `json:"item_id"`
	LocalId  string `json:"local_id"`
	Path     string `json:"path"`
	Location string `json:"location"`
}

func NewImportWarning(warning entity.ImportWarning) ImportWarning {
	return ImportWarning{
		Code:     warning.Code.String(),
		Severity: warning.Severity.String(),
		Message:  warning.Message,
		ItemKey:  warning.ItemKey,
		ItemId:   warning.ItemID,
		LocalId:  warning.LocalID,
		Path:     warning.Path,
		Location: warning.Location,
	}
}

func NewImportWarnings(warnings []entity.ImportWarning) []ImportWarning {
	result := make([]ImportWarning, 0, len(warnings))
	for _, warning := range warnings {
		result = append(result, NewImportWarning(warning))
	}
	return result
}
