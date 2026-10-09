package dto

import "github.com/brienze1/xenon/internal/domain/entity"

type FormDataRow struct {
	Enabled     *bool  `json:"enabled,omitempty"`
	Key         string `json:"key"`
	Type        string `json:"type" validate:"omitempty,oneof=text file"`
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
	row := entity.NewFormDataRow(f.Key, f.Value).WithType(f.Type)
	row.Src = entity.FileRef{Src: f.Src}
	row.Description = f.Description
	if f.Enabled != nil {
		row.Enabled = *f.Enabled
	}
	return row
}
