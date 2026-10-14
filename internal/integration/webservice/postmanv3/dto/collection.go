package dto

import (
	"go.yaml.in/yaml/v3"

	"github.com/brienze1/xenon/internal/domain/entity"
)

const (
	collectionKind = "collection"
	v21Schema      = "https://schema.getpostman.com/json/collection/v2.1.0/collection.json"
)

type Collection struct {
	Extra       map[string]any `yaml:",inline"`
	Order       *float64       `yaml:"order"`
	ID          any            `yaml:"id"`
	Description any            `yaml:"description"`
	Kind        string         `yaml:"$kind"`
	Name        string         `yaml:"name"`
	Items       []yaml.Node    `yaml:"items"`
	Variables   []Variable     `yaml:"variables"`
	Scripts     []Script       `yaml:"scripts"`
	Auth        Auth           `yaml:"auth"`
	Children    []Child        `yaml:"-"`
}

func NewNamedCollection(name string) *Collection {
	return &Collection{Name: name, Children: []Child{}}
}

func (c *Collection) WithFallbackName(name string) *Collection {
	if c.Name == "" {
		c.Name = name
	}
	return c
}

func (c *Collection) ResolveItems() error {
	c.Children = make([]Child, 0, len(c.Items))
	for index := range c.Items {
		child, err := resolveItem(&c.Items[index])
		if err != nil {
			return err
		}
		c.Children = append(c.Children, child)
	}
	return nil
}

func resolveItem(node *yaml.Node) (Child, error) {
	var head struct {
		Kind string `yaml:"$kind"`
		Name string `yaml:"name"`
	}
	if err := node.Decode(&head); err != nil {
		return Child{}, err
	}
	switch {
	case head.Kind == collectionKind:
		var folder Collection
		if err := node.Decode(&folder); err != nil {
			return Child{}, err
		}
		return NewFolderChild(&folder, folder.Name), folder.ResolveItems()
	case IsSupportedRequestKind(head.Kind):
		var request Request
		if err := node.Decode(&request); err != nil {
			return Child{}, err
		}
		examples, err := request.inlineExamples()
		request.examples = examples
		return NewRequestChild(&request, request.Name), err
	case IsRequestKind(head.Kind):
		return NewSkippedChild(head.Name, head.Kind, head.Name, nil), nil
	default:
		return Child{}, UnknownKindError{Kind: head.Kind}
	}
}

func (c *Collection) ToPostmanV21(conversion *Conversion) (map[string]any, error) {
	extensions := map[string]any{}
	document := copied(c.Extra)
	info := map[string]any{"name": c.Name, "schema": v21Schema}
	setIfPresent(info, "_postman_id", c.ID)
	setIfPresent(info, "description", c.Description)
	document["info"] = info
	setIfPresent(document, "auth", c.Auth.toPostmanV21(conversion, c.Name, c.Name, extensions))
	setIfPresent(document, "event", scriptsToPostmanV21(c.Scripts, conversion, c.Name, c.Name, extensions))
	setIfPresent(document, "variable", c.variablesToPostmanV21())
	items, err := c.childrenToPostmanV21(conversion, c.Name)
	if err != nil {
		return nil, err
	}
	document["item"] = items
	setIfPresent(document, extensionsKey, extensions)
	return document, nil
}

func (c *Collection) folderToPostmanV21(conversion *Conversion, parentPath string) (map[string]any, error) {
	path := entity.JoinImportPath(parentPath, c.Name)
	extensions := map[string]any{}
	folder := copied(c.Extra)
	folder["name"] = c.Name
	setIfPresent(folder, "id", c.ID)
	setIfPresent(folder, "description", c.Description)
	setIfPresent(folder, "auth", c.Auth.toPostmanV21(conversion, c.Name, path, extensions))
	setIfPresent(folder, "event", scriptsToPostmanV21(c.Scripts, conversion, c.Name, path, extensions))
	setIfPresent(folder, "variable", c.variablesToPostmanV21())
	items, err := c.childrenToPostmanV21(conversion, path)
	if err != nil {
		return nil, err
	}
	folder["item"] = items
	setIfPresent(folder, extensionsKey, extensions)
	return folder, nil
}

func (c *Collection) childrenToPostmanV21(conversion *Conversion, path string) ([]any, error) {
	items := make([]any, 0, len(c.Children))
	for _, child := range c.Children {
		switch {
		case child.Folder != nil:
			folder, err := child.Folder.folderToPostmanV21(conversion, path)
			if err != nil {
				return nil, err
			}
			items = append(items, folder)
		case child.Request != nil:
			request, err := child.Request.toPostmanV21(conversion, path)
			if err != nil {
				return nil, err
			}
			items = append(items, request)
		default:
			conversion.skippedRequest(child.Name, child.SkippedKind, path)
		}
	}
	return items, nil
}

func (c *Collection) variablesToPostmanV21() []any {
	variables := make([]any, 0, len(c.Variables))
	for index := range c.Variables {
		variables = append(variables, c.Variables[index].toCollectionVariable())
	}
	return variables
}
