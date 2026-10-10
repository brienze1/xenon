package entity

import (
	"strings"

	"github.com/brienze1/xenon/internal/domain/enums"
	"github.com/brienze1/xenon/internal/domain/enums/formdatatype"
)

type FormDataRow struct {
	Src         FileRef
	Type        enums.FormDataType
	Key         string
	Value       string
	Description string
	Enabled     bool
}

func NewFormDataRow(key, value string) FormDataRow {
	return FormDataRow{Key: key, Value: value, Type: formdatatype.Text, Enabled: true}
}

func (f FormDataRow) WithType(raw string) FormDataRow {
	if strings.TrimSpace(raw) != "" {
		f.Type = formdatatype.GetFormDataType(raw)
	}
	return f
}

func (f FormDataRow) MapText(transform func(string) string) FormDataRow {
	f.Key = transform(f.Key)
	f.Value = transform(f.Value)
	f.Src = FileRef{Src: transform(f.Src.Src)}
	return f
}

func (f FormDataRow) IsFile() bool {
	return f.Type == formdatatype.File
}

func (f FormDataRow) IsActive() bool {
	return f.Enabled && strings.TrimSpace(f.Key) != ""
}

func (f FormDataRow) LogValue() any {
	return map[string]any{"key": f.Key, "type": f.Type, "enabled": f.Enabled}
}
