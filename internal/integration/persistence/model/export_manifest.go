package model

import (
	"go.yaml.in/yaml/v3"

	"github.com/brienze1/xenon/pkg/yamlnode"
)

const (
	exportManifestKind = "export"
	secretsInFile      = "file"
	SecretsKey         = "secrets"
)

type ExportManifest struct {
	Kind                   string `yaml:"kind"`
	Root                   string `yaml:"root"`
	ExportedAt             string `yaml:"exportedAt"`
	ExportedBy             string `yaml:"exportedBy"`
	Secrets                string `yaml:"secrets"`
	Xenon                  int    `yaml:"xenon"`
	CurrentValuesAsInitial bool   `yaml:"currentValuesAsInitial"`
}

func NewExportManifest(root, exportedAt, exportedBy string, currentValuesAsInitial, hasSecrets bool) ExportManifest {
	manifest := ExportManifest{Xenon: formatVersion, Kind: exportManifestKind, Root: root, ExportedAt: exportedAt, ExportedBy: exportedBy, CurrentValuesAsInitial: currentValuesAsInitial}
	if hasSecrets {
		manifest.Secrets = secretsInFile
	}
	return manifest
}

func (e *ExportManifest) HasSecretsFile() bool {
	return e.Secrets == secretsInFile
}

func (e *ExportManifest) Document() *yaml.Node {
	secrets := NullNode()
	if e.HasSecretsFile() {
		secrets = yamlnode.StringNode(e.Secrets)
	}
	return &yaml.Node{Kind: yaml.DocumentNode, Content: []*yaml.Node{yamlnode.MappingNode(
		yamlnode.Int("xenon", e.Xenon),
		yamlnode.String("kind", e.Kind),
		yamlnode.String("root", e.Root),
		yamlnode.String("exportedAt", e.ExportedAt),
		yamlnode.String("exportedBy", e.ExportedBy),
		yamlnode.Bool("currentValuesAsInitial", e.CurrentValuesAsInitial),
		yamlnode.Field{Key: SecretsKey, Value: secrets},
	)}}
}

func NullNode() *yaml.Node {
	return &yaml.Node{Kind: yaml.ScalarNode, Tag: "!!null", Value: "null"}
}
