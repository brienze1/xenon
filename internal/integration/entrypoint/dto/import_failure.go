package dto

import "github.com/brienze1/xenon/internal/domain/entity"

type ImportFailure struct {
	ItemKey string `json:"item_key"`
	Name    string `json:"name"`
	Code    string `json:"code"`
	Message string `json:"message"`
}

func NewImportFailure(failure entity.ImportFailure) ImportFailure {
	return ImportFailure{ItemKey: failure.ItemKey, Name: failure.Name, Code: failure.Code, Message: failure.Message}
}
