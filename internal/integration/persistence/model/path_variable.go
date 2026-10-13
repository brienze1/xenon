package model

import (
	"go.yaml.in/yaml/v3"

	"github.com/brienze1/xenon/internal/domain/entity"
	"github.com/brienze1/xenon/pkg/yamlnode"
)

type PathVariable struct {
	Key         string `yaml:"key" json:"key"`
	Value       string `yaml:"value" json:"value"`
	Description string `yaml:"description" json:"description"`
}

func NewPathVariable(variable entity.PathVariable) PathVariable {
	return PathVariable{Key: variable.Key, Value: variable.Value, Description: variable.Description}
}

func (p *PathVariable) ToEntity() entity.PathVariable {
	return entity.PathVariable{Key: p.Key, Value: p.Value, Description: p.Description}
}

func pathVariableNodes(models []PathVariable) []*yaml.Node {
	nodes := make([]*yaml.Node, 0, len(models))
	for index := range models {
		nodes = append(nodes, models[index].Node())
	}
	return nodes
}

func (p *PathVariable) Node() *yaml.Node {
	return yamlnode.MappingNode(
		yamlnode.String("key", p.Key),
		yamlnode.String("value", p.Value),
		yamlnode.String("description", p.Description),
	)
}
