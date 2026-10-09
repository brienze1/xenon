package model

import (
	"go.yaml.in/yaml/v3"

	"github.com/brienze1/xenon/internal/domain/entity"
	"github.com/brienze1/xenon/internal/domain/enums/formdatatype"
	"github.com/brienze1/xenon/internal/integration/persistence/yamlnode"
)

type FormDataRow struct {
	Enabled     *bool  `yaml:"enabled" json:"enabled,omitempty"`
	Key         string `yaml:"key" json:"key"`
	Type        string `yaml:"type" json:"type"`
	Value       string `yaml:"value" json:"value"`
	Src         string `yaml:"src" json:"src"`
	Description string `yaml:"description" json:"description"`
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

func (f *FormDataRow) Node() *yaml.Node {
	fields := []yamlnode.Field{yamlnode.String("key", f.Key)}
	if f.ToEntity().IsFile() {
		fields = append(fields, yamlnode.String("type", formdatatype.File.String()), yamlnode.String("src", f.Src))
	} else {
		fields = append(fields, yamlnode.String("type", formdatatype.Text.String()), yamlnode.String("value", f.Value))
	}
	fields = append(fields, yamlnode.Bool("enabled", isEnabled(f.Enabled)), yamlnode.String("description", f.Description))
	return yamlnode.MappingNode(fields...)
}
