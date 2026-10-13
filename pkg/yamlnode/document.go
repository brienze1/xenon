package yamlnode

import (
	"bytes"
	"errors"

	"go.yaml.in/yaml/v3"
)

var utf8Bom = []byte{0xEF, 0xBB, 0xBF}

var errNotAMapping = errors.New("the YAML document is not a mapping")

func Parse(data []byte) (*yaml.Node, error) {
	var document yaml.Node
	if err := yaml.Unmarshal(bytes.TrimPrefix(data, utf8Bom), &document); err != nil {
		return nil, err
	}
	if document.Kind == 0 {
		return NewDocument(), nil
	}
	if len(document.Content) != 1 || document.Content[0].Kind != yaml.MappingNode {
		return nil, errNotAMapping
	}
	return &document, nil
}

func NewDocument() *yaml.Node {
	return &yaml.Node{Kind: yaml.DocumentNode, Content: []*yaml.Node{{Kind: yaml.MappingNode, Tag: "!!map"}}}
}

func Root(document *yaml.Node) *yaml.Node {
	return document.Content[0]
}

func Encode(document *yaml.Node) ([]byte, error) {
	var buffer bytes.Buffer
	encoder := yaml.NewEncoder(&buffer)
	encoder.SetIndent(2)
	if err := encoder.Encode(document); err != nil {
		return nil, err
	}
	if err := encoder.Close(); err != nil {
		return nil, err
	}
	return buffer.Bytes(), nil
}
