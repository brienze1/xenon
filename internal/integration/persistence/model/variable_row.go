package model

import (
	"go.yaml.in/yaml/v3"

	"github.com/brienze1/xenon/internal/domain/entity"
	"github.com/brienze1/xenon/pkg/yamlnode"
)

type VariableRow struct {
	Enabled     *bool  `yaml:"enabled"`
	Key         string `yaml:"key"`
	Value       string `yaml:"value"`
	Description string `yaml:"description"`
	Secret      bool   `yaml:"secret"`
}

func NewVariableRow(row entity.VariableRow) VariableRow {
	stored := row.WithoutSecretValue()
	return VariableRow{Key: stored.Key, Value: stored.InitialValue, Secret: stored.Secret, Enabled: &stored.Enabled, Description: stored.Description}
}

func (v *VariableRow) ToEntity() entity.VariableRow {
	row := entity.NewVariableRow(v.Key, v.Value)
	row.Secret = v.Secret
	row.Description = v.Description
	if v.Enabled != nil {
		row.Enabled = *v.Enabled
	}
	return row
}

func (v *VariableRow) Node() *yaml.Node {
	fields := []yamlnode.Field{yamlnode.String("key", v.Key)}
	if v.Secret {
		fields = append(fields, yamlnode.Bool("secret", true))
	} else {
		fields = append(fields, yamlnode.String("value", v.Value))
	}
	return yamlnode.MappingNode(append(fields,
		yamlnode.Bool("enabled", v.ToEntity().Enabled),
		yamlnode.String("description", v.Description),
	)...)
}
