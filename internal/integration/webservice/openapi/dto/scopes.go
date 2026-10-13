package dto

import "go.yaml.in/yaml/v3"

type Scopes []string

func (s Scopes) MarshalYAML() (any, error) {
	mapping := emptyMapping()
	for _, scope := range s {
		mapping.Style = 0
		mapping.Content = append(mapping.Content, keyNode(scope), &yaml.Node{Kind: yaml.ScalarNode, Tag: stringTag, Value: ""})
	}
	return mapping, nil
}

func (s *Scopes) UnmarshalYAML(node *yaml.Node) error {
	node = resolved(node)
	scopes := make(Scopes, 0, len(node.Content)/2)
	for index := 0; index+1 < len(node.Content); index += 2 {
		scopes = append(scopes, node.Content[index].Value)
	}
	*s = scopes
	return nil
}
