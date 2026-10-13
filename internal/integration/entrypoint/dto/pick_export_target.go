package dto

import (
	"github.com/brienze1/xenon/internal/domain/enums"
	"github.com/brienze1/xenon/internal/domain/enums/exportformat"
	"github.com/brienze1/xenon/internal/domain/enums/exportshape"
)

type PickExportTarget struct {
	SuggestedFileName string `json:"suggested_file_name" validate:"required,not_blank"`
	Format            string `json:"format" validate:"required,oneof=xenon_yaml postman"`
	Shape             string `json:"shape" validate:"omitempty,oneof=zip bundle file"`
}

func (p *PickExportTarget) ExportFormat() enums.ExportFormat {
	return exportformat.GetExportFormat(p.Format)
}

func (p *PickExportTarget) ExportShape() enums.ExportShape {
	return exportshape.GetExportShape(p.Shape)
}
