package dto

import (
	"go.yaml.in/yaml/v3"

	"github.com/brienze1/xenon/internal/domain/entity"
)

type Paths []PathItem

// NewPaths groups the operations by path and method; a later operation on the same method and path is a variant.
func NewPaths(operations []entity.ApiOperation) Paths {
	paths := Paths{}
	for _, operation := range operations {
		index := len(paths) - 1
		if index < 0 || paths[index].Path != operation.Path {
			paths = append(paths, PathItem{Path: operation.Path})
			index++
		}
		item := &paths[index]
		last := len(item.Operations) - 1
		if operation.Variant && last >= 0 {
			item.Operations[last].Operation.Variants = append(item.Operations[last].Operation.Variants, NewOperation(operation))
			continue
		}
		item.Operations = append(item.Operations, PathOperation{Method: operation.Method, Operation: NewOperation(operation)})
	}
	return paths
}

func (p Paths) MarshalYAML() (any, error) {
	return mappingOf(p, func(item PathItem) string { return item.Path })
}

func (p *Paths) UnmarshalYAML(node *yaml.Node) error {
	entries, err := entriesOf(node, func(item *PathItem, path string) { item.Path = path })
	*p = entries
	return err
}

func (p Paths) ToEntity() ([]entity.ApiOperation, error) {
	operations := make([]entity.ApiOperation, 0)
	for _, item := range p {
		for index := range item.Operations {
			read, err := item.Operations[index].Operation.ToEntity(item.Operations[index].Method, item.Path, item.Parameters)
			if err != nil {
				return nil, err
			}
			operations = append(operations, read...)
		}
	}
	return operations, nil
}
