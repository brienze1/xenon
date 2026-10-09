package dto

import (
	"github.com/brienze1/xenon/internal/domain/entity"
	"github.com/brienze1/xenon/internal/domain/enums/formdatatype"
)

type FormDataRow struct {
	Enabled     *bool  `json:"enabled,omitempty"`
	Key         string `json:"key"`
	Type        string `json:"type"`
	Value       string `json:"value"`
	Src         string `json:"src"`
	Description string `json:"description"`
}

func NewFormDataRow(row entity.FormDataRow) FormDataRow {
	return FormDataRow{
		Key:         row.Key,
		Type:        row.Type.String(),
		Value:       row.Value,
		Src:         row.Src.Src,
		Enabled:     &row.Enabled,
		Description: row.Description,
	}
}

func (f *FormDataRow) ToEntity() entity.FormDataRow {
	return entity.FormDataRow{
		Key:         f.Key,
		Type:        formdatatype.GetFormDataType(f.Type),
		Value:       f.Value,
		Src:         entity.FileRef{Src: f.Src},
		Enabled:     isEnabled(f.Enabled),
		Description: f.Description,
	}
}
