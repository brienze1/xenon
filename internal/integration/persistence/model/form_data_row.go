package model

import (
	"go.yaml.in/yaml/v3"

	"github.com/brienze1/xenon/internal/domain/entity"
	"github.com/brienze1/xenon/internal/domain/enums/filestate"
	"github.com/brienze1/xenon/internal/domain/enums/formdatatype"
	"github.com/brienze1/xenon/pkg/yamlnode"
)

type FormDataRow struct {
	Enabled     *bool  `yaml:"enabled" json:"enabled,omitempty"`
	Key         string `yaml:"key" json:"key"`
	Type        string `yaml:"type" json:"type"`
	Value       string `yaml:"value" json:"value"`
	Src         string `yaml:"src" json:"src"`
	Description string `yaml:"description" json:"description"`
	SrcState    string `yaml:"srcState" json:"src_state,omitempty"`
}

func NewFormDataRow(row entity.FormDataRow) FormDataRow {
	return FormDataRow{
		Key:         row.Key,
		Type:        row.Type.String(),
		Value:       row.Value,
		Src:         row.Src.Src,
		SrcState:    fileRefState(row.Src),
		Enabled:     &row.Enabled,
		Description: row.Description,
	}
}

func (f *FormDataRow) ToEntity() entity.FormDataRow {
	row := entity.NewFormDataRow(f.Key, f.Value).WithType(f.Type)
	row.Src = fileRefToEntity(f.Src, f.SrcState)
	row.Description = f.Description
	if f.Enabled != nil {
		row.Enabled = *f.Enabled
	}
	return row
}

func (f *FormDataRow) Node() *yaml.Node {
	row := f.ToEntity()
	fields := []yamlnode.Field{yamlnode.String("key", f.Key)}
	if row.IsFile() {
		fields = append(fields, yamlnode.String("type", formdatatype.File.String()), yamlnode.String("src", f.Src))
		if row.Src.NeedsRepick() {
			fields = append(fields, yamlnode.String("srcState", filestate.Repick.String()))
		}
	} else {
		fields = append(fields, yamlnode.String("type", formdatatype.Text.String()), yamlnode.String("value", f.Value))
	}
	fields = append(fields, yamlnode.Bool("enabled", row.Enabled), yamlnode.String("description", f.Description))
	return yamlnode.MappingNode(fields...)
}
