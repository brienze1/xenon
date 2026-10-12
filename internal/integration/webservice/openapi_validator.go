package webservice

import (
	"slices"
	"strings"

	"go.yaml.in/yaml/v3"

	"github.com/brienze1/xenon/internal/integration/webservice/openapi/dto"
)

const (
	structureCode = "DOC-02400"
	duplicateCode = "DOC-03400"
	xenonIdKey    = "x-xenon-id"
	variantsKey   = "x-xenon-variants"
)

var parameterLocations = []string{"path", "query", "header", "cookie"}

type openApiValidator struct {
	ids map[string]bool
}

// validateOpenApi checks the shape Xenon reads, pointing at the line and column of the first problem.
func validateOpenApi(document *yaml.Node) error {
	if len(document.Content) == 0 {
		return documentProblem(structureCode, "the document is empty; it must be an OpenAPI 3 mapping", 1, 1)
	}
	root := document.Content[0]
	if root.Kind != yaml.MappingNode {
		return problemAt(root, "the document must be a mapping with openapi, info and paths")
	}
	validator := openApiValidator{ids: map[string]bool{}}
	return validator.root(root)
}

func (v openApiValidator) root(root *yaml.Node) error {
	version, found := field(root, "openapi")
	if !found {
		return problemAt(root, "the openapi field is missing; this must be an OpenAPI 3 document")
	}
	if !strings.HasPrefix(version.Value, "3.") {
		return problemAt(version, "openapi must be a 3.x version, not \""+version.Value+"\"")
	}
	info, found := field(root, "info")
	if !found {
		return problemAt(root, "the info field is missing")
	}
	if info.Kind != yaml.MappingNode {
		return problemAt(info, "info must be a mapping")
	}
	if title, found := field(info, "title"); !found || title.Kind != yaml.ScalarNode || strings.TrimSpace(title.Value) == "" {
		return problemAt(info, "info.title is required")
	}
	paths, found := field(root, "paths")
	if !found {
		return nil
	}
	if paths.Kind != yaml.MappingNode {
		return problemAt(paths, "paths must be a mapping")
	}
	return v.paths(paths)
}

func (v openApiValidator) paths(paths *yaml.Node) error {
	for index := 0; index+1 < len(paths.Content); index += 2 {
		key, item := paths.Content[index], paths.Content[index+1]
		if !strings.HasPrefix(key.Value, "/") {
			return problemAt(key, "the path \""+key.Value+"\" must start with /")
		}
		if item.Kind != yaml.MappingNode {
			return problemAt(item, "the path "+key.Value+" must be a mapping of methods")
		}
		if err := v.pathItem(item); err != nil {
			return err
		}
	}
	return nil
}

func (v openApiValidator) pathItem(item *yaml.Node) error {
	for index := 0; index+1 < len(item.Content); index += 2 {
		key, value := item.Content[index], item.Content[index+1]
		switch {
		case key.Value == "$ref":
			return problemAt(key, "references ($ref) are not supported")
		case key.Value == "parameters":
			if err := v.parameters(value); err != nil {
				return err
			}
		case dto.IsMethod(key.Value):
			if err := v.operation(key.Value, value); err != nil {
				return err
			}
		}
	}
	return nil
}

func (v openApiValidator) operation(method string, operation *yaml.Node) error {
	if operation.Kind != yaml.MappingNode {
		return problemAt(operation, "the "+method+" operation must be a mapping")
	}
	if id, found := field(operation, xenonIdKey); found {
		if id.Kind != yaml.ScalarNode {
			return problemAt(id, "x-xenon-id must be text")
		}
		if v.ids[id.Value] {
			return documentProblem(duplicateCode, "x-xenon-id "+id.Value+" is used by another operation", id.Line, id.Column)
		}
		v.ids[id.Value] = true
	}
	if tags, found := field(operation, "tags"); found && tags.Kind != yaml.SequenceNode {
		return problemAt(tags, "tags must be a list")
	}
	if parameters, found := field(operation, "parameters"); found {
		if err := v.parameters(parameters); err != nil {
			return err
		}
	}
	for _, name := range []string{"requestBody", "responses"} {
		if value, found := field(operation, name); found && value.Kind != yaml.MappingNode {
			return problemAt(value, name+" must be a mapping")
		}
	}
	variants, found := field(operation, variantsKey)
	if !found {
		return nil
	}
	if variants.Kind != yaml.SequenceNode {
		return problemAt(variants, "x-xenon-variants must be a list of operations")
	}
	for _, variant := range variants.Content {
		if err := v.operation(method, variant); err != nil {
			return err
		}
	}
	return nil
}

func (v openApiValidator) parameters(parameters *yaml.Node) error {
	if parameters.Kind != yaml.SequenceNode {
		return problemAt(parameters, "parameters must be a list")
	}
	for _, parameter := range parameters.Content {
		if parameter.Kind != yaml.MappingNode {
			return problemAt(parameter, "a parameter must be a mapping with name and in")
		}
		if ref, found := field(parameter, "$ref"); found {
			return problemAt(ref, "references ($ref) are not supported")
		}
		if name, found := field(parameter, "name"); !found || strings.TrimSpace(name.Value) == "" {
			return problemAt(parameter, "a parameter needs a name")
		}
		in, found := field(parameter, "in")
		if !found || !slices.Contains(parameterLocations, in.Value) {
			return problemAt(parameter, "a parameter's in must be path, query, header or cookie")
		}
	}
	return nil
}

func field(mapping *yaml.Node, name string) (*yaml.Node, bool) {
	for index := 0; index+1 < len(mapping.Content); index += 2 {
		if mapping.Content[index].Value == name {
			return mapping.Content[index+1], true
		}
	}
	return nil, false
}

func problemAt(node *yaml.Node, message string) error {
	return documentProblem(structureCode, message, node.Line, node.Column)
}
