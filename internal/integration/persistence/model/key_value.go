package model

import (
	"go.yaml.in/yaml/v3"

	"github.com/brienze1/xenon/internal/domain/entity"
	"github.com/brienze1/xenon/internal/integration/persistence/yamlnode"
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
	return entity.KeyValue{Key: k.Key, Value: k.Value, Enabled: isEnabled(k.Enabled), Description: k.Description}
}

func (k *KeyValue) Node() *yaml.Node {
	return yamlnode.MappingNode(
		yamlnode.String("key", k.Key),
		yamlnode.String("value", k.Value),
		yamlnode.Bool("enabled", isEnabled(k.Enabled)),
		yamlnode.String("description", k.Description),
	)
}

func isEnabled(enabled *bool) bool {
	return enabled == nil || *enabled
}
