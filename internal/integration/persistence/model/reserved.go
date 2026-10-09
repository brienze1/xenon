package model

import (
	"go.yaml.in/yaml/v3"

	"github.com/brienze1/xenon/internal/integration/persistence/yamlnode"
)

const formatVersion = 1

func emptyScriptsNode() *yaml.Node {
	return yamlnode.MappingNode(yamlnode.String("preRequest", ""), yamlnode.String("postResponse", ""))
}

func inheritScriptsNode() *yaml.Node {
	return yamlnode.Bool("inheritScripts", true).Value
}

func headerFields(kind, id, name, description string) []yamlnode.Field {
	return []yamlnode.Field{
		yamlnode.Int("xenon", formatVersion),
		yamlnode.String("kind", kind),
		yamlnode.String("id", id),
		yamlnode.String("name", name),
		yamlnode.String("description", description),
	}
}

func nonNil(values []string) []string {
	if values == nil {
		return []string{}
	}
	return values
}
