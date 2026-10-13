package dto

import "go.yaml.in/yaml/v3"

type Content []MediaType

func (c Content) MarshalYAML() (any, error) {
	return mappingOf(c, func(mediaType MediaType) string { return mediaType.Name })
}

func (c *Content) UnmarshalYAML(node *yaml.Node) error {
	entries, err := entriesOf(node, func(mediaType *MediaType, name string) { mediaType.Name = name })
	*c = entries
	return err
}
