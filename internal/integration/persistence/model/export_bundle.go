package model

import (
	"time"

	"go.yaml.in/yaml/v3"

	"github.com/brienze1/xenon/pkg/yamlnode"
)

const (
	bundleKind  = "bundle"
	ChildrenKey = "children"
	itemKey     = "item"
)

type ExportBundle struct {
	Kind       string    `yaml:"kind"`
	Root       string    `yaml:"root"`
	ExportedAt string    `yaml:"exportedAt"`
	ExportedBy string    `yaml:"exportedBy"`
	Secrets    string    `yaml:"secrets"`
	Item       yaml.Node `yaml:"item"`
	Xenon      int       `yaml:"xenon"`
}

func NewExportBundle(root string, exportedAt time.Time, exportedBy string, item *yaml.Node, secrets string) ExportBundle {
	return ExportBundle{Xenon: formatVersion, Kind: bundleKind, Root: root, ExportedAt: exportedAt.UTC().Format(exportedAtLayout), ExportedBy: exportedBy, Item: *item, Secrets: secrets}
}

func (e *ExportBundle) Document() *yaml.Node {
	secrets := NullNode()
	if e.Secrets != "" {
		secrets = yamlnode.StringNode(e.Secrets)
	}
	return &yaml.Node{Kind: yaml.DocumentNode, Content: []*yaml.Node{yamlnode.MappingNode(
		yamlnode.Int("xenon", e.Xenon),
		yamlnode.String("kind", e.Kind),
		yamlnode.String("root", e.Root),
		yamlnode.String("exportedAt", e.ExportedAt),
		yamlnode.String("exportedBy", e.ExportedBy),
		yamlnode.Field{Key: itemKey, Value: &e.Item},
		yamlnode.Field{Key: SecretsKey, Value: secrets},
	)}}
}

func WithChildren(item *yaml.Node, children []*yaml.Node) *yaml.Node {
	sequence := &yaml.Node{Kind: yaml.SequenceNode, Tag: "!!seq", Content: children}
	if len(children) == 0 {
		sequence.Style = yaml.FlowStyle
	}
	item.Content = append(item.Content, yamlnode.StringNode(ChildrenKey), sequence)
	return item
}
