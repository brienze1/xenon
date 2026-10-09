package model

import (
	"go.yaml.in/yaml/v3"

	"github.com/brienze1/xenon/internal/domain/entity"
	"github.com/brienze1/xenon/pkg/yamlnode"
)

type KeyValue struct {
	Enabled     *bool  `yaml:"enabled" json:"enabled,omitempty"`
	Key         string `yaml:"key" json:"key"`
	Value       string `yaml:"value" json:"value"`
	Description string `yaml:"description" json:"description"`
}

func NewKeyValue(row entity.KeyValue) KeyValue {
	return KeyValue{Key: row.Key, Value: row.Value, Enabled: &row.Enabled, Description: row.Description}
}

func (k *KeyValue) ToEntity() entity.KeyValue {
	row := entity.NewKeyValue(k.Key, k.Value)
	row.Description = k.Description
	if k.Enabled != nil {
		row.Enabled = *k.Enabled
	}
	return row
}

func (k *KeyValue) Node() *yaml.Node {
	return yamlnode.MappingNode(
		yamlnode.String("key", k.Key),
		yamlnode.String("value", k.Value),
		yamlnode.Bool("enabled", k.ToEntity().Enabled),
		yamlnode.String("description", k.Description),
	)
}
