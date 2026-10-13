package model

import (
	"go.yaml.in/yaml/v3"

	"github.com/brienze1/xenon/pkg/yamlnode"
)

const (
	formatVersion = 1
	scriptsKey    = "scripts"
	authKey       = "auth"
)

func inheritScriptsNode(inherit *bool) *yaml.Node {
	return yamlnode.Bool("inheritScripts", inheritsScripts(inherit)).Value
}

func inheritsScripts(inherit *bool) bool {
	return inherit == nil || *inherit
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
