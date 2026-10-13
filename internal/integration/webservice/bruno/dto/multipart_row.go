package dto

import (
	"github.com/brienze1/xenon/internal/domain/entity"
	"github.com/brienze1/xenon/internal/domain/enums/formdatatype"
)

type MultipartRow struct {
	Key         string
	Value       string
	ContentType string
	Files       []string
	Enabled     bool
}

func NewMultipartRow(key, value, contentType string, enabled bool) MultipartRow {
	return MultipartRow{Key: key, Value: value, ContentType: contentType, Enabled: enabled, Files: []string{}}
}

func (m MultipartRow) toEntity() entity.FormDataRow {
	row := entity.NewFormDataRow(m.Key, m.Value)
	row.Enabled = m.Enabled
	if len(m.Files) > 0 {
		row = entity.NewFormDataRow(m.Key, "").WithType(formdatatype.File.String())
		row.Enabled = m.Enabled
		row.Src = entity.NewRepickFileRef(m.Files[0])
	}
	return row
}
