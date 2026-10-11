package dto

import "github.com/brienze1/xenon/internal/domain/entity"

type ExportDrop struct {
	ItemPath string `json:"item_path"`
	Field    string `json:"field"`
	Reason   string `json:"reason"`
}

func NewExportDrop(drop entity.ExportDrop) ExportDrop {
	return ExportDrop{ItemPath: drop.ItemPath, Field: drop.Field, Reason: drop.Reason}
}
