package dto

import (
	"github.com/brienze1/xenon/internal/domain/enums"
	"github.com/brienze1/xenon/internal/domain/enums/exportitemkind"
)

type ExportItemReference struct {
	LocalId  string `json:"local_id" validate:"required,not_blank"`
	ItemKind string `json:"item_kind" validate:"required,oneof=workspace collection folder request environment globals"`
	ItemId   string `json:"item_id" validate:"required_unless=ItemKind workspace ItemKind globals"`
}

func (e *ExportItemReference) Kind() enums.ExportItemKind {
	return exportitemkind.GetExportItemKind(e.ItemKind)
}
